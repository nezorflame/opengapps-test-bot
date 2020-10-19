package telegram

import (
	"errors"

	"github.com/spf13/viper"
)

type OptionFunc func(*bot) error

func WithConfig(cfg *viper.Viper) OptionFunc {
	return func(b *bot) error {
		if cfg == nil {
			return errors.New("empty config")
		}
		b.cfg = cfg
		return nil
	}
}

func WithLastOffset(offset int) OptionFunc {
	return func(b *bot) error {
		b.lastOffset = offset
		return nil
	}
}
