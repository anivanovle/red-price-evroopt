package main

import (
	"database/sql"
	"log"
	"log/slog"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"parser1.0/config"
	"parser1.0/notifier"
	"parser1.0/parser"
	"parser1.0/store"
)

func main() {
	c := config.MustConfig()
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		RunMigrations(c)

		os.Exit(1)
	}
	go scheduleRestart(c.Logger)
	store, err := store.NewStore(c.DbUrl)
	if err != nil {
		c.Logger.Error("failed to create store obj:", "error", err)

		os.Exit(1)
	}
	defer store.Close()

	bot, err := tgbotapi.NewBotAPI(c.BotToken)
	if err != nil {
		c.Logger.Error("failed to create tgbot obj:", "error", err)

		os.Exit(1)
	}
	p, err := parser.NewParser(store.Sources, store.Products, c.Logger)

	if err != nil {
		c.Logger.Error("failed to create parser obj:", "error", err)

		os.Exit(1)
	}
	if err := p.InitSources(); err != nil {
		c.Logger.Error("failed to init sources:", "error", err)

		os.Exit(1)
	}
	if err := p.Run(); err != nil {
		c.Logger.Error("failed to parse products:", "error", err)

		return
	}
	n := notifier.NewNotifier(store.Products, store.Sources, bot, c.Logger)
	if err := n.Run(); err != nil {
		c.Logger.Error("failed in notifier component:", "error", err)

		os.Exit(1)
	}

}

func scheduleRestart(logger *slog.Logger) {
	for {
		now := time.Now()
		next := now.Add(24 * time.Hour).Truncate(24 * time.Hour)
		diff := next.Sub(now)

		time.Sleep(diff)
		logger.Info("bot finishes work and restarting")
		os.Exit(0)
	}
}

func RunMigrations(c *config.Config) {
	db, err := sql.Open("postgres", c.DbUrl)
	if err != nil {
		log.Fatal(err)
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://./migrations", "postgres", driver)
	if err != nil {
		log.Fatal(err)
	}
	m.Up()
	c.Logger.Info("Migrations applyed successfully")
}
