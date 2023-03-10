package net

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/strib/forgefs"
	"golang.org/x/time/rate"
)

const (
	apiKeyHeader   = "Api-Key"
	dokCallsPerSec = 0.41 // 25 calls per minute for lowest patreon level
	dokBurst       = 1
)

// DoKAPI enables API calls to the decksofkeyforge server.  Calls are
// rate-limited to ensure compliance with the decksofkeyforge API
// rules.
type DoKAPI struct {
	baseURL string
	apiKey  string
	limiter *rate.Limiter
}

var _ forgefs.DataFetcher = (*DoKAPI)(nil)

// NewDoKAPI returns a new instance using the given address and API key.
func NewDoKAPI(addr, apiKey string) *DoKAPI {
	return &DoKAPI{
		baseURL: addr + "/public-api/",
		apiKey:  apiKey,
		limiter: rate.NewLimiter(dokCallsPerSec, dokBurst),
	}
}

func (da *DoKAPI) wait(ctx context.Context) error {
	return da.limiter.Wait(ctx)
}

// GetCards implements the forgefs.DataFetcher interface.
func (da *DoKAPI) GetCards(ctx context.Context) (
	cards []forgefs.Card, err error) {
	err = da.wait(ctx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx, "GET", da.baseURL+"v1/cards", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add(apiKeyHeader, da.apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		closeErr := resp.Body.Close()
		if err == nil {
			err = closeErr
		}
	}()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &cards)
	if err != nil {
		return nil, err
	}
	return cards, nil
}

// GetMyDecks implements the forgefs.DataFetcher interface.
func (da *DoKAPI) GetMyDecks(ctx context.Context) (
	decks []forgefs.Deck, err error) {
	err = da.wait(ctx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx, "GET", da.baseURL+"v1/my-decks", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add(apiKeyHeader, da.apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		closeErr := resp.Body.Close()
		if err == nil {
			err = closeErr
		}
	}()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &decks)
	if err != nil {
		return nil, err
	}
	return decks, nil
}

// GetDeck implements the forgefs.DataFetcher interface.
func (da *DoKAPI) GetDeck(ctx context.Context, id string, deck *forgefs.Deck) (
	updatedDeck forgefs.Deck, err error) {
	err = da.wait(ctx)
	if err != nil {
		return forgefs.Deck{}, err
	}

	req, err := http.NewRequestWithContext(
		ctx, "GET", da.baseURL+"v3/decks/"+id, nil)
	if err != nil {
		return forgefs.Deck{}, err
	}

	req.Header.Add(apiKeyHeader, da.apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return forgefs.Deck{}, err
	}
	defer func() {
		closeErr := resp.Body.Close()
		if err == nil {
			err = closeErr
		}
	}()
	if resp.StatusCode != 200 {
		return forgefs.Deck{}, fmt.Errorf("Error: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return forgefs.Deck{}, err
	}
	if deck != nil {
		updatedDeck = *deck
	}
	err = json.Unmarshal(body, &updatedDeck)
	if err != nil {
		return forgefs.Deck{}, err
	}
	return updatedDeck, nil
}
