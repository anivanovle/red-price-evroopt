package store

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type SourcesRepository struct {
	db *sqlx.DB
}
type Source struct {
	Id    int    `db:"id"`
	Link  string `db:"link"`
	Title string `db:"title"`
}

func NewSourcesRepository(db *sqlx.DB) *SourcesRepository {
	return &SourcesRepository{
		db: db,
	}
}

func (s *SourcesRepository) Sources(ctx context.Context) ([]Source, error) {
	var sources []Source
	query := `SELECT * FROM sources`
	err := s.db.SelectContext(ctx, &sources, query)
	if err != nil {
		return nil, err
	}
	return sources, nil
}

func (s *SourcesRepository) AddSource(ctx context.Context, link, title string) error {
	query := `INSERT INTO sources (link, title)
		VALUES ($1, $2)
		ON CONFLICT (link) DO NOTHING;`
	_, err := s.db.ExecContext(ctx, query, link, title)
	if err != nil {
		return err
	}
	return nil
}

func (s *SourcesRepository) Category(ctx context.Context, link string) (string, error) {
	query := `SELECT title FROM sources WHERE link=$1`
	var c string
	err := s.db.GetContext(ctx, &c, query, link)
	if err != nil {
		return "", nil
	}
	return c, nil
}

func (s *SourcesRepository) Categories(ctx context.Context) ([]string, error) {
	var c []string
	query := `SELECT title FROM sources`
	err := s.db.SelectContext(ctx, &c, query)
	return c, err
}
