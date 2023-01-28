package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/strib/forgefs"
)

var apiKey = flag.String("api-key", "", "Your decksofkeyforge API key")
var addr = flag.String(
	"addr", "https://decksofkeyforge.com", "The decksofkeyforge host address")
var dbFile = flag.String("db-file", ".forgefs.sqlite", "Local database file")
var mountpoint = flag.String("mountpoint", "ffs", "Mountpoint for forgefs")
var imageCacheDir = flag.String(
	"image-cache-dir", ".forgefs_images", "image cache directory")

func doMain() error {
	flag.Parse()
	if apiKey == nil || *apiKey == "" {
		return errors.New("No API key given")
	}
	if dbFile == nil || *dbFile == "" {
		return errors.New("No DB file given")
	}
	if mountpoint == nil || *mountpoint == "" {
		return errors.New("No DB file given")
	}

	ctx := context.Background()

	s, err := forgefs.NewSQLiteStorage(ctx, *dbFile)
	if err != nil {
		return err
	}
	defer s.Shutdown()

	count, err := s.GetCardsCount(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("Found %d cards\n", count)

	da := forgefs.NewDoKAPI(*addr, *apiKey)
	if count == 0 {
		cards, err := da.GetCards(ctx)
		if err != nil {
			return err
		}
		err = s.StoreCards(ctx, cards)
		if err != nil {
			return err
		}
	}

	count, err = s.GetDecksCount(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("Found %d decks\n", count)

	if count == 0 {
		decks, err := da.GetMyDecks(ctx)
		if err != nil {
			return err
		}
		err = s.StoreDecks(ctx, decks)
		if err != nil {
			return err
		}
	}

	im, err := forgefs.NewImageManager(*imageCacheDir)
	if err != nil {
		return err
	}

	root := forgefs.NewFSRoot(s, da, im)
	server, err := fs.Mount(*mountpoint, root, &fs.Options{
		MountOptions: fuse.MountOptions{
			Debug: true,
		},
	})
	if err != nil {
		return err
	}

	server.Wait()
	return nil
}

func main() {
	err := doMain()
	if err != nil {
		log.Fatal(err)
	}
}
