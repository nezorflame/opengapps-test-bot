package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// bot describes Telegram bot
type bot struct {
	api *tgbotapi.BotAPI
	cfg *viper.Viper

	lastOffset int
}

const (
	defaultLastOffset = -1
)

// NewBot creates new instance of Bot
func NewBot(opts ...OptionFunc) (*bot, error) {
	b := &bot{}
	for _, o := range opts {
		if err := o(b); err != nil {
			return nil, fmt.Errorf("unable to create new bot: %w", err)
		}
	}

	if b.lastOffset == 0 {
		log.WithField("offset", defaultLastOffset).Warn("Setting default last offset")
		b.lastOffset = defaultLastOffset
	}

	api, err := tgbotapi.NewBotAPI(b.cfg.GetString("telegram.token"))
	if err != nil {
		return nil, fmt.Errorf("unable to connect to Telegram: %w", err)
	}
	b.api = api

	if err = tgbotapi.SetLogger(log.WithField("source", "telegram-api")); err != nil {
		return nil, fmt.Errorf("unable to set Telegram logger: %w", err)
	}

	if b.cfg.GetBool("telegram.debug") {
		log.Debug("Enabling debug mode for bot")
		api.Debug = true
	}

	log.Debugf("Authorized on account %s", api.Self.UserName)
	return b, nil
}

// Start starts to listen the bot updates channel
func (b *bot) Start() error {
	update := tgbotapi.NewUpdate(b.lastOffset + 1)
	update.Timeout = b.cfg.GetInt("telegram.timeout")
	updates, err := b.api.GetUpdatesChan(update)
	if err != nil {
		return fmt.Errorf("unable to start listening to bot updates: %w", err)
	}

	go b.listen(updates)
	return nil
}

// Close stops the bot
func (b *bot) Close() error {
	b.api.StopReceivingUpdates()
	return nil
}

// Name returns the bot identifier
func (b *bot) Name() string {
	return "bot"
}
