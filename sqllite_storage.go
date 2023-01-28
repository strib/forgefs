package forgefs

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStorage struct {
	db *sql.DB
}

func NewSQLiteStorage(
	ctx context.Context, file string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		return nil, err
	}
	s := &SQLiteStorage{
		db: db,
	}

	err = s.init(ctx)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *SQLiteStorage) Shutdown() error {
	return s.db.Close()
}

const sqlCardsCreate string = `
    CREATE TABLE IF NOT EXISTS cards (
    id varchar(36) NOT NULL PRIMARY KEY,
    title varchar(1024) NOT NULL,
    house varchar(64) NOT NULL,
    expansion varchar(64) NOT NULL,
    json blob NOT NULL
);`

func (s *SQLiteStorage) init(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, sqlCardsCreate)
	if err != nil {
		return err
	}

	return nil
}

const sqlCardsCount string = `
    SELECT COUNT(*) FROM cards;
`

func (s *SQLiteStorage) GetCardsCount(ctx context.Context) (
	count int, err error) {
	row := s.db.QueryRowContext(ctx, sqlCardsCount)
	err = row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

const sqlCardStore string = `
    INSERT OR IGNORE INTO cards (id, title, house, expansion, json)
    VALUES (?, ?, ?, ?, ?);
`

func (s *SQLiteStorage) StoreCards(ctx context.Context, cards []Card) error {
	for _, card := range cards {
		j, err := json.Marshal(card)
		if err != nil {
			return err
		}

		result, err := s.db.ExecContext(
			ctx, sqlCardStore,
			card.ID, card.CardTitle, card.House, card.ExpansionEnum, j)
		if err != nil {
			return err
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if rowsAffected != 1 {
			return errors.New("card not inserted")
		}
	}
	return nil
}
