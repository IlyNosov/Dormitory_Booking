package bot

import (
	appbooking "Dormitory_Booking/internal/application/booking"
	roomsvc "Dormitory_Booking/internal/application/room"
	domain "Dormitory_Booking/internal/domain/booking"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// mskLoc is Moscow Standard Time (UTC+3), no tzdata needed.
var mskLoc = time.FixedZone("MSK", 3*60*60)

// parseFlexDate parses a date string in several formats:
//   - "02.01.2006" — full
//   - "02.01.06"   — 2-digit year (e.g. 12.06.25)
//   - "02.01"      — no year, assumes current year
func parseFlexDate(text string, loc *time.Location) (time.Time, error) {
	formats := []string{"02.01.2006", "02.01.06", "02.01"}
	for _, f := range formats {
		d, err := time.ParseInLocation(f, text, loc)
		if err != nil {
			continue
		}
		if f == "02.01" {
			d = time.Date(time.Now().In(loc).Year(), d.Month(), d.Day(), 0, 0, 0, 0, loc)
		}
		return d, nil
	}
	return time.Time{}, fmt.Errorf("unsupported date format: %q", text)
}

// thin wrappers so tests can override them
var newHTTPRequest = func(method, url string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, url, body)
}
var httpDo = func(req *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}

// ── Interfaces ────────────────────────────────────────────────────────────────

type AuthService interface {
	RequestOTP(ctx context.Context, email string) error
	VerifyOTPOnly(ctx context.Context, email, code string) error
}

type UserStore interface {
	Save(ctx context.Context, chatID int64, email string) error
	LoadAll(ctx context.Context) (map[int64]string, error)
}

// ── Session state ─────────────────────────────────────────────────────────────

type step int

const (
	stepAskEmail step = iota
	stepAskOTP
	stepMain
	stepListDate
	stepBookRoom
	stepBookDate
	stepBookDateCustom
	stepBookTime
	stepBookDuration
	stepBookDurationCustom
	stepBookTitle
	stepBookTelegram
	stepBookTelegramCustom
	stepBookType
	stepBookConfirm
	stepMyBookings
	stepCancelConfirm
	stepEditMenu
	stepEditTitle
	stepDateBookings
)

type session struct {
	Step         step
	PendingEmail string
	MsgID        int // current "screen" message ID for edit-in-place

	// booking draft (shared between create & edit flows)
	Room        int
	DateStr     string // "DD.MM.YYYY" Moscow
	StartStr    string // "HH:MM"
	DurationMin int
	Title       string
	TelegramStr string
	IsPrivate   bool

	// edit mode: non-empty means we are editing an existing booking
	EditID string

	// my bookings pagination
	MyBookings []domain.Booking
	BookingIdx int

	// date bookings pagination
	DateBookings []domain.Booking
	DateBookingIdx int
	DateLabel string
}

// ── Bot ───────────────────────────────────────────────────────────────────────

type BookingBot struct {
	bot        *tgbotapi.BotAPI
	svc        *appbooking.Service
	rooms      *roomsvc.Service
	auth       AuthService
	store      UserStore
	mu         sync.Mutex
	sessions   map[int64]*session
	emails     map[int64]string // chatID → verified email
	usernames  map[int64]string // chatID → @username from TG profile
	backendURL string
	botSecret  string
}

func NewBotAPI(token string) (*tgbotapi.BotAPI, error) {
	b, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	_, _ = b.Request(tgbotapi.DeleteWebhookConfig{})
	return b, nil
}

