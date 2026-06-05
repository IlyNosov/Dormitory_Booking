package bot

import (
	appbooking "Dormitory_Booking/internal/application/booking"
	domain "Dormitory_Booking/internal/domain/booking"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type sessionStep int

const (
	stepRoom sessionStep = iota
	stepDate
	stepStartTime
	stepDuration
	stepTitle
	stepDescription
	stepPrivate
)

type bookingSession struct {
	Step        sessionStep
	Room        domain.Room
	DateStr     string
	StartStr    string
	DurationMin int
	Title       string
	Description string
	IsPrivate   bool
}

// authStep описывает шаг в процессе авторизации по email.
type authStep int

const (
	authStepOTP authStep = iota // ожидаем ввода кода
)

type authSession struct {
	Email string
	Step  authStep
}

type BookingBot struct {
	bot      *tgbotapi.BotAPI
	svc      *appbooking.Service
	sessions map[int64]*bookingSession
	mu       sync.Mutex

	authSessions map[int64]*authSession
	authMu       sync.Mutex

	backendURL string
	botSecret  string
}

// API возвращает BotAPI для использования в других нотификаторах.
func (b *BookingBot) API() *tgbotapi.BotAPI {
	return b.bot
}

// NewBotAPI создаёт и инициализирует BotAPI без запуска polling.
// Используется для создания нотификаторов до запуска основного бота.
func NewBotAPI(token string) (*tgbotapi.BotAPI, error) {
	b, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	_, _ = b.Request(tgbotapi.DeleteWebhookConfig{})
	return b, nil
}

// StartBookingBot запускает polling на уже созданном BotAPI.
func StartBookingBot(ctx context.Context, api *tgbotapi.BotAPI, svc *appbooking.Service) *BookingBot {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := api.GetUpdatesChan(u)

	backendURL := os.Getenv("BACKEND_INTERNAL_URL")
	if backendURL == "" {
		backendURL = "http://localhost:8080"
	}

	instance := &BookingBot{
		bot:          api,
		svc:          svc,
		sessions:     make(map[int64]*bookingSession),
		authSessions: make(map[int64]*authSession),
		backendURL:   backendURL,
		botSecret:    os.Getenv("BOT_INTERNAL_SECRET"),
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case upd := <-updates:
				if upd.Message == nil {
					continue
				}
				if upd.Message.Chat == nil || !upd.Message.Chat.IsPrivate() {
					continue
				}
				instance.handleMessage(upd.Message)
			}
		}
	}()

	return instance
}

func (b *BookingBot) handleMessage(m *tgbotapi.Message) {
	text := strings.TrimSpace(m.Text)
	userID := int64(m.From.ID)

	switch {
	case strings.HasPrefix(text, "/start"):
		parts := strings.Fields(text)
		if len(parts) == 2 {
			b.handleLinkToken(m, parts[1])
			return
		}
		b.reply(m.Chat.ID,
			"Привет! Я помогу забронировать комнату.\n\n"+
				"Команды:\n"+
				"/auth EMAIL — войти через корпоративную почту\n"+
				"/book — начать бронирование\n"+
				"/link КОД — привязать Telegram к аккаунту на сайте\n"+
				"/cancel — отменить текущую операцию",
		)
		return

	case strings.HasPrefix(text, "/cancel"):
		b.mu.Lock()
		delete(b.sessions, userID)
		b.mu.Unlock()
		b.authMu.Lock()
		delete(b.authSessions, userID)
		b.authMu.Unlock()
		b.reply(m.Chat.ID, "Ок, отменил текущую операцию.")
		return

	case strings.HasPrefix(text, "/auth"):
		parts := strings.Fields(text)
		if len(parts) != 2 {
			b.reply(m.Chat.ID, "Используй: /auth email@edu.hse.ru")
			return
		}
		b.handleAuthRequest(m, parts[1])
		return

	case strings.HasPrefix(text, "/book"):
		sess := &bookingSession{Step: stepRoom}
		b.mu.Lock()
		b.sessions[userID] = sess
		b.mu.Unlock()
		b.reply(m.Chat.ID, "Давайте забронируем. Укажите номер комнаты (21, 132, 256):")
		return

	case strings.HasPrefix(text, "/link"):
		parts := strings.Fields(text)
		if len(parts) != 2 {
			b.reply(m.Chat.ID, "Используй: /link КОД\n\nКод можно получить на сайте в разделе «Привязать Telegram».")
			return
		}
		b.handleLinkToken(m, parts[1])
		return
	}

	// Если пользователь в режиме ввода OTP — обрабатываем сначала.
	b.authMu.Lock()
	authSess, hasAuth := b.authSessions[userID]
	b.authMu.Unlock()
	if hasAuth {
		b.handleAuthOTP(m, authSess)
		return
	}

	b.mu.Lock()
	sess, ok := b.sessions[userID]
	b.mu.Unlock()
	if !ok {
		b.reply(m.Chat.ID, "Не понял. Наберите /book чтобы начать бронирование или /start.")
		return
	}

	b.handleBookingStep(m, sess)
}

