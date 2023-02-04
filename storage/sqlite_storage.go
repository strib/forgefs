package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3" // load sqlite driver
	"github.com/strib/forgefs"
	"github.com/strib/forgefs/filter"
)

const (
	// If the existing data version is lower than this, we should
	// delete the DB on startup.   If it's larger, we should error.
	sqlDataVersion = 1
)

// SQLiteStorage stores deck and card info in an on-disk SQLite file.
type SQLiteStorage struct {
	db *sql.DB
}

var _ forgefs.Storage = (*SQLiteStorage)(nil)

// NewSQLiteStorage creates a new SQLiteStorage instance.
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

// Shutdown shuts down the storage instance.
func (s *SQLiteStorage) Shutdown() error {
	return s.db.Close()
}

const sqlVersionCreate string = `
    CREATE TABLE IF NOT EXISTS version (
    version integer NOT NULL PRIMARY KEY
);`

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
    aerc integer NOT NULL,
    a real NOT NULL,
    e real NOT NULL,
    r real NOT NULL,
    c real NOT NULL,
    f real NOT NULL,
    d real NOT NULL,
    house1 varchar(64) NOT NULL,
    house2 varchar(64) NOT NULL,
    house3 varchar(64) NOT NULL,
    owned_by_me boolean NOT NULL,
    funny boolean NOT NULL,
    wish_list boolean NOT NULL,
    json blob NOT NULL
);`

const sqlVersion string = `
    SELECT COALESCE(MAX(version), 0) FROM version;
`

const sqlDropTables string = `
    DROP TABLE IF EXISTS decks;
    DROP TABLE IF EXISTS cards;
`

const sqlWriteVersion string = `
    INSERT INTO version (version) VALUES (?);
`

func (s *SQLiteStorage) init(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, sqlVersionCreate)
	if err != nil {
		return err
	}

	var currVersion int
	row := s.db.QueryRowContext(ctx, sqlVersion)
	err = row.Scan(&currVersion)
	if err != nil {
		return err
	}
	switch {
	case currVersion < sqlDataVersion:
		// Drop all tables.
		_, err := s.db.ExecContext(ctx, sqlDropTables)
		if err != nil {
			return err
		}
		fallthrough
	case currVersion == 0:
		_, err := s.db.ExecContext(ctx, sqlWriteVersion, sqlDataVersion)
		if err != nil {
			return err
		}
	case currVersion > sqlDataVersion:
		return fmt.Errorf(
			"Current DB version (%d) is bigger than the code version (%d)",
			currVersion, sqlDataVersion)
	default:
		// Nothing to do since the version matches.
	}

	_, err = s.db.ExecContext(ctx, sqlCardsCreate)
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

// GetCardsCount implements the forgefs.Storage interface.
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

// StoreCards implements the forgefs.Storage interface.
func (s *SQLiteStorage) StoreCards(
	ctx context.Context, cards []forgefs.Card) error {
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

// GetCardTitles implements the forgefs.Storage interface.
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

// GetCardImageURL implements the forgefs.Storage interface.
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

// GetCard implements the forgefs.Storage interface.
func (s *SQLiteStorage) GetCard(ctx context.Context, id string) (
	card *forgefs.Card, err error) {
	row := s.db.QueryRowContext(ctx, sqlCardJSON, id)
	var cardJSON string
	err = row.Scan(&cardJSON)
	if err != nil {
		return nil, err
	}
	var c forgefs.Card
	err = json.Unmarshal([]byte(cardJSON), &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

const sqlDecksCount string = `
    SELECT COUNT(*) FROM decks;
`

// GetDecksCount implements the forgefs.Storage interface.
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
    INSERT OR REPLACE INTO decks (
        id, name, expansion, sas, sas_version, aerc, a, e, r, c, f, d,
        house1, house2, house3, owned_by_me, funny, wish_list, json
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
`

