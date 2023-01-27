package forgefs

import (
	"context"
	"fmt"
	"io"

	"net/http"
)

const (
	apiKeyHeader = "Api-Key"
)

// DoKAPI enables API calls to the decksofkeyforge server.
type DoKAPI struct {
	baseURL string
	apiKey  string
}

// NewDoKAPI returns a new instance using the given address and API key.
func NewDoKAPI(addr, apiKey string) *DoKAPI {
	return &DoKAPI{
		baseURL: addr + "/public-api/",
		apiKey:  apiKey,
	}
}

func (da *DoKAPI) GetCards(ctx context.Context) (jsonString string, err error) {
	req, err := http.NewRequestWithContext(
		ctx, "GET", da.baseURL+"v1/cards", nil)
	if err != nil {
		return "", err
	}

	req.Header.Add(apiKeyHeader, da.apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		closeErr := resp.Body.Close()
		if err == nil {
			err = closeErr
		}
	}()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Error: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
