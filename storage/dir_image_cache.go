package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/strib/forgefs"
)

type DirImageCache struct {
	cacheDir string
}

var _ forgefs.ImageCache = (*DirImageCache)(nil)

func NewDirImageCache(cacheDir string) (*DirImageCache, error) {
	err := os.MkdirAll(cacheDir, 0755)
	if err != nil {
		return nil, err
	}
	return &DirImageCache{
		cacheDir: cacheDir,
	}, nil
}

func (dic *DirImageCache) getImage(
	ctx context.Context, prefix, fileType string) ([]byte, bool, error) {
	cacheFile := filepath.Join(dic.cacheDir, prefix+"."+fileType)

	f, err := os.Open(cacheFile)
	if err == nil {
		defer func() { _ = f.Close() }()
		data, err := io.ReadAll(f)
		if err != nil {
			return nil, false, err
		}
		return data, true, nil
	} else if os.IsNotExist(err) {
		return nil, false, nil
	}
	return nil, false, err
}

func (dic *DirImageCache) GetCardImage(
	ctx context.Context, cardID, fileType string) ([]byte, bool, error) {
	return dic.getImage(ctx, cardID, fileType)
}

func (dic *DirImageCache) storeImage(
	ctx context.Context, prefix, fileType string, data []byte) error {
	cacheFile := filepath.Join(dic.cacheDir, prefix+"."+fileType)
	f, err := os.OpenFile(cacheFile, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = f.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (dic *DirImageCache) StoreCardImage(
	ctx context.Context, cardID, fileType string, data []byte) error {
	return dic.storeImage(ctx, cardID, fileType, data)
}

func (dic *DirImageCache) GetDeckImage(
	ctx context.Context, deckID, fileType string) ([]byte, bool, error) {
	return dic.getImage(ctx, deckID, fileType)
}

func (dic *DirImageCache) StoreDeckImage(
	ctx context.Context, deckID, fileType string, data []byte) error {
	return dic.storeImage(ctx, deckID, fileType, data)
}
