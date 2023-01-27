package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/strib/forgefs"
)

var apiKey = flag.String("api-key", "", "Your decksofkeyforge API key")
var addr = flag.String(
	"addr", "https://decksofkeyforge.com", "The decksofkeyforge host address")

func main() {
	flag.Parse()
	if apiKey == nil || *apiKey == "" {
		log.Fatal("No API key given")
	}
	da := forgefs.NewDoKAPI(*addr, *apiKey)
	ctx := context.Background()
	cards, err := da.GetCards(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(cards)
}
