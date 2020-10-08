package telegram

import (
	"errors"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Bot describes Telegram bot
type Bot struct {
	api *tgbotapi.BotAPI
	cfg *viper.Viper
}

// NewBot creates new instance of Bot
func NewBot(cfg *viper.Viper) (*Bot, error) {
	if cfg == nil {
		return nil, errors.New("empty config")
	}

	api, err := tgbotapi.NewBotAPI(cfg.GetString("telegram.token"))
	if err != nil {
		return nil, fmt.Errorf("unable to connect to Telegram: %w", err)
	}

	if err = tgbotapi.SetLogger(log.WithField("source", "telegram-api")); err != nil {
		return nil, fmt.Errorf("unable to set Telegram logger: %w", err)
	}

	if cfg.GetBool("telegram.debug") {
		log.Debug("Enabling debug mode for bot")
		api.Debug = true
	}

	log.Debugf("Authorized on account %s", api.Self.UserName)
	return &Bot{api: api, cfg: cfg}, nil
}

// Start starts to listen the bot updates channel
func (b *Bot) Start(lastOffset int) error {
	update := tgbotapi.NewUpdate(lastOffset + 1)
	update.Timeout = b.cfg.GetInt("telegram.timeout")
	updates, err := b.api.GetUpdatesChan(update)
	if err != nil {
		return fmt.Errorf("unable to start listening to bot updates: %w", err)
	}

	go b.listen(updates)
	return nil
}

// Close stops the bot
func (b *Bot) Close() error {
	b.api.StopReceivingUpdates()
	return nil
}

// Name returns the bot identifier
func (b *Bot) Name() string {
	return "bot"
}
