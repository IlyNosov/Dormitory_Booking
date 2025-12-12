package notifier

import (
	"context"
	"fmt"
	"log"
	"strconv"

	domain "Dormitory_Booking/internal/domain/booking"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TelegramNotifier отправляет сообщение в указанный чат при создании брони.
// Токен и chat ID передаются через переменные окружения TELEGRAM_BOT_TOKEN и TELEGRAM_CHAT_ID.

type TelegramNotifier struct {
	bot    *tgbotapi.BotAPI
	chatID int64
}

func NewTelegramNotifier(token string, chatID string) *TelegramNotifier {
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
	return &TelegramNotifier{bot: bot, chatID: id}
}

func (t *TelegramNotifier) NotifyNewBooking(ctx context.Context, b domain.Booking) error {
	if t == nil || t.bot == nil {
		return nil
	}
	msg := tgbotapi.NewMessage(t.chatID, formatBooking(b))
	_, err := t.bot.Send(msg)
	if err != nil {
		log.Printf("telegram send error: %v", err)
	}
	return err
}

func formatBooking(b domain.Booking) string {
	room := int(b.Room)
	return fmt.Sprintf("🟢 Новая бронь\nКомната: %d\nС: %s\nПо: %s\nНазвание: %s\nАвтор: %s\nЧП: %t",
		room,
		b.Start.Format("02.01.2006 15:04"),
		b.End.Format("02.01.2006 15:04"),
		b.Title,
		b.TelegramID,
		b.IsPrivate,
	)
}
