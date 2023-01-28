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

const sqlDecksCreate string = `
    CREATE TABLE IF NOT EXISTS decks (
    id varchar(36) NOT NULL PRIMARY KEY,
    name varchar(1024) NOT NULL,
    expansion varchar(64) NOT NULL,
    sas integer NOT NULL,
    sas_version integer NOT NULL,
    owned_by_me boolean NOT NULL,
    funny boolean NOT NULL,
    wish_list boolean NOT NULL,
    json blob NOT NULL
);`

func (s *SQLiteStorage) init(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, sqlCardsCreate)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, sqlDecksCreate)
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

const sqlDecksCount string = `
    SELECT COUNT(*) FROM decks;
`

func (s *SQLiteStorage) GetDecksCount(ctx context.Context) (
	count int, err error) {
	row := s.db.QueryRowContext(ctx, sqlDecksCount)
	err = row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

const sqlDeckStore string = `
    INSERT OR REPLACE INTO decks (id, name, expansion, sas, sas_version, owned_by_me, funny, wish_list, json)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
`

func (s *SQLiteStorage) StoreDecks(ctx context.Context, decks []Deck) error {
	for _, deck := range decks {
		j, err := json.Marshal(deck)
		if err != nil {
			return err
		}

		result, err := s.db.ExecContext(
			ctx, sqlDeckStore,
			deck.DeckInfo.KeyforgeID, deck.DeckInfo.Name,
			deck.DeckInfo.Expansion, deck.DeckInfo.SasRating,
			deck.DeckInfo.AercVersion, deck.OwnedByMe, deck.Funny,
			deck.Wishlist, j)
		if err != nil {
			return err
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if rowsAffected != 1 {
			return errors.New("deck not inserted")
		}
	}
	return nil
}

const sqlMyDeckNames string = `
    SELECT id, name FROM decks
    WHERE owned_by_me = 1;
`

func (s *SQLiteStorage) GetMyDeckNames(ctx context.Context) (
	names map[string]string, err error) {
	names = make(map[string]string)
	rows, err := s.db.QueryContext(ctx, sqlMyDeckNames)
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
		var id, name string
		err = rows.Scan(&id, &name)
		if err != nil {
			return nil, err
		}
		names[id] = name
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return names, nil
}

const sqlDeckJSON string = `
    SELECT json FROM decks
    WHERE id=?;
`

func (s *SQLiteStorage) GetDeck(ctx context.Context, id string) (
	deck *Deck, err error) {
	row := s.db.QueryRowContext(ctx, sqlDeckJSON, id)
	var deckJSON string
	err = row.Scan(&deckJSON)
	if err != nil {
		return nil, err
	}
	var d Deck
	err = json.Unmarshal([]byte(deckJSON), &d)
	if err != nil {
		return nil, err
	}
	return &d, nil
}
