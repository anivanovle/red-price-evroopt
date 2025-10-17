package notifier

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (n *Notifier) HandleMessage(ctx context.Context, message *tgbotapi.Message, markup tgbotapi.ReplyKeyboardMarkup) {
	n.logger.Info("start Notifier.HandleMessage")
	select {
	case <-ctx.Done():
		n.logger.Info("HandleMessage canceled for user", "chatId", message.From.ID)
		return
	default:
	}

	length := utf8.RuneCountInString(message.Text)
	if length < 3 {
		msg := tgbotapi.NewMessage(message.From.ID, "🔍 Введите хотя бы 3 символа для поиска")
		msg.ReplyMarkup = markup
		n.Bot.Send(msg)
	} else {
		n.logger.Info("message text", message.Text, "message from", "chatId", message.From)
		n.mu.RLock()
		_, ok := n.CategoriesMap[message.Text]
		n.mu.RUnlock()
		if !ok {
			n.logger.Info("get products by title")
			posts, err := n.ProductsByTitle(ctx, *message)
			if err != nil {
				msg := tgbotapi.NewMessage(message.From.ID, "🔍 Товаров с таким названием не существует")
				msg.ParseMode = "MarkdownV2"
				msg.ReplyMarkup = markup
				n.Bot.Send(msg)
			} else {
				post := strings.Join(posts, "\n\n\n")
				msg := tgbotapi.NewMessage(message.From.ID, post)
				msg.ParseMode = "MarkdownV2"
				msg.ReplyMarkup = markup
				n.Bot.Send(msg)
			}

		} else {
			n.logger.Info("get products by category")
			posts, err := n.ProductsByCategory(ctx, *message)
			if err != nil {
				msg := tgbotapi.NewMessage(message.From.ID, "🔍 Товаров с такой категорией не существует")
				msg.ReplyMarkup = markup
				n.Bot.Send(msg)
			}
			post := strings.Join(posts, "\n\n\n")
			msg := tgbotapi.NewMessage(message.From.ID, post)
			msg.ParseMode = "MarkdownV2"
			msg.ReplyMarkup = markup
			n.Bot.Send(msg)
		}
	}

	n.logger.Info("finish Notifier.HandleMessage")

}

func (n *Notifier) ProductsByTitle(ctx context.Context, message tgbotapi.Message) ([]string, error) {
	n.logger.Info("start Notifier.ProductByTitle")
	dbModels, err := n.Products.ProductsByTitle(ctx, message.Text)
	if err != nil {
		n.logger.Error("failed to get products in notifier.ProductsByTitle")
		return nil, err
	}
	if len(dbModels) == 0 {
		return nil, errors.New("products by title is not found")
	}
	var products []Product
	for i := 0; i < len(dbModels); i++ {
		p := &Product{}
		p.modelFromStore(*dbModels[i])
		fmt.Println(p)
		products = append(products, *p)
	}
	var results []string
	for i := 0; i < len(products); i++ {
		r, err := formatProduct(&products[i])
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	n.logger.Info("len of results", "len", len(results))
	n.logger.Info("finish Notifier.ProductByTitle")
	return results, nil
}

func (n *Notifier) ProductsByCategory(ctx context.Context, message tgbotapi.Message) ([]string, error) {
	n.logger.Info("start Notifier.ProductByCategory")
	dbModels, err := n.Products.ProductsByCategory(ctx, message.Text)
	if err != nil {
		n.logger.Error("failed to get prodcuts in notifier.ProductsByCategory")
		return nil, err
	}
	if len(dbModels) == 0 {
		return nil, errors.New("products by category is not found")
	}
	var products []Product
	for i := 0; i < len(dbModels); i++ {
		p := &Product{}
		p.modelFromStore(*dbModels[i])
		products = append(products, *p)
	}
	var results []string
	for i := 0; i < len(products); i++ {
		r, err := formatProduct(&products[i])
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	n.logger.Info("len of results", "len", len(results))
	n.logger.Info("finish Notifier.ProductByCategory")
	return results, nil

}

func formatProduct(p *Product) (string, error) {
	titleEmoji := "🛒"
	priceEmoji := "💰"
	discountEmoji := "🔥"
	amountEmoji := "📦"
	categoryEmoji := "🏷️"
	dateEmoji := "📅"

	var fullPriceStr string
	if p.Fullprice != 0 {
		fullPriceStr = strconv.FormatFloat(p.Fullprice, 'f', 2, 64) + " BYN"
	}

	var discountLine string
	if p.Discount != "" {
		discountLine = fmt.Sprintf(
			"\n%s %s",
			discountEmoji,
			tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, "Скидка: "+p.Discount),
		)
	}

	var amountLine string
	if p.Amount != "" {
		amountLine = fmt.Sprintf(
			"%s %s",
			amountEmoji,
			tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, "Количество: "+p.Amount),
		)
	}

	text := fmt.Sprintf(
		"%s *%s*\n%s %s BYN ~%s~\n%s %s\n%s %s",
		titleEmoji,
		tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, p.Title),
		priceEmoji,
		tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, strconv.FormatFloat(p.Price, 'f', 2, 64)),
		tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, fullPriceStr),
		categoryEmoji,
		tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, "Категория: "+p.CatheghoryTitle),
		dateEmoji,
		tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, "Акция действует по: "+p.Date.Format("02.01.2006")),
	)

	if discountLine != "" {
		if amountLine != "" {
			text += discountLine + "\n" + amountLine
		} else {
			text += discountLine
		}
	} else if amountLine != "" {
		text += "\n" + amountLine
	}

	return text, nil
}
