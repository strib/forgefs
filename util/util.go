package util

import (
	"context"
	"fmt"

	"github.com/strib/forgefs"
)

// CheckSASVersion ensures the current SAS version isn't bigger than
// what we have stored; if it is, then reset the storage to force
// re-fetching of all the decks.
func CheckSASVersion(
	ctx context.Context, df forgefs.DataFetcher, s forgefs.Storage) error {
	deckID, v, err := s.GetSampleDeckWithVersion(ctx)
	if err != nil {
		return err
	}
	if deckID == "" {
		// No stored decks yet.
		return nil
	}

	deck, err := df.GetDeck(ctx, deckID, nil)
	if err != nil {
		return nil
	}

	if deck.SASVersion <= v {
		return nil
	}

	fmt.Printf(
		"Current SAS version is %d, compared to DB version %d; "+
			"resetting storage\n", deck.SASVersion, v)
	return s.Reset(ctx)
}
