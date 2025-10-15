package store

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Store struct {
	db       *sqlx.DB
	Sources  *SourcesRepository
	Products *ProductsRepository
}

func NewStore(dbUrl string) (*Store, error) {
	db, err := sqlx.Connect("postgres", dbUrl)
	if err != nil {
		return nil, err
	}
	return &Store{
		db:       db,
		Sources:  NewSourcesRepository(db),
		Products: NewProductRepository(db),
	}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) ClearDB() error {

	if _, err := s.db.Exec(`DELETE FROM products`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`DELETE FROM sources`); err != nil {
		return err
	}
	return nil
}
