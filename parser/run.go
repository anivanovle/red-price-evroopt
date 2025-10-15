package parser

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"parser1.0/store"
)

// Парсинг продуктов по источникам полученным в InitSources()

func (p *Parser) Run() error {
	p.logger.Info("Start Parser.Run")

	sources, err := p.Sources.Sources(context.Background())
	if err != nil {
		p.logger.Error("failed to get sources", "error", err)
		return err
	}

	sourceChan := make(chan string, len(sources))
	productChan := make(chan Product, 10)
	for _, source := range sources {
		sourceChan <- source.Link
	}
	close(sourceChan)
	var wgTCP sync.WaitGroup

	for i := 0; i < p.PoolTCPCount; i++ {
		wgTCP.Add(1)
		p.logger.Info("Add worker in TCP pool")
		go p.workerTCP(sourceChan, productChan, &wgTCP)
	}

	go func() {
		wgTCP.Wait()
		close(productChan)
	}() //Отдельная горутина ждет пока последний обработчик из пула TCP положит полученнй продукт в productChan и закрывает его

	var wgDB sync.WaitGroup
	for i := 0; i < p.PoolPostgresCount; i++ {
		wgDB.Add(1)
		p.logger.Info("Add worker in DB pool")
		go p.workerDB(productChan, p.Sources, p.Products, &wgDB)
	}

	wgDB.Wait()
	p.logger.Info("finish Parser.Run")
	return nil
}

func (p *Parser) workerTCP(sourceChan <-chan string, productChan chan<- Product, wg *sync.WaitGroup) {
	p.logger.Info("start worker TCP")
	defer wg.Done()
	for source := range sourceChan {
		body, err := Html(source)
		if err != nil {
			fmt.Println(err)
			continue
		}
		err = parseProduct(body, productChan, source)
		if err != nil {
			continue
		}
	}

	p.logger.Info("finish worker TCP")

}

func (p *Parser) workerDB(productChan <-chan Product, sources Sources, products Products, wg *sync.WaitGroup) {
	p.logger.Info("start DB worker")
	defer wg.Done()
	for product := range productChan {
		c, err := sources.Category(context.Background(), product.SourceLink)
		if err != nil {
			p.logger.Error("failed to get category", "error", err)
		}
		p.logger.Info("get category", "category", c)
		product.CategoryTitle = c
		err = products.Save(context.Background(), product.modelToStore())
		if err != nil {
			p.logger.Error("failed to save product in db", "error", err)
		}
	}
	p.logger.Info("finish worker TCP")
}

func Html(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "+
		"AppleWebKit/537.36 (KHTML, like Gecko) "+
		"Chrome/123.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func parseProduct(body []byte, productChan chan<- Product, source string) error {
	var product Product
	product.SourceLink = source
	r := bytes.NewReader(body)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return err
	}

	doc.Find("article[class^='Promotion_product']").Each(func(i int, s *goquery.Selection) {
		title, amount, err := TitleAndAmount(s)
		if err != nil {
			return
		}

		price, err := Price(s)
		if err != nil {
			return
		}

		fullPrice, err := FullPrice(s)
		if err != nil {
			return
		}

		date, err := Date(s)
		if err != nil {
			return
		}

		discount, err := Discount(s)
		if err != nil {
			return
		}
		product = Product{
			Title:      title,
			Amount:     amount,
			Price:      price,
			Fullprice:  fullPrice,
			Date:       date,
			Discount:   discount,
			SourceLink: source,
		}
		productChan <- product

	})
	return nil

}

func TitleAndAmount(s *goquery.Selection) (string, string, error) {
	text := s.Find("div[class^='Promotion_name']").Text()
	re := regexp.MustCompile(`\d.+`)
	amount := string(re.Find([]byte(text)))
	title := strings.TrimRight(text, amount)
	return title, amount, nil

}

func Price(s *goquery.Selection) (int, error) {
	text := s.Find("div[class^='price_mainPrice']").Text()
	text = strings.ReplaceAll(text, ",", ".")
	f, err := strconv.ParseFloat(text, 64)
	i := int64(f * 100)
	if err != nil {
		return 0, err
	}
	return int(i), nil
}

func FullPrice(s *goquery.Selection) (int, error) {
	text := s.Find("span[class^='price_discountPrice']").Text()
	text = strings.ReplaceAll(text, ",", ".")
	f, err := strconv.ParseFloat(text, 64)
	i := int64(f * 100)
	if err != nil {
		return 0, err
	}
	return int(i), nil
}

func Date(s *goquery.Selection) (time.Time, error) {
	text := s.Find("div[class^='Promotion_date']").Text()
	dateStr := strings.TrimPrefix(text, "Только по ") + ".2025"
	date, err := time.Parse("02.01.2006", dateStr)
	if err != nil {
		return time.Time{}, err
	}
	return date.UTC(), nil
}

func Discount(s *goquery.Selection) (string, error) {
	text := s.Find("div[class^='Promotion_discount']").Text()
	return text, nil
}

func (p *Product) modelToStore() store.Product {
	return store.Product{
		Title:           p.Title,
		Amount:          p.Amount,
		Price:           p.Price,
		Fullprice:       p.Fullprice,
		Date:            p.Date,
		Discount:        p.Discount,
		SourceLink:      p.SourceLink,
		CatheghoryTitle: p.CategoryTitle,
	}
}
