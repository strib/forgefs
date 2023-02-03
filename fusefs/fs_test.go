package fusefs

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/stretchr/testify/require"
	"github.com/strib/forgefs"
	"github.com/strib/forgefs/filter"
	"github.com/strib/forgefs/fsutil"
)

// Data fetcher.

type mockDataFetcher struct {
	cards   []forgefs.Card
	myDecks map[string]forgefs.Deck // id -> Deck
}

func (mdf *mockDataFetcher) GetCards(_ context.Context) ([]forgefs.Card, error) {
	return mdf.cards, nil
}

func (mdf *mockDataFetcher) GetMyDecks(
	_ context.Context) (decks []forgefs.Deck, err error) {
	decks = make([]forgefs.Deck, 0, len(mdf.myDecks))
	for _, d := range mdf.myDecks {
		decks = append(decks, d)
	}
	return decks, nil
}

func (mdf *mockDataFetcher) GetDeck(
	_ context.Context, id string) (deck forgefs.Deck, err error) {
	return mdf.myDecks[id], nil
}

// Card image fetcher.

type mockCardImageFetcher struct {
	cardImages map[string][]byte // url -> image
}

func (mcif *mockCardImageFetcher) GetCardImage(
	_ context.Context, imageURL string) (
	data []byte, err error) {
	return mcif.cardImages[imageURL], nil
}

// Deck image fetcher.

type mockDeckImageFetcher struct {
	suffix     string
	deckImages map[string][]byte // id -> image
}

func (mdif *mockDeckImageFetcher) GetDeckImageSuffix() string {
	return mdif.suffix
}

func (mdif *mockDeckImageFetcher) GetDeckImage(
	_ context.Context, deckID string) (
	data []byte, err error) {
	return mdif.deckImages[deckID], nil
}

// Storage.

type mockStorage struct {
	cards map[string]forgefs.Card // id -> Card
	decks map[string]forgefs.Deck // id -> Deck
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		cards: make(map[string]forgefs.Card),
		decks: make(map[string]forgefs.Deck),
	}
}

func (ms *mockStorage) StoreCards(
	_ context.Context, cards []forgefs.Card) error {
	for _, c := range cards {
		ms.cards[c.ID] = c
	}
	return nil
}

func (ms *mockStorage) GetCardTitles(_ context.Context) (
	titles map[string]string, err error) {
	titles = make(map[string]string, len(ms.cards))
	for id, c := range ms.cards {
		titles[id] = c.CardTitle
	}
	return titles, nil
}

func (ms *mockStorage) GetCardImageURL(_ context.Context, id string) (
	url string, err error) {
	return ms.cards[id].FrontImage, nil
}

func (ms *mockStorage) GetCard(_ context.Context, id string) (
	card *forgefs.Card, err error) {
	c := ms.cards[id]
	return &c, nil
}

func (ms *mockStorage) StoreDecks(
	_ context.Context, decks []forgefs.Deck) error {
	for _, d := range decks {
		ms.decks[d.DeckInfo.KeyforgeID] = d
	}
	return nil
}

func (ms *mockStorage) GetMyDeckNames(_ context.Context) (
	names map[string]string, err error) {
	names = make(map[string]string, len(ms.decks))
	for id, d := range ms.decks {
		if d.OwnedByMe {
			names[id] = d.DeckInfo.Name
		}
	}
	return names, nil
}

func mockFilter(n *filter.Node, d forgefs.Deck) (bool, error) {
	if n.Constraint != nil {
		var value float64

		switch n.Constraint.Var.(type) {
		case filter.AmberControl:
			value = d.DeckInfo.AmberControl
		case filter.ExpectedAmber:
			value = d.DeckInfo.ExpectedAmber
		default:
			return false, errors.New("Not implemented in the mock")
		}

		if n.Constraint.Value.Int != nil {
			return value == float64(*n.Constraint.Value.Int), nil
		}
		if len(n.Constraint.Value.Range) > 0 {
			min := n.Constraint.Value.MinString()
			if min != "" {
				minFloat, err := strconv.ParseFloat(min, 64)
				if err != nil {
					return false, err
				}
				if value < minFloat {
					return false, nil
				}
			}
			max := n.Constraint.Value.MaxString()
			if max != "" {
				maxFloat, err := strconv.ParseFloat(max, 64)
				if err != nil {
					return false, err
				}
				if value > maxFloat {
					return false, nil
				}
			}
			return true, nil
		}
		return false, errors.New("Value not implemented")
	}

	left, err := mockFilter(n.Left, d)
	if err != nil {
		return false, err
	}
	right, err := mockFilter(n.Right, d)
	if err != nil {
		return false, err
	}
	switch n.Op.(type) {
	case filter.And:
		return left && right, nil
	case filter.Or:
		return left || right, nil
	default:
		return false, errors.New("Unrecognized boolean op")
	}
}