func StartBookingBot(ctx context.Context, api *tgbotapi.BotAPI, svc *appbooking.Service, rooms *roomsvc.Service, auth AuthService, store UserStore) {
	backendURL := os.Getenv("BACKEND_INTERNAL_URL")
	if backendURL == "" {
		backendURL = "http://localhost:8080"
	}

	emails := make(map[int64]string)
	if store != nil {
		if all, err := store.LoadAll(ctx); err == nil {
			emails = all
		}
	}
	log.Printf("telegram bot @%s started (backend: %s, %d authenticated users)", api.Self.UserName, backendURL, len(emails))

	b := &BookingBot{
		bot:        api,
		svc:        svc,
		rooms:      rooms,
		auth:       auth,
		store:      store,
		sessions:   make(map[int64]*session),
		emails:     emails,
		usernames:  make(map[int64]string),
		backendURL: backendURL,
		botSecret:  os.Getenv("BOT_INTERNAL_SECRET"),
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := api.GetUpdatesChan(u)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case upd := <-updates:
				if upd.Message != nil && upd.Message.Chat != nil && upd.Message.Chat.IsPrivate() {
					if upd.Message.From != nil && upd.Message.From.UserName != "" {
						b.mu.Lock()
						b.usernames[upd.Message.Chat.ID] = upd.Message.From.UserName
						b.mu.Unlock()
					}
					b.handleMessage(upd.Message)
				} else if upd.CallbackQuery != nil {
					b.handleCallback(upd.CallbackQuery)
				}
			}
		}
	}()
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// eh escapes HTML special characters for safe use in HTML parse mode.
func eh(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// editKB edits an existing message in place. Falls back to Send if msgID==0 or edit fails.
// Returns the message ID of the resulting message.
func (b *BookingBot) editKB(chatID int64, msgID int, text string, k tgbotapi.InlineKeyboardMarkup) int {
	if msgID != 0 {
		cfg := tgbotapi.NewEditMessageText(chatID, msgID, text)
		cfg.ParseMode = tgbotapi.ModeHTML
		cfg.ReplyMarkup = &k
		if msg, err := b.bot.Send(cfg); err == nil {
			return msg.MessageID
		} else if strings.Contains(err.Error(), "message is not modified") {
			return msgID
		}
	}
	m := tgbotapi.NewMessage(chatID, text)
	m.ParseMode = tgbotapi.ModeHTML
	m.ReplyMarkup = k
	result, _ := b.bot.Send(m)
	return result.MessageID
}

// sendKB sends a new message with inline keyboard, returns its MessageID.
func (b *BookingBot) sendKB(chatID int64, text string, k tgbotapi.InlineKeyboardMarkup) int {
	m := tgbotapi.NewMessage(chatID, text)
	m.ParseMode = tgbotapi.ModeHTML
	m.ReplyMarkup = k
	result, _ := b.bot.Send(m)
	return result.MessageID
}

// send sends a plain text message (for auth flow where edit-in-place is not possible).
func (b *BookingBot) send(chatID int64, text string) {
	m := tgbotapi.NewMessage(chatID, text)
	m.ParseMode = tgbotapi.ModeHTML
	_, _ = b.bot.Send(m)
}

func (b *BookingBot) ack(cq *tgbotapi.CallbackQuery) {
	_, _ = b.bot.Request(tgbotapi.NewCallback(cq.ID, ""))
}

func btn(text, data string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData(text, data)
}

func kb(rows ...[]tgbotapi.InlineKeyboardButton) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func cancelKB() tgbotapi.InlineKeyboardMarkup {
	return kb([]tgbotapi.InlineKeyboardButton{btn("❌ Отмена", "cancel")})
}

func menuKB() tgbotapi.InlineKeyboardMarkup {
	return kb([]tgbotapi.InlineKeyboardButton{btn("🏠 В меню", "menu:main")})
}

func (b *BookingBot) getEmail(chatID int64) (string, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	e, ok := b.emails[chatID]
	return e, ok
}

func (b *BookingBot) setEmail(chatID int64, email string) {
	b.mu.Lock()
	b.emails[chatID] = email
	b.mu.Unlock()
	if b.store != nil {
		_ = b.store.Save(context.Background(), chatID, email)
	}
}

func (b *BookingBot) sess(chatID int64) *session {
	b.mu.Lock()
	defer b.mu.Unlock()
	if s, ok := b.sessions[chatID]; ok {
		return s
	}
	s := &session{}
	b.sessions[chatID] = s
	return s
}

func (b *BookingBot) reset(chatID int64) {
	b.mu.Lock()
	b.sessions[chatID] = &session{}
	b.mu.Unlock()
}

func (b *BookingBot) knownUsername(chatID int64) string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.usernames[chatID]
}

func displayRoom(r int) string {
	switch r {
	case 2812:
		return "812 (2к)"
	case 3812:
		return "812 (3к)"
	}
	return strconv.Itoa(r)
}

// ── Auth flow ─────────────────────────────────────────────────────────────────

func (b *BookingBot) showWelcome(chatID int64, firstName string) {
	greeting := "👋 Привет"
	if firstName != "" {
		greeting = "👋 Привет, " + eh(firstName) + "!"
	} else {
		greeting += "!"
	}
	text := greeting + `

🏢 <b>Бот бронирования досуговых комнат</b>
Общежитие Дубки ВШЭ

Здесь вы можете:
📅 Бронировать комнаты (21, 132, 256, 812)
👤 Управлять своими бронями
📋 Смотреть расписание на любой день

Для входа нужна почта ВШЭ
<i>(@edu.hse.ru или @hse.ru)</i>`
	b.sendKB(chatID, text, kb(
		[]tgbotapi.InlineKeyboardButton{btn("🔑 Войти", "auth:start")},
	))
}

func (b *BookingBot) askEmail(chatID int64) {
	b.sess(chatID).Step = stepAskEmail
	b.send(chatID, "👋 Для входа введите почту ВШЭ\n<i>(@edu.hse.ru или @hse.ru)</i>")
}

func (b *BookingBot) handleAuthEmail(m *tgbotapi.Message, s *session) {
	email := strings.TrimSpace(strings.ToLower(m.Text))
	if !strings.HasSuffix(email, "@edu.hse.ru") && !strings.HasSuffix(email, "@hse.ru") {
		b.send(m.Chat.ID, "❌ Принимаются только @edu.hse.ru или @hse.ru.\nПопробуйте ещё раз:")
		return
	}
	if err := b.auth.RequestOTP(context.Background(), email); err != nil {
		b.send(m.Chat.ID, "❌ Не удалось отправить код: "+eh(err.Error()))
		return
	}
	s.PendingEmail = email
	s.Step = stepAskOTP
	b.send(m.Chat.ID, fmt.Sprintf("📧 Код отправлен на <code>%s</code>\n\nВведите 6-значный код из письма:", eh(email)))
}

func (b *BookingBot) handleAuthOTP(m *tgbotapi.Message, s *session) {
	if err := b.auth.VerifyOTPOnly(context.Background(), s.PendingEmail, strings.TrimSpace(m.Text)); err != nil {
		b.send(m.Chat.ID, "❌ Неверный или устаревший код.\nВведите почту заново:")
		s.Step = stepAskEmail
		return
	}
	b.setEmail(m.Chat.ID, s.PendingEmail)
	s.Step = stepMain
	b.send(m.Chat.ID, fmt.Sprintf("✅ Вы вошли как <code>%s</code>", eh(s.PendingEmail)))
	b.showMenu(m.Chat.ID, 0)
}

// ── Main menu ─────────────────────────────────────────────────────────────────

