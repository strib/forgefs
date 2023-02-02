package net

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/strib/forgefs"
)

type CardFetcher struct{}

var _ forgefs.CardImageFetcher = (*CardFetcher)(nil)

func (cf *CardFetcher) GetCardImage(ctx context.Context, imageURL string) (
	data []byte, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", imageURL, nil)
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
