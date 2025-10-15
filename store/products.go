package store

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type ProductsRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) *ProductsRepository {
	return &ProductsRepository{db: db}
}

type Product struct {
	Id              int       `db:"id"`
	Title           string    `db:"title"`
	Amount          string    `db:"amount"`
	Price           int       `db:"price"`
	Discount        string    `db:"discount"`
	Fullprice       int       `db:"fullprice"`
	CatheghoryTitle string    `db:"category_title"`
	SourceLink      string    `db:"source_link"`
	Date            time.Time `db:"date_until"`
}

func (r *ProductsRepository) Save(ctx context.Context, p Product) error {
	query := `INSERT INTO products 
		(title, amount, price, discount, fullprice, category_title, source_link, date_until)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT (title) DO NOTHING;`
	_, err := r.db.ExecContext(ctx, query,
		p.Title, p.Amount, p.Price, p.Discount, p.Fullprice,
		p.CatheghoryTitle, p.SourceLink, p.Date,
	)
	return err
}

func (r *ProductsRepository) ProductsByTitle(ctx context.Context, title string) ([]*Product, error) {
	var p []*Product
	query := `SELECT * FROM products WHERE title ILIKE $1`
	err := r.db.SelectContext(ctx, &p, query, "%"+title+"%")
	return p, err
}

func (r *ProductsRepository) ProductsByCategory(ctx context.Context, category string) ([]*Product, error) {
	var p []*Product
	query := `SELECT * FROM products WHERE category_title=$1`
	err := r.db.SelectContext(ctx, &p, query, category)
	return p, err
}