func (b *BookingBot) showMenu(chatID int64, msgID int) {
	b.sess(chatID).Step = stepMain
	text := "🏢 <b>Дубки — бронирование комнат</b>\n\n" +
		"Выберите действие:"
	b.editKB(chatID, msgID, text, kb(
		[]tgbotapi.InlineKeyboardButton{btn("📅 Забронировать", "menu:book")},
		[]tgbotapi.InlineKeyboardButton{btn("👤 Мои брони", "menu:mine")},
		[]tgbotapi.InlineKeyboardButton{btn("📋 Сегодня", "menu:today"), btn("📋 Завтра", "menu:tomorrow")},
		[]tgbotapi.InlineKeyboardButton{btn("📋 На другую дату", "menu:listdate")},
	))
}

// ── List all bookings ─────────────────────────────────────────────────────────

func (b *BookingBot) listDate(chatID int64, msgID int, date time.Time) {
	all, err := b.svc.ListBookings(context.Background())
	if err != nil {
		b.editKB(chatID, msgID, "❌ Не удалось загрузить брони.", menuKB())
		return
	}
	dateMSK := date.In(mskLoc)
	y, mo, d := dateMSK.Date()
	var found []domain.Booking
	for _, bk := range all {
		by, bmo, bd := bk.Start.In(mskLoc).Date()
		if by == y && bmo == mo && bd == d {
			found = append(found, bk)
		}
	}
	label := dateMSK.Format("02.01.2006")
	if len(found) == 0 {
		b.editKB(chatID, msgID, fmt.Sprintf("📋 <b>%s</b>\n\nБронирований нет.", label), menuKB())
		return
	}
	sort.Slice(found, func(i, j int) bool { return found[i].Start.Before(found[j].Start) })
	s := b.sess(chatID)
	s.DateBookings = found
	s.DateBookingIdx = 0
	s.DateLabel = label
	s.Step = stepDateBookings
	b.showDateBookingPage(chatID, msgID, 0)
}

func (b *BookingBot) showDateBookingPage(chatID int64, msgID int, idx int) {
	s := b.sess(chatID)
	if len(s.DateBookings) == 0 {
		b.showMenu(chatID, msgID)
		return
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= len(s.DateBookings) {
		idx = len(s.DateBookings) - 1
	}
	s.DateBookingIdx = idx

	bk := s.DateBookings[idx]
	n := len(s.DateBookings)

	typeIcon := "🔓 Публичное"
	if bk.IsPrivate {
		typeIcon = "🔒 Частное"
	}
	contact := ""
	if bk.TelegramID != "" {
		contact = fmt.Sprintf("\n📱 @%s", eh(bk.TelegramID))
	}
	text := fmt.Sprintf(
		"📋 <b>Брони на %s</b>  (%d / %d)\n\n🕐 %s – %s\n🏠 Комн. %s\n📝 %s\n%s%s",
		s.DateLabel, idx+1, n,
		bk.Start.In(mskLoc).Format("15:04"),
		bk.End.In(mskLoc).Format("15:04"),
		eh(displayRoom(bk.Room)),
		eh(bk.Title),
		typeIcon,
		contact,
	)

	var rows [][]tgbotapi.InlineKeyboardButton
	if n > 1 {
		var navRow []tgbotapi.InlineKeyboardButton
		if idx > 0 {
			navRow = append(navRow, btn("⬅️", "datebk_prev"))
		}
		if idx < n-1 {
			navRow = append(navRow, btn("➡️", "datebk_next"))
		}
		rows = append(rows, navRow)
	}
	rows = append(rows, []tgbotapi.InlineKeyboardButton{btn("🏠 В меню", "menu:main")})

	newMsgID := b.editKB(chatID, msgID, text, tgbotapi.NewInlineKeyboardMarkup(rows...))
	s.MsgID = newMsgID
}

// ── My bookings (paginated single message) ────────────────────────────────────

func (b *BookingBot) loadMyBookings(chatID int64, msgID int, startIdx int) {
	email, _ := b.getEmail(chatID)
	all, err := b.svc.ListBookings(context.Background())
	if err != nil {
		b.editKB(chatID, msgID, "❌ Не удалось загрузить брони.", menuKB())
		return
	}
	now := time.Now().In(mskLoc)
	var mine []domain.Booking
	for _, bk := range all {
		if bk.UserEmail == email && bk.End.In(mskLoc).After(now) {
			mine = append(mine, bk)
		}
	}
	if len(mine) == 0 {
		b.editKB(chatID, msgID, "👤 <b>Мои брони</b>\n\nУ вас нет предстоящих бронирований.", menuKB())
		return
	}
	sort.Slice(mine, func(i, j int) bool { return mine[i].Start.Before(mine[j].Start) })
	s := b.sess(chatID)
	s.MyBookings = mine
	s.Step = stepMyBookings
	idx := startIdx
	if idx >= len(mine) {
		idx = len(mine) - 1
	}
	if idx < 0 {
		idx = 0
	}
	b.showBookingPage(chatID, msgID, idx)
}

func (b *BookingBot) showBookingPage(chatID int64, msgID int, idx int) {
	s := b.sess(chatID)
	if len(s.MyBookings) == 0 {
		b.showMenu(chatID, msgID)
		return
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= len(s.MyBookings) {
		idx = len(s.MyBookings) - 1
	}
	s.BookingIdx = idx

	bk := s.MyBookings[idx]
	n := len(s.MyBookings)

	typeIcon := "🔓 Публичное"
	if bk.IsPrivate {
		typeIcon = "🔒 Частное"
	}
	text := fmt.Sprintf(
		"👤 <b>Ваши брони</b>  (%d / %d)\n\n📅 %s  —  Комн. %s\n🕐 %s – %s\n📝 %s\n%s",
		idx+1, n,
		bk.Start.In(mskLoc).Format("02.01.2006"),
		eh(displayRoom(bk.Room)),
		bk.Start.In(mskLoc).Format("15:04"),
		bk.End.In(mskLoc).Format("15:04"),
		eh(bk.Title),
		typeIcon,
	)

	var rows [][]tgbotapi.InlineKeyboardButton
	// Row 1: edit + delete
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		btn("✏️ Редактировать", "edit_bk:"+bk.ID),
		btn("🗑 Удалить", "cancel_ask:"+bk.ID),
	})
	// Row 2: navigation (only if more than one booking)
	if n > 1 {
		var navRow []tgbotapi.InlineKeyboardButton
		if idx > 0 {
			navRow = append(navRow, btn("⬅️", "mybk_prev"))
		}
		if idx < n-1 {
			navRow = append(navRow, btn("➡️", "mybk_next"))
		}
		rows = append(rows, navRow)
	}
	// Row 3: back to menu
	rows = append(rows, []tgbotapi.InlineKeyboardButton{btn("🏠 В меню", "menu:main")})

	newMsgID := b.editKB(chatID, msgID, text, tgbotapi.NewInlineKeyboardMarkup(rows...))
	s.MsgID = newMsgID
}

