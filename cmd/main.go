package main

import (
	"context"
	"database/sql"
	"log"
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
		return
	}

	ticker := Ticker()

	ctx := context.Background()
	// Вот тут непоняно ты хочешь чтобы я вообще не выходил из приложения или как его тогда правильно класть
	if err := StartBot(ctx, c, ticker.C); err != nil {
		c.Logger.Info("Bot fell")
		log.Fatal("Bot upal")
	}

	for {
		select {
		case <-ticker.C:
			ticker.Reset(Diff())
			if err := StartBot(ctx, c, ticker.C); err != nil {
				c.Logger.Info("Upal restart 3...2...1..")
				log.Fatal("Bot upal")
			}
		case <-ctx.Done():
			c.Logger.Info("Upal...")
			return
		}
	}

}

func StartBot(ctx context.Context, c *config.Config, tick <-chan time.Time) error {
	store, err := store.NewStore(c.DbUrl)
	if err != nil {
		c.Logger.Error("failed to create store obj:", "error", err)
		return err
	}
	defer store.Close()

	if err := store.ClearDB(); err != nil {
		c.Logger.Error("failed to clear db:", "error", err)
		return err
	}

	bot, err := tgbotapi.NewBotAPI(c.BotToken)
	if err != nil {
		c.Logger.Error("failed to create tgbot obj:", "error", err)
		return err
	}
	p, err := parser.NewParser(store.Sources, store.Products, c.Logger)

	if err != nil {
		c.Logger.Error("failed to create parser obj:", "error", err)
		return err
	}
	if err := p.InitSources(ctx); err != nil {
		c.Logger.Error("failed to init sources:", "error", err)
		return err
	}
	if err := p.ParsingProducts(context.Background()); err != nil {
		c.Logger.Error("failed to parse products:", "error", err)

		return err
	}
	n := notifier.NewNotifier(store.Products, store.Sources, bot, c.Logger)

	go func() {
		if err := n.Run(ctx); err != nil {
			c.Logger.Error("failed in notifier component:", "error", err)
		}
	}()

	select {
	case <-ctx.Done():
		c.Logger.Info("Context is canceled app stops")
		return err
	case <-tick:
		c.Logger.Info("Middnight, bot restarted")
		return nil
	}

}

func Ticker() time.Ticker {
	duration := Diff()
	timer := time.NewTicker(duration)

	return *timer
}

func Diff() time.Duration {
	now := time.Now()
	next := now.Add(24 * time.Hour).Truncate(24 * time.Hour)
	diff := next.Sub(now)
	return diff
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
