package forgefs

import (
	"context"

	"github.com/strib/forgefs/filter"
)

type DataFetcher interface {
	GetCards(ctx context.Context) ([]Card, error)
	GetMyDecks(ctx context.Context) (decks []Deck, err error)
	GetDeck(ctx context.Context, id string) (deck Deck, err error)
}

type CardImageFetcher interface {
	GetCardImage(ctx context.Context, imageURL string) (
		data []byte, err error)
}

type DeckImageFetcher interface {
	GetDeckImageSuffix() string
	GetDeckImage(ctx context.Context, deckID string) (
		data []byte, err error)
}

type Storage interface {
	StoreCards(ctx context.Context, cards []Card) error
	GetCardTitles(ctx context.Context) (titles map[string]string, err error)
	GetCardImageURL(ctx context.Context, id string) (url string, err error)
	GetCard(ctx context.Context, id string) (card *Card, err error)
	StoreDecks(ctx context.Context, decks []Deck) error
	GetMyDeckNames(ctx context.Context) (names map[string]string, err error)
	GetMyDeckNamesWithFilter(ctx context.Context, filterRoot *filter.Node) (
		names map[string]string, err error)
	GetDeck(ctx context.Context, id string) (deck *Deck, err error)
}

type ImageCache interface {
	GetCardImage(ctx context.Context, cardID, fileType string) (
		[]byte, bool, error)
	StoreCardImage(
		ctx context.Context, cardID, fileType string, data []byte) error
	GetDeckImage(ctx context.Context, deckID, fileType string) (
		[]byte, bool, error)
	StoreDeckImage(
		ctx context.Context, deckID, fileType string, data []byte) error
}
