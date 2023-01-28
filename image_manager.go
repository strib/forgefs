package forgefs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func getImageFile(ctx context.Context, imageURL string) (
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

func getImageURLSuffix(imageURL string) string {
	split := strings.Split(imageURL, ".")
	suffix := ""
	if len(split) > 1 {
		suffix = split[len(split)-1]
	}
	return suffix
}

type ImageManager struct {
	cacheDir string
}

func NewImageManager(cacheDir string) (*ImageManager, error) {
	err := os.MkdirAll(cacheDir, 0755)
	if err != nil {
		return nil, err
	}
	return &ImageManager{
		cacheDir: cacheDir,
	}, nil
}

func (im *ImageManager) GetCardImage(
	ctx context.Context, cardID string, imageURL string) (
	[]byte, error) {
	suffix := getImageURLSuffix(imageURL)
	cacheFile := filepath.Join(im.cacheDir, cardID+"."+suffix)

	// Read from the cache if it exists.
	f, err := os.Open(cacheFile)
	if err == nil {
		defer func() { _ = f.Close() }()
		return io.ReadAll(f)
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	// If not, fetch from the URL.  (TODO: lock this to make sure only
	// one goroutine is fetching and writing the cache file for each
	// image at a time.)
	b, err := getImageFile(ctx, imageURL)
	if err != nil {
		return nil, err
	}

	f, err = os.OpenFile(cacheFile, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	_, err = f.Write(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}
