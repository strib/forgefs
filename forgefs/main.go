package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/strib/forgefs"
)

var apiKey = flag.String("api-key", "", "Your decksofkeyforge API key")
var addr = flag.String(
	"addr", "https://decksofkeyforge.com", "The decksofkeyforge host address")
var dbFile = flag.String("db-file", ".forgefs.sqlite", "Local database file")

func doMain() error {
	flag.Parse()
	if apiKey == nil || *apiKey == "" {
		return errors.New("No API key given")
	}
	if dbFile == nil || *dbFile == "" {
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

	if count > 0 {
		return nil
	}

	da := forgefs.NewDoKAPI(*addr, *apiKey)
	cards, err := da.GetCards(ctx)
	if err != nil {
		return err
	}
	return s.StoreCards(ctx, cards)
}

func main() {
	err := doMain()
	if err != nil {
		log.Fatal(err)
	}
}
