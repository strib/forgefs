package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/strib/forgefs"
)

var apiKey = flag.String("api-key", "", "Your decksofkeyforge API key")
var addr = flag.String(
	"addr", "https://decksofkeyforge.com", "The decksofkeyforge host address")
var dbFile = flag.String("db-file", ".forgefs.sqlite", "Local database file")
var mountpoint = flag.String("mountpoint", "ffs", "Mountpoint for forgefs")

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

	if count == 0 {
		da := forgefs.NewDoKAPI(*addr, *apiKey)
		cards, err := da.GetCards(ctx)
		if err != nil {
			return err
		}
		err = s.StoreCards(ctx, cards)
		if err != nil {
			return err
		}
	}

	root := forgefs.NewFSRoot(s)
	server, err := fs.Mount(*mountpoint, root, &fs.Options{})
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
