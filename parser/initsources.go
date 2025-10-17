package parser

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/chromedp/chromedp"
)

const ClickTimeOut = 5

// Получение категории и ссылки источников

func (p *Parser) InitSources(ctx context.Context) error {

	p.logger.Info("start parser.InitSources")

	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	url := os.Getenv("ENTRY_PAGE")
	if url == "" {
		p.logger.Error("url is empty")

		return errors.New("entry_page is empty")
	}

	p.logger.Info("open browser with chromedp")
	if err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(ClickTimeOut*time.Second),
	); err != nil {
		p.logger.Error("failed to get entry page")

		return err
	}
	var categoryDivs []string
	p.logger.Info("start clicking on checkbox with category")
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll("div[class^='Checkboxes_item'] > span:not([class^='Checkbox_success'])"))
				.map(el => el.innerText)
		`, &categoryDivs),
	); err != nil {
		p.logger.Error("failed get category name and category link")

		return err
	}

	p.logger.Info("find:", "category", len(categoryDivs))

	for i, name := range categoryDivs {
		jsClick := fmt.Sprintf(`
			document.querySelectorAll("div[class^='Checkboxes_item'] span[class^='Checkbox_success']")[%d].click()
		`, i)

		var currentURL string
		if err := chromedp.Run(ctx,
			chromedp.Evaluate(jsClick, nil),
			chromedp.Sleep(2*time.Second),
			chromedp.Location(&currentURL),
			chromedp.Evaluate(jsClick, nil),
			chromedp.Sleep(1*time.Second),
		); err != nil {
			p.logger.Error("failed to click category block", "error", err)
			continue
		}

		err := p.Sources.AddSource(ctx, currentURL, name)
		if err != nil {
			p.logger.Error("failed to add source in db", "error", err)
			return err
		}
	}
	return nil

}
