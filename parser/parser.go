package parser

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"time"

	"parser1.0/store"
)

type Product struct {
	Id            int
	Title         string
	Amount        string
	Price         int
	Discount      string
	Fullprice     int
	CategoryTitle string
	SourceLink    string
	Date          time.Time
}

type Source struct {
	Id    int
	Title string
	Link  string
}

type Sources interface {
	Sources(ctx context.Context) ([]store.Source, error)
	AddSource(ctx context.Context, link, title string) error
	Category(ctx context.Context, link string) (string, error)
}

type Products interface {
	Save(ctx context.Context, product store.Product) error
}

type Parser struct {
	Sources           Sources
	Products          Products
	PoolPostgresCount int
	PoolTCPCount      int
	logger            *slog.Logger
}

func NewParser(sources Sources, products Products, logger *slog.Logger) (*Parser, error) {
	pcStr := os.Getenv("POOL_POSTGRES_COUNT")
	tcStr := os.Getenv("POOL_TCP_COUNT")
	pc, err := strconv.Atoi(pcStr)
	if err != nil {
		return nil, err
	}
	tc, err := strconv.Atoi(tcStr)
	if err != nil {
		return nil, err
	}
	return &Parser{
		Sources:           sources,
		Products:          products,
		logger:            logger,
		PoolPostgresCount: pc,
		PoolTCPCount:      tc,
	}, nil
}
