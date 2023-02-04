package forgefs

import (
	"context"

	"github.com/strib/forgefs/filter"
)

// DataFetcher provides card and deck data.
type DataFetcher interface {
	// GetCards gets a full set of card objects.
	GetCards(ctx context.Context) ([]Card, error)
	// GetMyDecks gets all the decks associated with the user running
	// this program.
	GetMyDecks(ctx context.Context) (decks []Deck, err error)
	// GetDeck returns the full deck object for the given `id`. If
	// `deck` is not nil, it will be updated with whatever data has
	// changed, leaving the other existing fields intact.
	GetDeck(ctx context.Context, id string, deck *Deck) (
		updatedDeck Deck, err error)
}

// CardImageFetcher gets the front images for cards.
type CardImageFetcher interface {
	// GetCardImage gets the data for a card image, given a URL.
	GetCardImage(ctx context.Context, imageURL string) (
		data []byte, err error)
}

// DeckImageFetcher gets the front images for decks.
type DeckImageFetcher interface {
	// GetDeckImageSuffix gets the file suffix for the filetype used
	// by all the images returned from `GetDeckImage`.
	GetDeckImageSuffix() string
	// GetDeckImage gets the data for a deck image for the given `deckID`.
	GetDeckImage(ctx context.Context, deckID string) (
		data []byte, err error)
}

// Storage stores, fetches and filters card and deck data.
type Storage interface {
	// StoreCards stores all the given cards, overwriting any existing
	// cards with the same IDs as the new cards.
	StoreCards(ctx context.Context, cards []Card) error
	// GetCardTitles returns a map of cardID -> cardTitle for every
	// stored card.
	GetCardTitles(ctx context.Context) (titles map[string]string, err error)
	// GetCardImageURL retrieves the URL to the given card's image.
	GetCardImageURL(ctx context.Context, id string) (url string, err error)
	// GetCard gets the full card object for the given `id`.
	GetCard(ctx context.Context, id string) (card *Card, err error)
	// StoreDecks stores all the given decks, overwriting any existing
	// decks with the same IDs as the new decks.
	StoreDecks(ctx context.Context, decks []Deck) error
	// GetMyDeckNames gets all the decks owned by the user running the
	// program.  It returns a map of deckID -> deckName.
	GetMyDeckNames(ctx context.Context) (names map[string]string, err error)
	// GetMyDeckNamesWithFilter gets all the decks owned by the user,
	// which also match the given filter. It returns a map of deckID
	// -> deckName.
	GetMyDeckNamesWithFilter(ctx context.Context, filterRoot *filter.Node) (
		names map[string]string, err error)
	// GetDeck returns the full deck object for the given `id`.
	GetDeck(ctx context.Context, id string) (deck *Deck, err error)
	// GetSampleDeckWithVersion returns a sample deck and its SAS version.
	GetSampleDeckWithVersion(ctx context.Context) (
		deckID string, sasVersion int, err error)
	// Resets the storage, deleting all current data.
	Reset(ctx context.Context) error
}

// ImageCache stores card and deck images locally, for performance.
type ImageCache interface {
	// GetCardImage returns the card image data and `true` if the card
	// exists in the cache.
	GetCardImage(ctx context.Context, cardID, fileType string) (
		[]byte, bool, error)
	// StoreCardImage stores the given card data as the given file
	// type, overwriting any existing data for that card and type.
	StoreCardImage(
		ctx context.Context, cardID, fileType string, data []byte) error
	// GetDeckImage returns the deck image data and `true` if the deck
	// exists in the cache.
	GetDeckImage(ctx context.Context, deckID, fileType string) (
		[]byte, bool, error)
	// StoreDeckImage stores the given deck data as the given file
	// type, overwriting any existing data for that deck and type.
	StoreDeckImage(
		ctx context.Context, deckID, fileType string, data []byte) error
}