// StoreDecks implements the forgefs.Storage interface.
func (s *SQLiteStorage) StoreDecks(
	ctx context.Context, decks []forgefs.Deck) error {
	for _, deck := range decks {
		j, err := json.Marshal(deck)
		if err != nil {
			return err
		}

		info := deck.DeckInfo
		var house1, house2, house3 string
		if len(info.Houses) == 3 {
			house1 = info.Houses[0].House
			house2 = info.Houses[1].House
			house3 = info.Houses[2].House
		}

		result, err := s.db.ExecContext(
			ctx, sqlDeckStore, info.KeyforgeID, info.Name, info.Expansion,
			info.SasRating, info.AercVersion, info.AercScore,
			info.AmberControl, info.ExpectedAmber, info.ArtifactControl,
			info.CreatureControl, info.Efficiency, info.Disruption,
			house1, house2, house3, deck.OwnedByMe, deck.Funny, deck.Wishlist,
			j)
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

// GetMyDeckNames implements the forgefs.Storage interface.
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

func normalizeExpansion(s string) string {
	switch strings.ToLower(s) {
	case "cota":
		return "CALL_OF_THE_ARCHONS"
	case "aoa":
		return "AGE_OF_ASCENSION"
	case "wc":
		return "WORLDS_COLLIDE"
	case "mm":
		return "MASS_MUTATION"
	case "dt":
		return "DARK_TIDINGS"
	}
	return s
}

func normalizeHouse(s string) string {
	switch strings.ToLower(s) {
	case "brobnar":
		return "Brobnar"
	case "dis":
		return "Dis"
	case "logos":
		return "Logos"
	case "mars":
		return "Mars"
	case "sanctum":
		return "Sanctum"
	case "saurian":
		return "Saurian"
	case "shadows":
		return "Shadows"
	case "staralliance", "sa":
		return "StarAlliance"
	case "unfathomable", "fish":
		return "Unfathomable"
	case "untamed":
		return "Untamed"
	}
	return s
}

func filterNodeToSQLConstraint(n *filter.Node) (string, error) {
	if n.Constraint != nil {
		var col string
		var normalizeString func(string) string
		switch n.Constraint.Var.(type) {
		case filter.AmberControl:
			col = "a"
		case filter.ExpectedAmber:
			col = "e"
		case filter.ArtifactControl:
			col = "r"
		case filter.CreatureControl:
			col = "c"
		case filter.Efficiency:
			col = "f"
		case filter.Disruption:
			col = "d"
		case filter.SAS:
			col = "sas"
		case filter.AERC:
			col = "aerc"
		case filter.Expansion:
			col = "expansion"
			normalizeString = normalizeExpansion
		case filter.House:
			col = "house"
			normalizeString = normalizeHouse
		default:
			return "", fmt.Errorf("unrecognized var type: %T", n.Constraint.Var)
		}

		var op string
		if n.Constraint.Op.Equal {
			op = "="
		} else {
			return "", fmt.Errorf("unrecognized op")
		}

		if n.Constraint.Value.Float != nil {
			return fmt.Sprintf(
				"%s %s %f", col, op, *n.Constraint.Value.Float), nil
		}
		if n.Constraint.Value.Int != nil {
			return fmt.Sprintf(
				"%s %s %d", col, op, *n.Constraint.Value.Int), nil
		}
		if n.Constraint.Value.String != nil {
			val := normalizeString(*n.Constraint.Value.String)
			if col == "house" {
				return fmt.Sprintf(
					"(house1 = \"%s\" OR house2 = \"%s\" OR house3 = \"%s\")",
					val, val, val), nil
			}
			return fmt.Sprintf("%s %s \"%s\"", col, op, val), nil
		}
		if len(n.Constraint.Value.Range) > 0 {
			min := n.Constraint.Value.MinString()
			max := n.Constraint.Value.MaxString()
			if min != "" && max != "" {
				return fmt.Sprintf(
					"(%s >= %s AND %s <= %s)", col, min, col, max), nil
			} else if min != "" {
				return fmt.Sprintf("%s >= %s", col, min), nil
			}
			return fmt.Sprintf("%s <= %s", col, max), nil
		}
		return "", fmt.Errorf("unrecognized value")
	}

	var boolOp string
	switch n.Op.(type) {
	case filter.And:
		boolOp = "AND"
	case filter.Or:
		boolOp = "OR"
	default:
		return "", fmt.Errorf("unrecognized bool op type: %T", n.Op)
	}

	left, err := filterNodeToSQLConstraint(n.Left)
	if err != nil {
		return "", err
	}
	right, err := filterNodeToSQLConstraint(n.Right)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("(%s %s %s)", left, boolOp, right), nil
}

const sqlMyDeckNamesFilterPrefix string = `
    SELECT id, name FROM decks
    WHERE owned_by_me = 1 AND
`

// GetMyDeckNamesWithFilter implements the forgefs.Storage interface.
func (s *SQLiteStorage) GetMyDeckNamesWithFilter(
	ctx context.Context, filterRoot *filter.Node) (
	names map[string]string, err error) {
	constraint, err := filterNodeToSQLConstraint(filterRoot)
	if err != nil {
		return nil, err
	}

	names = make(map[string]string)
	rows, err := s.db.QueryContext(ctx, sqlMyDeckNamesFilterPrefix+constraint)
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

// GetDeck implements the forgefs.Storage interface.
func (s *SQLiteStorage) GetDeck(ctx context.Context, id string) (
	deck *forgefs.Deck, err error) {
	row := s.db.QueryRowContext(ctx, sqlDeckJSON, id)
	var deckJSON string
	err = row.Scan(&deckJSON)
	if err != nil {
		return nil, err
	}
	var d forgefs.Deck
	err = json.Unmarshal([]byte(deckJSON), &d)
	if err != nil {
		return nil, err
	}
	return &d, nil
}
