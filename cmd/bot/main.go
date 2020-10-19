package main

import (
	"context"
	"os"

	"github.com/nezorflame/opengapps-test-bot/internal/pkg/config"
	"github.com/nezorflame/opengapps-test-bot/internal/pkg/storage"
	"github.com/nezorflame/opengapps-test-bot/internal/pkg/telegram"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

var configName string

func init() {
	// get flags, init logger
	pflag.StringVar(&configName, "config", "config", "Config file name")
	level := pflag.String("log-level", "INFO", "Logrus log level (DEBUG, WARN, etc.)")
	pflag.Parse()

	logLevel, err := log.ParseLevel(*level)
	if err != nil {
		log.WithError(err).Fatal("Unable to parse log level")
	}
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetOutput(os.Stdout)
	log.SetLevel(logLevel)

	if configName == "" {
		pflag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {
	// init flags and ctx
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// init config and tracing
	log.Info("Starting the bot")
	cfg, err := config.New(configName)
	if err != nil {
		log.WithError(err).Fatal("Unable to parse config")
	}
	log.Info("Config parsed")

	// init DB
	db, err := storage.New(cfg.GetString("db.path"), cfg.GetDuration("db.timeout"))
	if err != nil {
		log.WithError(err).Fatal("Unable to init DB")
	}
	log.Info("DB initiated")

	// create bot
	bot, err := telegram.NewBot(cfg)
	if err != nil {
		log.WithError(err).Fatal("Unable to create bot")
	}
	log.Info("Bot created")

	// start the bot
	log.Info("Starting the bot")
	if err = bot.Start(-1); err != nil {
		log.WithError(err).Fatal("Unable to start the bot")
	}

	// graceful shutdown
	log.Debug("Initiating system signal watcher")
	<-gracefulShutdown(ctx, bot, db)
}
