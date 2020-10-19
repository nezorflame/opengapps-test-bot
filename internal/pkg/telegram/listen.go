package telegram

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
)

func (b *bot) listen(updates tgbotapi.UpdatesChannel) {
	for u := range updates {
		if u.Message == nil { // ignore any non-Message Updates
			continue
		}

		switch {
		case strings.HasPrefix(u.Message.Text, b.cfg.GetString("commands.start")):
			go b.hello(u.Message)
		case strings.HasPrefix(u.Message.Text, b.cfg.GetString("commands.help")):
			log.WithField("user_id", u.Message.From.ID).Debug("Got help request")
			go b.help(u.Message)
			// case strings.HasPrefix(u.Message.Text, b.cfg.GetString("commands.your_command")):
			// go b.yourBotAction(u.Message)
		}
	}
}

func (b *bot) hello(msg *tgbotapi.Message) {
	b.reply(msg.Chat.ID, msg.MessageID, b.cfg.GetString("messages.hello"))
}

func (b *bot) help(msg *tgbotapi.Message) {
	b.reply(msg.Chat.ID, msg.MessageID, b.cfg.GetString("messages.help"))
}

func (b *bot) reply(chatID int64, msgID int, text string) {
	log.WithField("chat_id", chatID).WithField("msg_id", msgID).Debug("Sending reply")
	msg := tgbotapi.NewMessage(chatID, fmt.Sprint(text))
	if msgID != 0 {
		msg.ReplyToMessageID = msgID
	}
	msg.ParseMode = tgbotapi.ModeMarkdown

	if _, err := b.api.Send(msg); err != nil {
		log.Errorf("Unable to send the message: %v", err)
		return
	}
}
