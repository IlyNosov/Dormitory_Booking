package bot

import (
	appbooking "Dormitory_Booking/internal/application/booking"
	domain "Dormitory_Booking/internal/domain/booking"
	"context"
	"fmt"
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

type BookingBot struct {
	bot      *tgbotapi.BotAPI
	svc      *appbooking.Service
	sessions map[int64]*bookingSession
	mu       sync.Mutex
}

func StartBookingBot(ctx context.Context, token string, svc *appbooking.Service) error {
	b, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return err
	}
	_, _ = b.Request(tgbotapi.DeleteWebhookConfig{})
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.GetUpdatesChan(u)

	bot := &BookingBot{
		bot:      b,
		svc:      svc,
		sessions: make(map[int64]*bookingSession),
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
				bot.handleMessage(upd.Message)
			}
		}
	}()

	return nil
}

func (b *BookingBot) handleMessage(m *tgbotapi.Message) {
	text := strings.TrimSpace(m.Text)
	userID := int64(m.From.ID)

	switch {
	case strings.HasPrefix(text, "/start"):
		b.reply(m.Chat.ID, "Привет! Я помогу забронировать комнату. Команды:\n/book — начать бронирование\n/cancel — отменить текущий ввод")
		return
	case strings.HasPrefix(text, "/cancel"):
		b.mu.Lock()
		delete(b.sessions, userID)
		b.mu.Unlock()
		b.reply(m.Chat.ID, "Ок, отменил текущую сессию бронирования.")
		return
	case strings.HasPrefix(text, "/book"):
		sess := &bookingSession{Step: stepRoom}
		b.mu.Lock()
		b.sessions[userID] = sess
		b.mu.Unlock()
		b.reply(m.Chat.ID, "Давайте забронируем. Укажите номер комнаты (21, 132, 256):")
		return
	}

	b.mu.Lock()
	sess, ok := b.sessions[userID]
	b.mu.Unlock()
	if !ok {
		b.reply(m.Chat.ID, "Не понял. Наберите /book чтобы начать бронирование или /start.")
		return
	}

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

		b.reply(m.Chat.ID, fmt.Sprintf("Готово! Бронь создана: %s, комната %d, %s–%s",
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