// handleLinkToken подтверждает привязку Telegram к платформе.
func (b *BookingBot) handleLinkToken(m *tgbotapi.Message, token string) {
	telegramID := fmt.Sprintf("%d", m.From.ID)

	payload, _ := json.Marshal(map[string]string{
		"token":      strings.ToUpper(strings.TrimSpace(token)),
		"telegramId": telegramID,
	})

	req, err := http.NewRequest("POST", b.backendURL+"/link/telegram/confirm", bytes.NewReader(payload))
	if err != nil {
		b.reply(m.Chat.ID, "Внутренняя ошибка. Попробуйте ещё раз.")
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Bot-Secret", b.botSecret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("link confirm request error: %v", err)
		b.reply(m.Chat.ID, "Не удалось связаться с сервером. Попробуйте позже.")
		return
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		name := m.From.FirstName
		if m.From.UserName != "" {
			name = "@" + m.From.UserName
		}
		b.reply(m.Chat.ID, fmt.Sprintf(
			"✅ Аккаунт %s привязан к платформе!\n\nТеперь твои брони с сайта будут отображаться здесь, а уведомления — приходить в этот чат.",
			name,
		))
	case http.StatusNotFound:
		b.reply(m.Chat.ID, "❌ Код не найден или уже использован. Получи новый код на сайте.")
	case http.StatusForbidden:
		b.reply(m.Chat.ID, "❌ Ошибка авторизации. Обратитесь к администратору.")
	default:
		b.reply(m.Chat.ID, "❌ Что-то пошло не так. Попробуйте позже.")
	}
}

// handleAuthRequest отправляет OTP на указанный email через backend.
func (b *BookingBot) handleAuthRequest(m *tgbotapi.Message, email string) {
	userID := int64(m.From.ID)

	payload, _ := json.Marshal(map[string]string{"email": strings.ToLower(strings.TrimSpace(email))})
	req, err := http.NewRequest("POST", b.backendURL+"/auth/otp/request", bytes.NewReader(payload))
	if err != nil {
		b.reply(m.Chat.ID, "Внутренняя ошибка. Попробуйте позже.")
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("auth otp request error: %v", err)
		b.reply(m.Chat.ID, "Не удалось связаться с сервером. Попробуйте позже.")
		return
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		b.authMu.Lock()
		b.authSessions[userID] = &authSession{Email: strings.ToLower(strings.TrimSpace(email))}
		b.authMu.Unlock()
		b.reply(m.Chat.ID, fmt.Sprintf("📧 Код отправлен на %s\n\nВведите 6-значный код:", email))
	case http.StatusBadRequest:
		b.reply(m.Chat.ID, "❌ Недопустимый email. Используй корпоративный адрес (edu.hse.ru).")
	case http.StatusTooManyRequests:
		b.reply(m.Chat.ID, "❌ Слишком много запросов. Попробуй через час.")
	default:
		b.reply(m.Chat.ID, "❌ Ошибка сервера. Попробуй позже.")
	}
}