func (ms *mockStorage) GetMyDeckNamesWithFilter(
	_ context.Context, filterRoot *filter.Node) (
	names map[string]string, err error) {
	names = make(map[string]string)
	for id, d := range ms.decks {
		if !d.OwnedByMe {
			continue
		}

		match, err := mockFilter(filterRoot, d)
		if err != nil {
			return nil, err
		}
		if match {
			names[id] = d.DeckInfo.Name
		}
	}
	return names, nil
}

func (ms *mockStorage) GetDeck(_ context.Context, id string) (
	deck *forgefs.Deck, err error) {
	d := ms.decks[id]
	return &d, nil
}

// Image cache.

type mockImageCache struct {
	cards map[string][]byte // id+suffix -> data
	decks map[string][]byte // id+suffix -> data
}

func newMockImageCache() *mockImageCache {
	return &mockImageCache{
		cards: make(map[string][]byte),
		decks: make(map[string][]byte),
	}
}

func (mic *mockImageCache) GetCardImage(
	_ context.Context, cardID, fileType string) ([]byte, bool, error) {
	data, ok := mic.cards[cardID+"."+fileType]
	if !ok {
		return nil, false, nil
	}
	return data, true, nil
}

func (mic *mockImageCache) StoreCardImage(
	_ context.Context, cardID, fileType string, data []byte) error {
	mic.cards[cardID+"."+fileType] = data
	return nil
}

func (mic *mockImageCache) GetDeckImage(
	_ context.Context, deckID, fileType string) ([]byte, bool, error) {
	data, ok := mic.decks[deckID+"."+fileType]
	if !ok {
		return nil, false, nil
	}
	return data, true, nil
}

func (mic *mockImageCache) StoreDeckImage(
	_ context.Context, deckID, fileType string, data []byte) error {
	mic.decks[deckID+"."+fileType] = data
	return nil
}

func readyMountTmpDir(t *testing.T) (
	mountpoint string, root *FSRoot, mdf *mockDataFetcher,
	mcif *mockCardImageFetcher, mdif *mockDeckImageFetcher, ms *mockStorage) {
	mountpoint, err := os.MkdirTemp(os.TempDir(), "forgefs-fs-test")
	require.NoError(t, err)
	t.Logf("Mountpoint: %s", mountpoint)
	t.Cleanup(func() { _ = os.RemoveAll(mountpoint) })

	mcif = &mockCardImageFetcher{}
	mdif = &mockDeckImageFetcher{suffix: "jpg"}
	im := fsutil.NewImageManager(mcif, mdif, newMockImageCache())
	ms = newMockStorage()
	mdf = &mockDataFetcher{}
	root = NewFSRoot(ms, mdf, im)

	return mountpoint, root, mdf, mcif, mdif, ms
}

func mountTmpDir(t *testing.T, mountpoint string, root *FSRoot) {
	server, err := fs.Mount(mountpoint, root, &fs.Options{
		MountOptions: fuse.MountOptions{
			Debug: false,
		},
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = server.Unmount()
		server.Wait()
	})
}

func makeCard(id, title, url string) forgefs.Card {
	return forgefs.Card{
		ID:         id,
		CardTitle:  title,
		FrontImage: url,
	}
}

func makeDeck(id, name string, mine bool, a, e float64) forgefs.Deck {
	return forgefs.Deck{
		DeckInfo: forgefs.DeckInfo{
			AmberControl:  a,
			ExpectedAmber: e,
			KeyforgeID:    id,
			Name:          name,
		},
		OwnedByMe: mine,
	}
}

