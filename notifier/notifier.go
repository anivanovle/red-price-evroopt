package notifier

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	tgbotAPI "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"parser1.0/store"
)

const UpdateTimeOut = 30

type Product struct {
	Id              int
	Title           string
	Amount          string
	Price           float64
	Discount        string
	Fullprice       float64
	CatheghoryTitle string
	SourceLink      string
	Date            time.Time
}

type Products interface {
	ProductsByCategory(ctx context.Context, catheghory string) ([]*store.Product, error)
	ProductsByTitle(ctx context.Context, title string) ([]*store.Product, error)
}

type Sources interface {
	Category(ctx context.Context, link string) (string, error)
	Categories(ctx context.Context) ([]string, error)
}

type Notifier struct {
	Bot           *tgbotAPI.BotAPI
	Products      Products
	Sources       Sources
	CategoriesMap map[string]struct{}
	logger        *slog.Logger
	mu            sync.RWMutex
}

func NewNotifier(products Products, sources Sources, bot *tgbotAPI.BotAPI, logger *slog.Logger) *Notifier {
	return &Notifier{
		Bot:           bot,
		Products:      products,
		Sources:       sources,
		CategoriesMap: map[string]struct{}{},
		logger:        logger,
	}
}
func (n *Notifier) Run(ctx context.Context) error {
	n.logger.Info("Start notifier.Run")
	n.Bot.Debug = true
	updateConfig := tgbotAPI.NewUpdate(0)
	updateConfig.Timeout = UpdateTimeOut
	updates := n.Bot.GetUpdatesChan(updateConfig)

	n.logger.Info("get categories")
	c, err := n.Sources.Categories(context.Background())

	if err != nil {
		n.logger.Info("failed get categories", "error", err)
		return err
	}

	for _, c := range c {
		fmt.Println(c)
		n.CategoriesMap[c] = struct{}{}
	}

	for {
		select {
		case <-ctx.Done():
			n.logger.Info("notifier stopped by context cancel")
			return nil
		case u, ok := <-updates:
			if !ok {
				n.logger.Info("chan with updates closed notifier stops")
				return nil
			} else {
				n.logger.Info("receive messages")
				if u.Message != nil {
					markup := n.BoardFromCategory(c)
					if u.Message.Text == "/start" {
						n.SendKeyBoard(u.FromChat().ID, markup) // кнопки отправил
					}
					/*Вынес в отдельную горутину но не знаю по идее у нас тут может быть и работа с бд и вычисления
					как в таком случае правильно делается асинхронност?*/
					go func(up *tgbotAPI.Update) {
						ctxMessage, cancel := context.WithTimeout(ctx, 10*time.Second)
						defer cancel()
						n.HandleMessage(ctxMessage, up.Message, markup)
					}(&u)
				}
			}

		}

	}

}

func (n *Notifier) SendKeyBoard(chatId int64, markup tgbotAPI.ReplyKeyboardMarkup) {
	msg := tgbotAPI.NewMessage(chatId, "Выберите категорию:")
	msg.ReplyMarkup = markup
	n.Bot.Send(msg)
}

func (n *Notifier) BoardFromCategory(categories []string) tgbotAPI.ReplyKeyboardMarkup {
	var rows [][]tgbotAPI.KeyboardButton
	row := []tgbotAPI.KeyboardButton{}
	for i, c := range categories {
		row = append(row, tgbotAPI.NewKeyboardButton(c))
		if (i+1)%3 == 0 {
			rows = append(rows, row)
			row = []tgbotAPI.KeyboardButton{}
		}
	}

	if len(row) > 0 {
		rows = append(rows, row)
	}
	keyboard := tgbotAPI.NewReplyKeyboard(rows...)
	return keyboard
}

func (p *Product) modelFromStore(dbmodel store.Product) {
	priceFloat := float64(dbmodel.Price) / 100
	fullPriceFloat := float64(dbmodel.Fullprice) / 100
	p.Id = dbmodel.Id
	p.Title = dbmodel.Title
	p.Price = priceFloat
	p.SourceLink = dbmodel.SourceLink
	p.CatheghoryTitle = dbmodel.CatheghoryTitle
	p.Discount = dbmodel.Discount
	p.Amount = dbmodel.Amount
	p.Date = dbmodel.Date
	p.Fullprice = fullPriceFloat
}