// handleAuthOTP проверяет введённый OTP-код через backend.
func (b *BookingBot) handleAuthOTP(m *tgbotapi.Message, sess *authSession) {
	userID := int64(m.From.ID)
	code := strings.TrimSpace(m.Text)

	telegramID := fmt.Sprintf("%d", m.From.ID)
	payload, _ := json.Marshal(map[string]string{
		"email":      sess.Email,
		"code":       code,
		"telegramId": telegramID,
	})
	req, err := http.NewRequest("POST", b.backendURL+"/auth/otp/verify", bytes.NewReader(payload))
	if err != nil {
		b.reply(m.Chat.ID, "Внутренняя ошибка. Попробуйте позже.")
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("auth otp verify error: %v", err)
		b.reply(m.Chat.ID, "Не удалось связаться с сервером. Попробуйте позже.")
		return
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		b.authMu.Lock()
		delete(b.authSessions, userID)
		b.authMu.Unlock()

		var result struct {
			Email        string `json:"email"`
			IsAdmin      bool   `json:"isAdmin"`
			IsSuperAdmin bool   `json:"isSuperAdmin"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&result)

		text := fmt.Sprintf("✅ Вы вошли как %s", result.Email)
		if result.IsSuperAdmin {
			text += " (суперадмин)"
		} else if result.IsAdmin {
			text += " (администратор)"
		}
		text += "\n\nТеперь вы можете бронировать через сайт и бота — всё будет связано."
		b.reply(m.Chat.ID, text)
	case http.StatusUnauthorized:
		b.reply(m.Chat.ID, "❌ Неверный или устаревший код. Попробуй ещё раз или начни заново с /auth.")
	default:
		b.reply(m.Chat.ID, "❌ Ошибка сервера. Попробуй позже.")
	}
}

func (b *BookingBot) handleBookingStep(m *tgbotapi.Message, sess *bookingSession) {
	text := strings.TrimSpace(m.Text)
	userID := int64(m.From.ID)

	switch sess.Step {
	case stepRoom:
		roomNum, err := strconv.Atoi(text)
		if err != nil {
			b.reply(m.Chat.ID, "Пожалуйста, укажите номер: 21, 132 или 256.")
			return
		}
		r := domain.Room(roomNum)
		if !domain.IsValidRoom(r) {
			b.reply(m.Chat.ID, "Неверная комната. Доступны: 21, 132, 256.")
			return
		}
		sess.Room = r
		sess.Step = stepDate
		b.reply(m.Chat.ID, "Дата брони (в формате YYYY-MM-DD):")

	case stepDate:
		if _, err := time.Parse("2006-01-02", text); err != nil {
			b.reply(m.Chat.ID, "Некорректная дата. Пример: 2025-12-31")
			return
		}
		sess.DateStr = text
		sess.Step = stepStartTime
		b.reply(m.Chat.ID, "Время начала (в формате HH:MM, 24ч):")

	case stepStartTime:
		if _, err := time.Parse("15:04", text); err != nil {
			b.reply(m.Chat.ID, "Некорректное время. Пример: 18:30")
			return
		}
		sess.StartStr = text
		sess.Step = stepDuration
		b.reply(m.Chat.ID, "Длительность в минутах (30–180):")

	case stepDuration:
		mins, err := strconv.Atoi(text)
		if err != nil || mins < 1 || mins > 180 {
			b.reply(m.Chat.ID, "Некорректная длительность. Введите число минут от 1 до 180.")
			return
		}
		sess.DurationMin = mins
		sess.Step = stepTitle
		b.reply(m.Chat.ID, "Название события (коротко):")

	case stepTitle:
		if text == "" {
			b.reply(m.Chat.ID, "Название не может быть пустым.")
			return
		}
		sess.Title = text
		sess.Step = stepDescription
		b.reply(m.Chat.ID, "Описание (опционально). Введите '-' чтобы пропустить:")

	case stepDescription:
		if text != "-" {
			sess.Description = text
		}
		sess.Step = stepPrivate
		b.reply(m.Chat.ID, "Частное мероприятие? (да/нет):")

	case stepPrivate:
		lower := strings.ToLower(text)
		switch lower {
		case "да", "yes", "y", "true":
			sess.IsPrivate = true
		case "нет", "no", "n", "false":
			sess.IsPrivate = false
		default:
			b.reply(m.Chat.ID, "Ответьте 'да' или 'нет'.")
			return
		}

		loc := time.Now().Location()
		start, err := time.ParseInLocation("2006-01-02 15:04", fmt.Sprintf("%s %s", sess.DateStr, sess.StartStr), loc)
		if err != nil {
			b.reply(m.Chat.ID, "Не удалось разобрать дату/время. Попробуйте снова с /book.")
			b.reset(userID)
			return
		}
		end := start.Add(time.Duration(sess.DurationMin) * time.Minute)
		in := appbooking.CreateBookingInput{
			Start:       start,
			End:         end,
			Room:        sess.Room,
			Title:       sess.Title,
			Description: sess.Description,
			TelegramID:  fmt.Sprintf("%d", m.From.ID),
			IsPrivate:   sess.IsPrivate,
		}

		booking, err := b.svc.CreateBooking(context.Background(), in)
		if err != nil {
			msg := humanizeError(err)
			b.reply(m.Chat.ID, fmt.Sprintf("Не удалось забронировать: %s", msg))
			b.reset(userID)
			return
		}

		b.reply(m.Chat.ID, fmt.Sprintf("✅ Готово! Бронь создана:\n%s, комната %d\n%s – %s",
			booking.Title,
			booking.Room,
			booking.Start.Format("02.01 15:04"),
			booking.End.Format("15:04"),
		))
		b.reset(userID)
	}
}

func (b *BookingBot) reply(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, _ = b.bot.Send(msg)
}

func (b *BookingBot) reset(userID int64) {
	b.mu.Lock()
	delete(b.sessions, userID)
	b.mu.Unlock()
}

func humanizeError(err error) string {
	switch {
	case strings.Contains(err.Error(), domain.ErrOverlap.Error()):
		return "пересечение с другой бронью"
	case strings.Contains(err.Error(), domain.ErrInvalidTime.Error()):
		return "это время недоступно по правилам"
	case strings.Contains(err.Error(), domain.ErrTooLongDuration.Error()):
		return "слишком длинная бронь (максимум 3 часа)"
	case strings.Contains(err.Error(), domain.ErrInvalidPeriod.Error()):
		return "конец раньше начала"
	case strings.Contains(err.Error(), domain.ErrInvalidRoom.Error()):
		return "неверная комната"
	case strings.Contains(err.Error(), domain.ErrPrivateDailyLimit.Error()):
		return "превышен дневной лимит частных мероприятий"
	case strings.Contains(err.Error(), domain.ErrPrivateEveningLimit.Error()):
		return "вечерний лимит частных мероприятий исчерпан"
	default:
		return err.Error()
	}
}