func TestFSSimple(t *testing.T) {
	ctx := context.Background()
	mountpoint, root, mdf, mcif, mdif, ms := readyMountTmpDir(t)

	c1Title := "card1"
	c1URL := "card1.jpg"
	c1Image := []byte{1, 2, 3, 4}
	c1 := makeCard("1", c1Title, c1URL)

	c2Title := "card2"
	c2URL := "card2.jpg"
	c2Image := []byte{2, 3, 4, 1}
	c2 := makeCard("2", c2Title, c2URL)

	mdf.cards = []forgefs.Card{c1, c2}

	d1ID := "3"
	d1Name := "deck1"
	d1Image := []byte{3, 4, 1, 2}
	d1 := makeDeck(d1ID, d1Name, true, 10, 20)
	d2ID := "4"
	d2Name := "deck2"
	d2Image := []byte{4, 1, 2, 3}
	d2 := makeDeck(d2ID, d2Name, true, 3, 30)

	decks := []forgefs.Deck{d1, d2}
	mdf.myDecks = map[string]forgefs.Deck{
		d1ID: d1,
		d2ID: d2,
	}

	mcif.cardImages = map[string][]byte{
		c1URL: c1Image,
		c2URL: c2Image,
	}
	mdif.deckImages = map[string][]byte{
		d1ID: d1Image,
		d2ID: d2Image,
	}

	err := ms.StoreCards(ctx, mdf.cards)
	require.NoError(t, err)
	err = ms.StoreDecks(ctx, decks)
	require.NoError(t, err)

	// Add houses to a deck after it's first added to storage.
	h1, h2, h3 := "Ekwidon", "Logos", "Shadows"
	d1.DeckInfo.Houses = []forgefs.HouseInDeck{
		{
			House: h1,
			Cards: []forgefs.CardInDeck{
				{
					CardTitle: c1Title,
				},
				{
					CardTitle: c2Title,
				},
			},
		},
		{
			House: h2,
		},
		{
			House: h3,
		},
	}
	mdf.myDecks[d1ID] = d1

	mountTmpDir(t, mountpoint, root)

	// Check the root dir.
	checkDir := func(dir string, expectedNames []string) {
		entries, err := os.ReadDir(dir)
		require.NoError(t, err)
		require.Len(t, entries, len(expectedNames))
		names := make([]string, len(expectedNames))
		for i, e := range entries {
			names[i] = e.Name()
		}
		require.ElementsMatch(t, expectedNames, names)
	}
	checkDir(mountpoint, []string{fsutil.CardsDir, fsutil.MyDecksDir})

	// Check cards.
	cardsDir := filepath.Join(mountpoint, fsutil.CardsDir)
	checkDir(cardsDir, []string{c1Title, c2Title})

	// Check card dir.
	c1Dir := filepath.Join(cardsDir, c1Title)
	checkDir(c1Dir, []string{
		fsutil.CardImagePrefix + "jpg",
		fsutil.CardJSONFilename,
	})

	// Check card image.
	c1ImageFile := filepath.Join(c1Dir, fsutil.CardImagePrefix)
	checkFile := func(filename string, expectedData []byte) {
		f, err := os.Open(filename)
		require.NoError(t, err)
		defer func() { _ = f.Close() }()
		data, err := io.ReadAll(f)
		require.NoError(t, err)
		require.Equal(t, expectedData, data)
	}
	checkFile(c1ImageFile, c1Image)

	// Check card JSON.
	c1JSONFile := filepath.Join(c1Dir, fsutil.CardJSONFilename)
	require.NoError(t, err)
	c1JSON, err := json.MarshalIndent(c1, "", "\t")
	require.NoError(t, err)
	checkFile(c1JSONFile, append(c1JSON, '\n'))

	// Check my-decks dir.
	decksDir := filepath.Join(mountpoint, fsutil.MyDecksDir)
	checkDir(decksDir, []string{d1Name, d2Name})

	// Check deck dir.
	d1Dir := filepath.Join(decksDir, d1Name)
	checkDir(d1Dir, []string{
		fsutil.DeckJSONFilename,
		fsutil.DeckCardsDir,
		fsutil.DeckImageFilename,
	})

	// Check deck image.
	d1ImageFile := filepath.Join(d1Dir, fsutil.DeckImageFilename)
	checkFile(d1ImageFile, d1Image)

	// Check deck JSON.
	d1JSONFile := filepath.Join(d1Dir, fsutil.DeckJSONFilename)
	require.NoError(t, err)
	d1JSON, err := json.MarshalIndent(d1, "", "\t")
	require.NoError(t, err)
	checkFile(d1JSONFile, append(d1JSON, '\n'))

	// Check cards for the deck.
	d1CardsDir := filepath.Join(d1Dir, fsutil.DeckCardsDir)
	checkDir(d1CardsDir, []string{h1, h2, h3})
	h1CardsDir := filepath.Join(d1CardsDir, h1)
	checkDir(
		h1CardsDir, []string{
			"01", "02", "03", "04", "05", "06",
			"07", "08", "09", "10", "11", "12",
		})
	c1ViaDeckDir := filepath.Join(h1CardsDir, "01")
	checkDir(c1ViaDeckDir, []string{
		fsutil.CardImagePrefix + "jpg",
		fsutil.CardJSONFilename,
	})
	c1ViaDeckJSONFile := filepath.Join(c1ViaDeckDir, fsutil.CardJSONFilename)
	require.NoError(t, err)
	checkFile(c1ViaDeckJSONFile, append(c1JSON, '\n'))

	// Check filtered decks.
	// (a >= 5)
	filteredDecksDir := filepath.Join(decksDir, "a=5:")
	checkDir(filteredDecksDir, []string{d1Name})
	// (a >= 1) and (e <= 31)
	filteredDecksDir = filepath.Join(decksDir, "a=1:,e=:31")
	checkDir(filteredDecksDir, []string{d1Name, d2Name})
	filteredDecksDir = filepath.Join(decksDir, "a=1:+e=:31")
	checkDir(filteredDecksDir, []string{d1Name, d2Name})
	filteredDecksDir = filepath.Join(decksDir, filepath.Join("a=1:", "e=:31"))
	checkDir(filteredDecksDir, []string{d1Name, d2Name})
	// (a >= 30) or (e >= 25)
	filteredDecksDir = filepath.Join(decksDir, "a=30:^e=25:")
	checkDir(filteredDecksDir, []string{d2Name})
	// a >= 5 and (e < 25 or e > 35)
	filteredDecksDir = filepath.Join(decksDir, "a=5:,(e=:25^e=35:)")
	checkDir(filteredDecksDir, []string{d1Name})
}