// ── Edit booking ──────────────────────────────────────────────────────────────

func (b *BookingBot) showEditMenu(chatID int64, msgID int, bk domain.Booking) {
	s := b.sess(chatID)
	s.Step = stepEditMenu
	s.EditID = bk.ID
	s.Room = bk.Room
	s.DateStr = bk.Start.In(mskLoc).Format("02.01.2006")
	s.StartStr = bk.Start.In(mskLoc).Format("15:04")
	s.DurationMin = int(bk.End.Sub(bk.Start).Minutes())
	s.Title = bk.Title
	s.TelegramStr = bk.TelegramID
	s.IsPrivate = bk.IsPrivate

	typeIcon := "🔓 Публичное"
	if bk.IsPrivate {
		typeIcon = "🔒 Частное"
	}
	text := fmt.Sprintf(
		"✏️ <b>Редактирование брони</b>\n\n📅 %s  —  Комн. %s\n🕐 %s – %s\n📝 %s\n%s\n\nЧто изменить?",
		bk.Start.In(mskLoc).Format("02.01.2006"),
		eh(displayRoom(bk.Room)),
		bk.Start.In(mskLoc).Format("15:04"),
		bk.End.In(mskLoc).Format("15:04"),
		eh(bk.Title),
		typeIcon,
	)
	newMsgID := b.editKB(chatID, msgID, text, kb(
		[]tgbotapi.InlineKeyboardButton{btn("📝 Название", "edit_field:title"), btn("🕐 Дата и время", "edit_field:time")},
		[]tgbotapi.InlineKeyboardButton{btn("🔒 Тип", "edit_field:type"), btn("📱 Telegram", "edit_field:tg")},
		[]tgbotapi.InlineKeyboardButton{btn("◀️ Назад", "mybk_cur")},
	))
	s.MsgID = newMsgID
}

func (b *BookingBot) doUpdate(chatID int64, msgID int) {
	s := b.sess(chatID)
	email, _ := b.getEmail(chatID)

	startParsed, err := time.ParseInLocation("02.01.2006 15:04",
		fmt.Sprintf("%s %s", s.DateStr, s.StartStr), mskLoc)
	if err != nil {
		b.editKB(chatID, msgID, "❌ Ошибка разбора времени.", kb(
			[]tgbotapi.InlineKeyboardButton{btn("◀️ Назад", "mybk_cur")},
		))
		return
	}
	endTime := startParsed.Add(time.Duration(s.DurationMin) * time.Minute)
	tgID := strings.TrimPrefix(s.TelegramStr, "@")

	_, err = b.svc.UpdateBooking(context.Background(), appbooking.UpdateBookingInput{
		ID:         s.EditID,
		Start:      startParsed,
		End:        endTime,
		Room:       s.Room,
		Title:      s.Title,
		TelegramID: tgID,
		IsPrivate:  s.IsPrivate,
	}, email, false)
	if err != nil {
		b.editKB(chatID, msgID, "❌ Не удалось обновить: "+eh(humanizeError(err)), kb(
			[]tgbotapi.InlineKeyboardButton{btn("◀️ Назад", "mybk_cur")},
		))
		return
	}
	prevIdx := s.BookingIdx
	s.EditID = ""
	newMsgID := b.editKB(chatID, msgID, "✅ <b>Бронь обновлена!</b>", kb(
		[]tgbotapi.InlineKeyboardButton{btn("👤 Мои брони", "menu:mine_at:"+strconv.Itoa(prevIdx))},
		[]tgbotapi.InlineKeyboardButton{btn("🏠 В меню", "menu:main")},
	))
	s.MsgID = newMsgID
}

// ── Booking creation flow ─────────────────────────────────────────────────────

