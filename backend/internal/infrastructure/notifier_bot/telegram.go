package notifier

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	domain "Dormitory_Booking/internal/domain/booking"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var msk = time.FixedZone("MSK", 3*60*60)

// GroupNotifier sends booking events to a group/channel chat.
type GroupNotifier struct {
	bot        *tgbotapi.BotAPI
	chatID     int64
	setMsgID   func(ctx context.Context, bookingID string, msgID int)
}

// NewTelegramNotifier creates a GroupNotifier.
// setMsgID is called after a new booking is posted so the message ID can be persisted.
// Pass nil if you don't need to persist it.
func NewTelegramNotifier(token, chatID string, setMsgID func(ctx context.Context, bookingID string, msgID int)) *GroupNotifier {
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
	return &GroupNotifier{bot: bot, chatID: id, setMsgID: setMsgID}
}

func (t *GroupNotifier) NotifyNewBooking(ctx context.Context, b domain.Booking) error {
	if t == nil || t.bot == nil {
		return nil
	}
	msg, err := t.bot.Send(tgbotapi.NewMessage(t.chatID, formatNewBooking(b)))
	if err != nil {
		log.Printf("telegram group send error: %v", err)
		return err
	}
	if t.setMsgID != nil && msg.MessageID != 0 {
		t.setMsgID(ctx, b.ID, msg.MessageID)
	}
	return nil
}

func (t *GroupNotifier) NotifyDeletedBooking(_ context.Context, b domain.Booking) error {
	if t == nil || t.bot == nil {
		return nil
	}
	text := formatDeletedBooking(b, t.tgLink(b.TgMsgID))
	msg := tgbotapi.NewMessage(t.chatID, text)
	if b.TgMsgID != 0 {
		msg.ReplyToMessageID = b.TgMsgID
	}
	if _, err := t.bot.Send(msg); err != nil {
		log.Printf("telegram group send error: %v", err)
	}
	return nil
}

func (t *GroupNotifier) Send(text string) error {
	if t == nil || t.bot == nil {
		return nil
	}
	if _, err := t.bot.Send(tgbotapi.NewMessage(t.chatID, text)); err != nil {
		log.Printf("telegram send error: %v", err)
		return err
	}
	return nil
}

// tgLink returns a t.me/c/... deep link for a channel message, or "" if not applicable.
func (t *GroupNotifier) tgLink(msgID int) string {
	if msgID == 0 {
		return ""
	}
	// Supergroup/channel IDs are like -1001234567890
	// Peer ID is obtained by: strconv.FormatInt(-chatID, 10) → "1001234567890" → strip "100" prefix → "1234567890"
	s := strconv.FormatInt(-t.chatID, 10)
	if strings.HasPrefix(s, "100") {
		return fmt.Sprintf("https://t.me/c/%s/%d", s[3:], msgID)
	}
	return ""
}

// DirectNotifier sends DMs to booking owners whose TelegramID is a numeric chat ID.
type DirectNotifier struct{ bot *tgbotapi.BotAPI }

func NewDirectNotifier(bot *tgbotapi.BotAPI) *DirectNotifier { return &DirectNotifier{bot: bot} }

func (d *DirectNotifier) NotifyNewBooking(_ context.Context, b domain.Booking) error {
	chatID, err := strconv.ParseInt(b.TelegramID, 10, 64)
	if err != nil || chatID == 0 {
		return nil
	}
	_, _ = d.bot.Send(tgbotapi.NewMessage(chatID, "✅ Бронь создана!\n"+formatNewBooking(b)))
	return nil
}

func (d *DirectNotifier) NotifyDeletedBooking(_ context.Context, b domain.Booking) error {
	chatID, err := strconv.ParseInt(b.TelegramID, 10, 64)
	if err != nil || chatID == 0 {
		return nil
	}
	_, _ = d.bot.Send(tgbotapi.NewMessage(chatID, "❌ Бронь отменена!\n"+formatDeletedBooking(b, "")))
	return nil
}

// CompositeNotifier fans out to multiple notifiers.
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
	contact := b.UserEmail
	if contact == "" && b.TelegramID != "" {
		contact = "@" + b.TelegramID
	}
	priv := ""
	if b.IsPrivate {
		priv = " · 🔒 Частное"
	}
	return fmt.Sprintf(
		"📅 %s\n🕐 %s – %s\n🏠 Комн. %d\n📝 %s%s\n👤 %s",
		b.Start.In(msk).Format("02.01.2006"),
		b.Start.In(msk).Format("15:04"),
		b.End.In(msk).Format("15:04"),
		b.Room, b.Title, priv, contact,
	)
}

func formatDeletedBooking(b domain.Booking, link string) string {
	contact := b.UserEmail
	if contact == "" && b.TelegramID != "" {
		contact = "@" + b.TelegramID
	}
	header := "🔴 Бронь отменена"
	if link != "" {
		header = "🔴 Бронь отменена — " + link
	}
	return fmt.Sprintf(
		"%s\n\n📅 %s\n🕐 %s – %s\n🏠 Комн. %d\n📝 %s\n👤 %s",
		header,
		b.Start.In(msk).Format("02.01.2006"),
		b.Start.In(msk).Format("15:04"),
		b.End.In(msk).Format("15:04"),
		b.Room, b.Title, contact,
	)
}
