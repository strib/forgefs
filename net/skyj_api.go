package net

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/strib/forgefs"
	"golang.org/x/time/rate"
)

type SkyJAPI struct {
	baseURL string
	limiter *rate.Limiter
}

var _ forgefs.DeckImageFetcher = (*SkyJAPI)(nil)

const (
	skyjCallsPerSec     = 1
	skyjBurst           = 5
	skyjDeckImageSuffix = ".jpg"
)

func NewSkyJAPI(baseURL string) *SkyJAPI {
	return &SkyJAPI{
		baseURL: baseURL,
		limiter: rate.NewLimiter(skyjCallsPerSec, skyjBurst),
	}
}

func (sja *SkyJAPI) GetDeckImageSuffix() string {
	return skyjDeckImageSuffix
}

func (sja *SkyJAPI) GetDeckImage(ctx context.Context, deckID string) (
	data []byte, err error) {
	err = sja.limiter.Wait(ctx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx, "GET", sja.baseURL+"/?type=deck-list&deckId="+deckID, nil)
	if err != nil {
		return nil, err
	}

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
	return body, nil
}