func (b *BookingBot) startBooking(chatID int64, msgID int) {
	s := b.sess(chatID)
	prevMsgID := s.MsgID
	if msgID != 0 {
		prevMsgID = msgID
	}
	*s = session{Step: stepBookRoom, MsgID: prevMsgID}

	activeRooms, _ := b.rooms.ActiveRooms(context.Background())
	var rows [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton
	for i, r := range activeRooms {
		label := "Комн. " + displayRoom(r.Number)
		row = append(row, btn(label, fmt.Sprintf("room:%d", r.Number)))
		if (i+1)%2 == 0 || i == len(activeRooms)-1 {
			rows = append(rows, row)
			row = nil
		}
	}
	rows = append(rows, []tgbotapi.InlineKeyboardButton{btn("❌ Отмена", "cancel")})
	newMsgID := b.editKB(chatID, prevMsgID, "📅 <b>Новое бронирование</b>\n\nВыберите комнату:", tgbotapi.NewInlineKeyboardMarkup(rows...))
	s.MsgID = newMsgID
}

func (b *BookingBot) askDate(chatID int64, msgID int) {
	s := b.sess(chatID)
	s.Step = stepBookDate
	todayStr := time.Now().In(mskLoc).Format("02.01")
	tomorrowStr := time.Now().In(mskLoc).AddDate(0, 0, 1).Format("02.01")
	newMsgID := b.editKB(chatID, msgID, "📅 Выберите дату:", kb(
		[]tgbotapi.InlineKeyboardButton{
			btn("Сегодня ("+todayStr+")", "date:today"),
			btn("Завтра ("+tomorrowStr+")", "date:tomorrow"),
		},
		[]tgbotapi.InlineKeyboardButton{btn("📅 Другая дата", "date:custom")},
		[]tgbotapi.InlineKeyboardButton{btn("❌ Отмена", "cancel")},
	))
	s.MsgID = newMsgID
}

func (b *BookingBot) askTime(chatID int64, msgID int) {
	s := b.sess(chatID)
	s.Step = stepBookTime
	newMsgID := b.editKB(chatID, msgID,
		"🕐 Введите время начала <i>(ЧЧ:ММ, например 18:00)</i>:", cancelKB())
	s.MsgID = newMsgID
}

func (b *BookingBot) askDuration(chatID int64, msgID int) {
	s := b.sess(chatID)
	s.Step = stepBookDuration
	newMsgID := b.editKB(chatID, msgID, "⏱️ Длительность:", kb(
		[]tgbotapi.InlineKeyboardButton{btn("30 мин", "dur:30"), btn("1 ч", "dur:60"), btn("1.5 ч", "dur:90")},
		[]tgbotapi.InlineKeyboardButton{btn("2 ч", "dur:120"), btn("2.5 ч", "dur:150"), btn("3 ч", "dur:180")},
		[]tgbotapi.InlineKeyboardButton{btn("✏️ Своя", "dur:custom")},
		[]tgbotapi.InlineKeyboardButton{btn("❌ Отмена", "cancel")},
	))
	s.MsgID = newMsgID
}

func (b *BookingBot) askTitle(chatID int64, msgID int) {
	s := b.sess(chatID)
	s.Step = stepBookTitle
	newMsgID := b.editKB(chatID, msgID, "📝 Введите название мероприятия:", cancelKB())
	s.MsgID = newMsgID
}

func (b *BookingBot) askTelegram(chatID int64, msgID int) {
	s := b.sess(chatID)
	s.Step = stepBookTelegram
	uname := b.knownUsername(chatID)
	var newMsgID int
	if uname != "" {
		newMsgID = b.editKB(chatID, msgID, "📱 Telegram для связи:", kb(
			[]tgbotapi.InlineKeyboardButton{btn("✅ @"+uname+" (мой)", "tg:auto")},
			[]tgbotapi.InlineKeyboardButton{btn("✏️ Другой", "tg:custom")},
			[]tgbotapi.InlineKeyboardButton{btn("❌ Отмена", "cancel")},
		))
	} else {
		s.Step = stepBookTelegramCustom
		newMsgID = b.editKB(chatID, msgID,
			"📱 Введите ваш Telegram <i>(например @username)</i>:", cancelKB())
	}
	s.MsgID = newMsgID
}

func (b *BookingBot) askType(chatID int64, msgID int) {
	s := b.sess(chatID)
	s.Step = stepBookType
	newMsgID := b.editKB(chatID, msgID, "🔒 Тип мероприятия:", kb(
		[]tgbotapi.InlineKeyboardButton{btn("🔓 Публичное", "type:public"), btn("🔒 Частное", "type:private")},
		[]tgbotapi.InlineKeyboardButton{btn("❌ Отмена", "cancel")},
	))
	s.MsgID = newMsgID
}

func (b *BookingBot) showConfirm(chatID int64, msgID int) {
	s := b.sess(chatID)
	s.Step = stepBookConfirm
	priv := "🔓 Публичное"
	if s.IsPrivate {
		priv = "🔒 Частное"
	}
	tgStr := s.TelegramStr
	if tgStr == "" {
		tgStr = "(не указан)"
	}
	endTime, _ := time.ParseInLocation("02.01.2006 15:04", s.DateStr+" "+s.StartStr, mskLoc)
	endTime = endTime.Add(time.Duration(s.DurationMin) * time.Minute)

	text := fmt.Sprintf(
		"✅ <b>Подтвердите бронирование</b>\n\n📅 %s  —  Комн. %s\n🕐 %s – %s  (%d мин)\n📝 %s\n📱 %s\n%s",
		eh(s.DateStr), eh(displayRoom(s.Room)),
		eh(s.StartStr), endTime.Format("15:04"), s.DurationMin,
		eh(s.Title), eh(tgStr), priv,
	)
	newMsgID := b.editKB(chatID, msgID, text, kb(
		[]tgbotapi.InlineKeyboardButton{btn("✅ Подтвердить", "confirm:yes"), btn("❌ Отмена", "cancel")},
	))
	s.MsgID = newMsgID
}

func (b *BookingBot) showConfirmEdit(chatID int64, msgID int) {
	s := b.sess(chatID)
	s.Step = stepBookConfirm
	endTime, _ := time.ParseInLocation("02.01.2006 15:04", s.DateStr+" "+s.StartStr, mskLoc)
	endTime = endTime.Add(time.Duration(s.DurationMin) * time.Minute)
	typeIcon := "🔓 Публичное"
	if s.IsPrivate {
		typeIcon = "🔒 Частное"
	}
	text := fmt.Sprintf(
		"✏️ <b>Подтвердите изменения</b>\n\n📅 %s  —  Комн. %s\n🕐 %s – %s  (%d мин)\n📝 %s\n%s",
		eh(s.DateStr), eh(displayRoom(s.Room)),
		eh(s.StartStr), endTime.Format("15:04"), s.DurationMin,
		eh(s.Title), typeIcon,
	)
	newMsgID := b.editKB(chatID, msgID, text, kb(
		[]tgbotapi.InlineKeyboardButton{btn("✅ Сохранить", "edit_confirm:yes"), btn("❌ Отмена", "cancel")},
	))
	s.MsgID = newMsgID
}

func (b *BookingBot) doCreate(chatID int64, msgID int) {
	s := b.sess(chatID)
	email, _ := b.getEmail(chatID)

	startParsed, err := time.ParseInLocation("02.01.2006 15:04",
		fmt.Sprintf("%s %s", s.DateStr, s.StartStr), mskLoc)
	if err != nil {
		b.editKB(chatID, msgID, "❌ Ошибка разбора времени. Начните заново.", menuKB())
		return
	}
	endTime := startParsed.Add(time.Duration(s.DurationMin) * time.Minute)
	tgID := strings.TrimPrefix(s.TelegramStr, "@")

	booking, err := b.svc.CreateBooking(context.Background(), appbooking.CreateBookingInput{
		Start:      startParsed,
		End:        endTime,
		Room:       s.Room,
		Title:      s.Title,
		IsPrivate:  s.IsPrivate,
		UserEmail:  email,
		TelegramID: tgID,
	})
	if err != nil {
		b.editKB(chatID, msgID, "❌ Не удалось забронировать: "+eh(humanizeError(err)), menuKB())
		return
	}
	b.reset(chatID)
	newMsgID := b.editKB(chatID, msgID, fmt.Sprintf(
		"✅ <b>Бронь создана!</b>\n\n📅 %s  —  Комн. %s\n🕐 %s – %s\n📝 %s",
		booking.Start.In(mskLoc).Format("02.01.2006"),
		eh(displayRoom(booking.Room)),
		booking.Start.In(mskLoc).Format("15:04"),
		booking.End.In(mskLoc).Format("15:04"),
		eh(booking.Title),
	), menuKB())
	b.sess(chatID).MsgID = newMsgID
}

// ── Message handler ───────────────────────────────────────────────────────────

func (b *BookingBot) handleMessage(m *tgbotapi.Message) {
	text := strings.TrimSpace(m.Text)
	chatID := m.Chat.ID

	// Delete user message immediately to keep chat clean
	_, _ = b.bot.Request(tgbotapi.NewDeleteMessage(chatID, m.MessageID))

	s := b.sess(chatID)

	// Global commands
	switch {
	case strings.HasPrefix(text, "/start"):
		parts := strings.Fields(text)
		if len(parts) == 2 {
			b.handleLinkToken(m, parts[1])
			return
		}
		if _, ok := b.getEmail(chatID); !ok {
			firstName := ""
			if m.From != nil {
				firstName = m.From.FirstName
			}
			b.showWelcome(chatID, firstName)
			return
		}
		b.showMenu(chatID, 0)
		return
	case text == "/menu" || text == "/help":
		if _, ok := b.getEmail(chatID); !ok {
			b.askEmail(chatID)
			return
		}
		b.showMenu(chatID, 0)
		return
	case strings.HasPrefix(text, "/link"):
		parts := strings.Fields(text)
		if len(parts) == 2 {
			b.handleLinkToken(m, parts[1])
		} else {
			b.send(chatID, "Используй: /link КОД")
		}
		return
	case strings.HasPrefix(text, "/cancel"):
		b.reset(chatID)
		if _, ok := b.getEmail(chatID); ok {
			b.showMenu(chatID, 0)
		} else {
			b.askEmail(chatID)
		}
		return
	}

	// Auth states
	if s.Step == stepAskEmail {
		b.handleAuthEmail(m, s)
		return
	}
	if s.Step == stepAskOTP {
		b.handleAuthOTP(m, s)
		return
	}

	// Need auth for everything else
	if _, ok := b.getEmail(chatID); !ok {
		b.askEmail(chatID)
		return
	}

	msgID := s.MsgID // edit the current "screen" message

	switch s.Step {
	case stepListDate:
		d, err := parseFlexDate(text, mskLoc)
		if err != nil {
			newID := b.editKB(chatID, msgID,
				"❌ Неверный формат. Введите дату <i>(например 05.06, 05.06.25 или 05.06.2026)</i>:", cancelKB())
			s.MsgID = newID
			return
		}
		b.listDate(chatID, msgID, d)

	case stepBookDateCustom:
		d, err := parseFlexDate(text, mskLoc)
		if err != nil {
			newID := b.editKB(chatID, msgID,
				"❌ Неверный формат. Введите дату <i>(например 05.06, 05.06.25 или 05.06.2026)</i>:", cancelKB())
			s.MsgID = newID
			return
		}
		s.DateStr = d.Format("02.01.2006")
		b.askTime(chatID, msgID)

	case stepBookTime:
		if _, err := time.Parse("15:04", text); err != nil {
			newID := b.editKB(chatID, msgID,
				"❌ Неверный формат. Введите время <i>(ЧЧ:ММ, например 18:00)</i>:", cancelKB())
			s.MsgID = newID
			return
		}
		s.StartStr = text
		b.askDuration(chatID, msgID)

	case stepBookDurationCustom:
		mins, err := strconv.Atoi(text)
		if err != nil || mins < 1 || mins > 360 {
			newID := b.editKB(chatID, msgID,
				"❌ Введите число минут от 1 до 360:", cancelKB())
			s.MsgID = newID
			return
		}
		s.DurationMin = mins
		if s.EditID != "" {
			b.showConfirmEdit(chatID, msgID)
		} else {
			b.askTitle(chatID, msgID)
		}

	case stepBookTitle:
		if text == "" {
			newID := b.editKB(chatID, msgID,
				"❌ Название не может быть пустым. Введите название:", cancelKB())
			s.MsgID = newID
			return
		}
		s.Title = text
		b.askTelegram(chatID, msgID)

	case stepBookTelegramCustom:
		tg := strings.TrimSpace(text)
		if !strings.HasPrefix(tg, "@") {
			tg = "@" + tg
		}
		s.TelegramStr = tg
		if s.EditID != "" {
			b.doUpdate(chatID, msgID)
		} else {
			b.askType(chatID, msgID)
		}

	case stepEditTitle:
		if text == "" {
			newID := b.editKB(chatID, msgID,
				"❌ Название не может быть пустым. Введите название:", cancelKB())
			s.MsgID = newID
			return
		}
		s.Title = text
		b.doUpdate(chatID, msgID)

	default:
		b.showMenu(chatID, 0)
	}
}

// ── Callback handler ──────────────────────────────────────────────────────────

func (b *BookingBot) handleCallback(cq *tgbotapi.CallbackQuery) {
	b.ack(cq)
	chatID := cq.Message.Chat.ID
	msgID := cq.Message.MessageID // always edit THIS message
	data := cq.Data
	s := b.sess(chatID)

	if cq.From != nil && cq.From.UserName != "" {
		b.mu.Lock()
		b.usernames[chatID] = cq.From.UserName
		b.mu.Unlock()
	}

	// auth:start is allowed without being logged in
	if data == "auth:start" {
		b.askEmail(chatID)
		return
	}

	if _, ok := b.getEmail(chatID); !ok {
		b.askEmail(chatID)
		return
	}

	switch {
	case data == "cancel":
		b.reset(chatID)
		b.showMenu(chatID, msgID)

	case data == "menu:main":
		b.showMenu(chatID, msgID)

	case data == "menu:book":
		b.startBooking(chatID, msgID)

	case data == "menu:mine":
		b.loadMyBookings(chatID, msgID, 0)

	case strings.HasPrefix(data, "menu:mine_at:"):
		idxStr := strings.TrimPrefix(data, "menu:mine_at:")
		idx, _ := strconv.Atoi(idxStr)
		b.loadMyBookings(chatID, msgID, idx)

	case data == "menu:today":
		b.listDate(chatID, msgID, time.Now().In(mskLoc))

	case data == "menu:tomorrow":
		b.listDate(chatID, msgID, time.Now().In(mskLoc).AddDate(0, 0, 1))

	case data == "menu:listdate":
		s.Step = stepListDate
		newID := b.editKB(chatID, msgID,
			"📅 Введите дату <i>(например 05.06, 05.06.25 или 05.06.2026)</i>:", cancelKB())
		s.MsgID = newID

	case data == "mybk_prev":
		b.showBookingPage(chatID, msgID, s.BookingIdx-1)

	case data == "mybk_next":
		b.showBookingPage(chatID, msgID, s.BookingIdx+1)

	case data == "datebk_prev":
		b.showDateBookingPage(chatID, msgID, s.DateBookingIdx-1)

	case data == "datebk_next":
		b.showDateBookingPage(chatID, msgID, s.DateBookingIdx+1)

	case data == "mybk_cur":
		b.showBookingPage(chatID, msgID, s.BookingIdx)

	case strings.HasPrefix(data, "room:"):
		roomNum, _ := strconv.Atoi(strings.TrimPrefix(data, "room:"))
		s.Room = roomNum
		b.askDate(chatID, msgID)

	case strings.HasPrefix(data, "date:"):
		switch strings.TrimPrefix(data, "date:") {
		case "today":
			s.DateStr = time.Now().In(mskLoc).Format("02.01.2006")
			b.askTime(chatID, msgID)
		case "tomorrow":
			s.DateStr = time.Now().In(mskLoc).AddDate(0, 0, 1).Format("02.01.2006")
			b.askTime(chatID, msgID)
		case "custom":
			s.Step = stepBookDateCustom
			newID := b.editKB(chatID, msgID,
				"📅 Введите дату <i>(например 05.06, 05.06.25 или 05.06.2026)</i>:", cancelKB())
			s.MsgID = newID
		}

	case strings.HasPrefix(data, "dur:"):
		val := strings.TrimPrefix(data, "dur:")
		if val == "custom" {
			s.Step = stepBookDurationCustom
			newID := b.editKB(chatID, msgID,
				"⏱️ Введите длительность в минутах <i>(1–360)</i>:", cancelKB())
			s.MsgID = newID
			return
		}
		mins, _ := strconv.Atoi(val)
		s.DurationMin = mins
		if s.EditID != "" {
			b.showConfirmEdit(chatID, msgID)
		} else {
			b.askTitle(chatID, msgID)
		}

	case strings.HasPrefix(data, "tg:"):
		switch strings.TrimPrefix(data, "tg:") {
		case "auto":
			uname := b.knownUsername(chatID)
			s.TelegramStr = "@" + uname
			if s.EditID != "" {
				b.doUpdate(chatID, msgID)
			} else {
				b.askType(chatID, msgID)
			}
		case "custom":
			s.Step = stepBookTelegramCustom
			newID := b.editKB(chatID, msgID,
				"📱 Введите Telegram <i>(например @username)</i>:", cancelKB())
			s.MsgID = newID
		}

	case strings.HasPrefix(data, "type:"):
		s.IsPrivate = strings.TrimPrefix(data, "type:") == "private"
		if s.EditID != "" {
			b.doUpdate(chatID, msgID)
		} else {
			b.showConfirm(chatID, msgID)
		}

	case data == "confirm:yes":
		b.doCreate(chatID, msgID)

	case data == "edit_confirm:yes":
		b.doUpdate(chatID, msgID)

	case strings.HasPrefix(data, "edit_bk:"):
		bookingID := strings.TrimPrefix(data, "edit_bk:")
		bk, err := b.svc.GetBooking(context.Background(), bookingID)
		if err != nil {
			b.editKB(chatID, msgID, "❌ Бронь не найдена.", kb(
				[]tgbotapi.InlineKeyboardButton{btn("◀️ Назад", "mybk_cur")},
			))
			return
		}
		b.showEditMenu(chatID, msgID, bk)

	case strings.HasPrefix(data, "edit_field:"):
		field := strings.TrimPrefix(data, "edit_field:")
		switch field {
		case "title":
			s.Step = stepEditTitle
			newID := b.editKB(chatID, msgID, "📝 Введите новое название:", cancelKB())
			s.MsgID = newID
		case "time":
			b.askDate(chatID, msgID)
		case "type":
			b.askType(chatID, msgID)
		case "tg":
			b.askTelegram(chatID, msgID)
		}

	case strings.HasPrefix(data, "cancel_ask:"):
		bookingID := strings.TrimPrefix(data, "cancel_ask:")
		all, _ := b.svc.ListBookings(context.Background())
		var title, startStr string
		for _, bk := range all {
			if bk.ID == bookingID {
				title = bk.Title
				startStr = bk.Start.In(mskLoc).Format("02.01  15:04")
				break
			}
		}
		b.editKB(chatID, msgID, fmt.Sprintf(
			"🗑 <b>Подтвердите удаление</b>\n\n📅 %s\n📝 %s",
			eh(startStr), eh(title),
		), kb(
			[]tgbotapi.InlineKeyboardButton{
				btn("✅ Да, удалить", "cancel_yes:"+bookingID),
				btn("◀️ Назад", "mybk_cur"),
			},
		))

	case strings.HasPrefix(data, "cancel_yes:"):
		bookingID := strings.TrimPrefix(data, "cancel_yes:")
		email, _ := b.getEmail(chatID)
		if err := b.svc.DeleteBooking(context.Background(), bookingID, email, "", false); err != nil {
			b.editKB(chatID, msgID, "❌ Не удалось удалить: "+eh(err.Error()), kb(
				[]tgbotapi.InlineKeyboardButton{btn("◀️ Назад", "mybk_cur")},
			))
			return
		}
		// Stay at same index after reload (shift back if we were at end)
		prevIdx := s.BookingIdx
		if prevIdx > 0 {
			prevIdx--
		}
		b.loadMyBookings(chatID, msgID, prevIdx)
	}
}

// ── Link token ────────────────────────────────────────────────────────────────

func (b *BookingBot) handleLinkToken(m *tgbotapi.Message, token string) {
	token = strings.ToUpper(strings.TrimSpace(token))
	telegramID := fmt.Sprintf("%d", m.From.ID)
	body := `{"token":"` + token + `","telegramId":"` + telegramID + `"}`

	req, err := newHTTPRequest("POST", b.backendURL+"/link/telegram/confirm", strings.NewReader(body))
	if err != nil {
		b.send(m.Chat.ID, "❌ Внутренняя ошибка.")
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Bot-Secret", b.botSecret)

	resp, err := httpDo(req)
	if err != nil {
		b.send(m.Chat.ID, "❌ Не удалось связаться с сервером.")
		return
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		name := m.From.FirstName
		if m.From.UserName != "" {
			name = "@" + m.From.UserName
		}
		b.send(m.Chat.ID, fmt.Sprintf("✅ Аккаунт %s привязан к сайту!", eh(name)))
	case 404:
		b.send(m.Chat.ID, "❌ Код не найден или уже использован.")
	case 403:
		b.send(m.Chat.ID, "❌ Ошибка авторизации.")
	default:
		b.send(m.Chat.ID, "❌ Что-то пошло не так.")
	}
}

// ── Error localisation ────────────────────────────────────────────────────────

func humanizeError(err error) string {
	s := err.Error()
	switch {
	case strings.Contains(s, domain.ErrOverlap.Error()):
		return "пересечение с другой бронью"
	case strings.Contains(s, domain.ErrInvalidTime.Error()):
		return "это время недоступно по правилам"
	case strings.Contains(s, domain.ErrTooLongDuration.Error()):
		return "максимум 3 часа для частных мероприятий"
	case strings.Contains(s, domain.ErrInvalidPeriod.Error()):
		return "конец раньше начала"
	case strings.Contains(s, domain.ErrInvalidRoom.Error()):
		return "неверная комната"
	case strings.Contains(s, domain.ErrPrivateDailyLimit.Error()):
		return "превышен дневной лимит частных мероприятий (3/день)"
	case strings.Contains(s, domain.ErrPrivateEveningLimit.Error()):
		return "вечерний лимит исчерпан (1 вечерняя частная/день)"
	default:
		return s
	}
}
