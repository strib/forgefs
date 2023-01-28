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
    image_url varchat(4096) NOT NULL,
    version integer NOT NULL,
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
    INSERT OR IGNORE INTO cards (id, title, house, expansion, image_url, version, json)
    VALUES (?, ?, ?, ?, ?, ?, ?);
`

func (s *SQLiteStorage) StoreCards(ctx context.Context, cards []Card) error {
	for _, card := range cards {
		j, err := json.Marshal(card)
		if err != nil {
			return err
		}

		result, err := s.db.ExecContext(
			ctx, sqlCardStore,
			card.ID, card.CardTitle, card.House, card.ExpansionEnum,
			card.FrontImage, card.ExtraCardInfo.Version, j)
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

const sqlCardNames string = `
    SELECT id, title FROM cards;
`

func (s *SQLiteStorage) GetCardTitles(ctx context.Context) (
	titles map[string]string, err error) {
	titles = make(map[string]string)
	rows, err := s.db.QueryContext(ctx, sqlCardNames)
	if err != nil {
		return nil, err
	}
	defer func() {
		closeErr := rows.Close()
		if err == nil {
			err = closeErr
		}
	}()
	for rows.Next() {
		var id, title string
		err = rows.Scan(&id, &title)
		if err != nil {
			return nil, err
		}
		titles[id] = title
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return titles, nil
}

const sqlCardImageURL string = `
    SELECT image_url FROM cards
    WHERE id=?;
`

func (s *SQLiteStorage) GetCardImageURL(ctx context.Context, id string) (
	url string, err error) {
	row := s.db.QueryRowContext(ctx, sqlCardImageURL, id)
	err = row.Scan(&url)
	if err != nil {
		return "", err
	}
	return url, nil
}

const sqlCardJSON string = `
    SELECT json FROM cards
    WHERE id=?;
`

func (s *SQLiteStorage) GetCard(ctx context.Context, id string) (
	card *Card, err error) {
	row := s.db.QueryRowContext(ctx, sqlCardJSON, id)
	var cardJSON string
	err = row.Scan(&cardJSON)
	if err != nil {
		return nil, err
	}
	var c Card
	err = json.Unmarshal([]byte(cardJSON), &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
