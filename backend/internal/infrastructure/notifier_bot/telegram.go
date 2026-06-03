package notifier

import (
	"context"
	"fmt"
	"log"
	"strconv"

	domain "Dormitory_Booking/internal/domain/booking"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// GroupNotifier отправляет сообщения в групповой чат при создании/удалении брони.
type GroupNotifier struct {
	bot    *tgbotapi.BotAPI
	chatID int64
}

func NewTelegramNotifier(token string, chatID string) *GroupNotifier {
	id, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		log.Printf("invalid TELEGRAM_CHAT_ID: %v", err)
		return nil
	}
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Printf("telegram bot init error: %v", err)
		return nil
	}
	return &GroupNotifier{bot: bot, chatID: id}
}

func (t *GroupNotifier) NotifyNewBooking(_ context.Context, b domain.Booking) error {
	if t == nil || t.bot == nil {
		return nil
	}
	msg := tgbotapi.NewMessage(t.chatID, formatNewBooking(b))
	_, err := t.bot.Send(msg)
	if err != nil {
		log.Printf("telegram group send error: %v", err)
	}
	return err
}

func (t *GroupNotifier) NotifyDeletedBooking(_ context.Context, b domain.Booking) error {
	if t == nil || t.bot == nil {
		return nil
	}
	msg := tgbotapi.NewMessage(t.chatID, formatDeletedBooking(b))
	_, err := t.bot.Send(msg)
	if err != nil {
		log.Printf("telegram group send error: %v", err)
	}
	return err
}

// DirectNotifier отправляет личные сообщения владельцу брони (если его TelegramID — числовой ID).
type DirectNotifier struct {
	bot *tgbotapi.BotAPI
}

func NewDirectNotifier(bot *tgbotapi.BotAPI) *DirectNotifier {
	return &DirectNotifier{bot: bot}
}

func (d *DirectNotifier) NotifyNewBooking(_ context.Context, b domain.Booking) error {
	chatID, err := strconv.ParseInt(b.TelegramID, 10, 64)
	if err != nil || chatID == 0 {
		return nil
	}
	text := fmt.Sprintf("✅ Бронь создана!\n%s", formatNewBooking(b))
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := d.bot.Send(msg); err != nil {
		log.Printf("telegram DM send error: %v", err)
	}
	return nil
}

func (d *DirectNotifier) NotifyDeletedBooking(_ context.Context, b domain.Booking) error {
	chatID, err := strconv.ParseInt(b.TelegramID, 10, 64)
	if err != nil || chatID == 0 {
		return nil
	}
	text := fmt.Sprintf("❌ Бронь отменена!\n%s", formatDeletedBooking(b))
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := d.bot.Send(msg); err != nil {
		log.Printf("telegram DM send error: %v", err)
	}
	return nil
}

// CompositeNotifier последовательно вызывает несколько нотификаторов.
type CompositeNotifier struct {
	notifiers []interface {
		NotifyNewBooking(context.Context, domain.Booking) error
		NotifyDeletedBooking(context.Context, domain.Booking) error
	}
}

func NewCompositeNotifier(nn ...interface {
	NotifyNewBooking(context.Context, domain.Booking) error
	NotifyDeletedBooking(context.Context, domain.Booking) error
}) *CompositeNotifier {
	return &CompositeNotifier{notifiers: nn}
}

func (c *CompositeNotifier) NotifyNewBooking(ctx context.Context, b domain.Booking) error {
	for _, n := range c.notifiers {
		_ = n.NotifyNewBooking(ctx, b)
	}
	return nil
}

func (c *CompositeNotifier) NotifyDeletedBooking(ctx context.Context, b domain.Booking) error {
	for _, n := range c.notifiers {
		_ = n.NotifyDeletedBooking(ctx, b)
	}
	return nil
}

func formatNewBooking(b domain.Booking) string {
	return fmt.Sprintf("🟢 Новая бронь\nКомната: %d\nС: %s\nПо: %s\nНазвание: %s\nАвтор: %s\nЧП: %t",
		int(b.Room),
		b.Start.Format("02.01.2006 15:04"),
		b.End.Format("02.01.2006 15:04"),
		b.Title,
		b.TelegramID,
		b.IsPrivate,
	)
}

func formatDeletedBooking(b domain.Booking) string {
	return fmt.Sprintf("🔴 Бронь отменена\nКомната: %d\nС: %s\nПо: %s\nНазвание: %s\nАвтор: %s",
		int(b.Room),
		b.Start.Format("02.01.2006 15:04"),
		b.End.Format("02.01.2006 15:04"),
		b.Title,
		b.TelegramID,
	)
}
