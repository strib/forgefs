package forgefs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/time/rate"
)

const (
	apiKeyHeader   = "Api-Key"
	dokCallsPerSec = 1
)

// DoKAPI enables API calls to the decksofkeyforge server.
type DoKAPI struct {
	baseURL string
	apiKey  string
	limiter *rate.Limiter
}

// NewDoKAPI returns a new instance using the given address and API key.
func NewDoKAPI(addr, apiKey string) *DoKAPI {
	return &DoKAPI{
		baseURL: addr + "/public-api/",
		apiKey:  apiKey,
		limiter: rate.NewLimiter(dokCallsPerSec, 5),
	}
}

func (da *DoKAPI) wait(ctx context.Context) error {
	return da.limiter.Wait(ctx)
}

func (da *DoKAPI) GetCards(ctx context.Context) (cards []Card, err error) {
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

func (da *DoKAPI) GetMyDecks(ctx context.Context) (decks []Deck, err error) {
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

func (da *DoKAPI) GetDeck(ctx context.Context, id string) (
	deck Deck, err error) {
	err = da.wait(ctx)
	if err != nil {
		return Deck{}, err
	}

	req, err := http.NewRequestWithContext(
		ctx, "GET", da.baseURL+"v3/decks/"+id, nil)
	if err != nil {
		return Deck{}, err
	}

	req.Header.Add(apiKeyHeader, da.apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Deck{}, err
	}
	defer func() {
		closeErr := resp.Body.Close()
		if err == nil {
			err = closeErr
		}
	}()
	if resp.StatusCode != 200 {
		return Deck{}, fmt.Errorf("Error: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Deck{}, err
	}
	err = json.Unmarshal(body, &deck)
	if err != nil {
		return Deck{}, err
	}
	return deck, nil
}
