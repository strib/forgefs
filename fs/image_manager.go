package fs

import (
	"context"
	"strings"

	"github.com/strib/forgefs"
)

type ImageManager struct {
	cardFetcher forgefs.CardImageFetcher
	deckFetcher forgefs.DeckImageFetcher
	cache       forgefs.ImageCache
}

func NewImageManager(
	cardFetcher forgefs.CardImageFetcher, deckFetcher forgefs.DeckImageFetcher,
	cache forgefs.ImageCache) *ImageManager {
	return &ImageManager{
		cardFetcher: cardFetcher,
		deckFetcher: deckFetcher,
		cache:       cache,
	}
}

func getImageURLSuffix(imageURL string) string {
	split := strings.Split(imageURL, ".")
	suffix := ""
	if len(split) > 1 {
		suffix = split[len(split)-1]
	}
	return suffix
}

func (im *ImageManager) GetCardImage(
	ctx context.Context, cardID, imageURL string) ([]byte, error) {
	suffix := getImageURLSuffix(imageURL)
	data, ok, err := im.cache.GetCardImage(ctx, cardID, suffix)
	if err != nil {
		return nil, err
	}
	if ok {
		return data, nil
	}

	data, err = im.cardFetcher.GetCardImage(ctx, imageURL)
	if err != nil {
		return nil, err
	}

	err = im.cache.StoreCardImage(ctx, cardID, suffix, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (im *ImageManager) GetDeckImage(
	ctx context.Context, deckID string) ([]byte, error) {
	suffix := im.deckFetcher.GetDeckImageSuffix()
	data, ok, err := im.cache.GetDeckImage(ctx, deckID, suffix)
	if err != nil {
		return nil, err
	}
	if ok {
		return data, nil
	}

	data, err = im.deckFetcher.GetDeckImage(ctx, deckID)
	if err != nil {
		return nil, err
	}

	err = im.cache.StoreDeckImage(ctx, deckID, suffix, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
