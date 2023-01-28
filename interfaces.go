package forgefs

import (
	"context"
)

type CardNumber struct {
	CardNumber string `json:"cardNumber,omitempty"`
	Expansion  string `json:"expansion,omitempty"`
}

type ExpansionWins struct {
	Losses int `json:"losses,omitempty"`
	Wins   int `json:"wins,omitempty"`
}

type Trait struct {
	CardName         string   `json:"cardName,omitempty"`
	CardTraits       []string `json:"cardTraits,omitempty"`
	CardTraitsString string   `json:"cardTraitsString,omitempty"`
	CardTypes        []string `json:"cardTypes,omitempty"`
	CardTypesString  string   `json:"cardTypesString,omitempty"`
	House            string   `json:"house,omitempty"`
	ID               string   `json:"id,omitempty"`
	NotCardTraits    bool     `json:"notCardTraits,omitempty"`
	Player           string   `json:"player,omitempty"`
	PowersString     string   `json:"powersString,omitempty"`
	PrimaryGroup     bool     `json:"primaryGroup,omitempty"`
	Rating           int      `json:"rating,omitempty"`
	SynergyGroup     string   `json:"synergyGroup,omitempty"`
	SynergyGroupMax  int      `json:"synergyGroupMax,omitempty"`
	Trait            string   `json:"trait,omitempty"`
}

type ExtraCardInfo struct {
	Active                bool    `json:"active,omitempty"`
	AdaptiveScore         int     `json:"adaptiveScore,omitempty"`
	AmberControl          float64 `json:"amberControl,omitempty"`
	AmberControlMax       float64 `json:"amberControlMax,omitempty"`
	ArtifactControl       float64 `json:"artifactControl,omitempty"`
	ArtifactControlMax    float64 `json:"artifactControlMax,omitempty"`
	BaseSynPercent        float64 `json:"baseSynPercent,omitempty"`
	CardName              string  `json:"cardName,omitempty"`
	Created               string  `json:"created,omitempty"`
	CreatureControl       float64 `json:"creatureControl,omitempty"`
	CreatureControlMax    float64 `json:"creatureControlMax,omitempty"`
	CreatureProtection    float64 `json:"creatureProtection,omitempty"`
	CreatureProtectionMax float64 `json:"creatureProtectionMax,omitempty"`
	Disruption            float64 `json:"disruption,omitempty"`
	DisruptionMax         float64 `json:"disruptionMax,omitempty"`
	EffectivePower        float64 `json:"effectivePower,omitempty"`
	EffectivePowerMax     float64 `json:"effectivePowerMax,omitempty"`
	Efficiency            float64 `json:"efficiency,omitempty"`
	EfficiencyMax         float64 `json:"efficiencyMax,omitempty"`
	EnhancementAmber      int     `json:"enhancementAmber,omitempty"`
	EnhancementCapture    int     `json:"enhancementCapture,omitempty"`
	EnhancementDamage     int     `json:"enhancementDamage,omitempty"`
	EnhancementDraw       int     `json:"enhancementDraw,omitempty"`
	ExpectedAmber         float64 `json:"expectedAmber,omitempty"`
	ExpectedAmberMax      float64 `json:"expectedAmber,omitempty"`
	ID                    string  `json:"id,omitempty"`
	Other                 float64 `json:"other,omitempty"`
	OtherMax              float64 `json:"otherMax,omitempty"`
	Published             string  `json:"published,omitempty"`
	PublishedDate         string  `json:"publishedDate,omitempty"`
	Recursion             float64 `json:"recursion,omitempty"`
	RecursionMax          float64 `json:"recursionMax,omitempty"`
	Synergies             []Trait `json:"synergies,omitempty"`
	Traits                []Trait `json:"traits,omitempty"`
	Updated               string  `json:"updated,omitempty"`
	Version               int     `json:"version,omitempty"`
}

type Card struct {
	AERCScore        float64                  `json:"aercScore,omitempty"`
	AERCScoreAverage float64                  `json:"aercScoreAverage,omitempty"`
	AERCScoreMax     float64                  `json:"aercScoreMax,omitempty"`
	Amber            int                      `json:"amber,omitempty"`
	Anomaly          bool                     `json:"anomaly,omitempty"`
	Armor            int                      `json:"armor,omitempty"`
	ArmorString      string                   `json:"armorString,omitempty"`
	Big              bool                     `json:"big,omitempty"`
	CardNumber       string                   `json:"cardNumber,omitempty"`
	CardNumbers      []CardNumber             `json:"cardNumbers,omitempty"`
	CardText         string                   `json:"cardTest,omitempty"`
	CardTitle        string                   `json:"cardTitle,omitempty"`
	CardType         string                   `json:"cardType,omitempty"`
	Created          string                   `json:"created,omitempty"`
	EffectivePower   int                      `json:"effectivePower,omitempty"`
	Enhanced         bool                     `json:"enhanced,omitempty"`
	EvilTwin         bool                     `json:"evilTwin,omitempty"`
	Expansion        int                      `json:"expansion,omitempty"`
	ExpansionEnum    string                   `json:"expansionEnum,omitempty"`
	ExpansionWins    map[string]ExpansionWins `json:"expansionWins,omitempty"`
	ExtraCardInfo    ExtraCardInfo            `json:"extraCardInfo,omitempty"`
	FlavorText       string                   `json:"flavorText,omitempty"`
	FrontImage       string                   `json:"frontImage,omitempty"`
	House            string                   `json:"house,omitempty"`
	Houses           []string                 `json:",omitempty"`
	ID               string                   `json:"id,omitempty"`
	Losses           int                      `json:"losses,omitempty"`
	Maverick         bool                     `json:"maverick,omitempty"`
	Power            int                      `json:"power,omitempty"`
	PowerString      string                   `json:"power,omitempty"`
	Rarity           string                   `json:rarity",omitempty"`
	Traits           []string                 `json:"traits,omitempty"`
	Wins             int                      `json:"wins,omitempty"`
}

type CardInDeck struct {
	CardTitle string `json:"cardTitle,omitempty"`
	Rarity    string `json:"rarity,omitempty"`
	Legacy    bool   `json:"legacy,omitempty"`
	Maverick  bool   `json:"maverick,omitempty"`
	Anomaly   bool   `json:"anomaly,omitempty"`
}

type HouseInDeck struct {
	House string       `json:"house,omitempty"`
	Cards []CardInDeck `json:"cards,omitempty"`
}

type DeckInfo struct {
	ActionCount            int           `json:actionCount",omitempty"`
	AercScore              int           `json:"aercScore,omitempty"`
	AercVersion            int           `json:"aercVersion,omitempty"`
	AmberControl           float64       `json:"amberControl,omitempty"`
	AntisynergyRating      int           `json:"antisynergyRating,omitempty"`
	ArtifactControl        float64       `json:"artifactControl,omitempty"`
	CreatureControl        float64       `json:"creatureControl,omitempty"`
	CreatureCount          int           `json:"creatureCount,omitempty"`
	CreatureProtection     float64       `json:creatureProtection",omitempty"`
	DateAdded              string        `json:"dateAdded,omitempty"`
	Disruption             float64       `json:"disruption,omitempty"`
	EffectivePower         int           `json:"effectivePower,omitempty"`
	Efficiency             float64       `json:"efficiency,omitempty"`
	EfficiencyBonus        float64       `json:"efficiencyBonus,omitempty"`
	Expansion              string        `json:"expansion,omitempty"`
	ExpectedAmber          float64       `json:"expectedAmber,omitempty"`
	Houses                 []HouseInDeck `json:"housesAndCards,omitempty"`
	ID                     int           `json:"id,omitempty"`
	KeyforgeID             string        `json:"keyforgeId,omitempty"`
	LastSasUpdate          string        `json:"lastSasUpdate,omitempty"`
	Name                   string        `json:"name,omitempty"`
	PreviousMajorSasRating int           `json:"previousMajorSasRating,omitempty"`
	PreviousSasRating      int           `json:"previosSasRating,omitempty"`
	RawAmber               int           `json:"rawAmber,omitempty"`
	SasPercentile          float64       `json:sasPercentile",omitempty"`
	SasRating              int           `json:"sasRating,omitempty"`
	SynergyRating          int           `json:"synergyRating,omitempty"`
	TotalArmor             int           `json:"totalArmor,omitempty"`
	TotalPower             int           `json:"totalPower,omitempty"`
}

type Deck struct {
	DeckInfo  DeckInfo `json:"deck,omitempty"`
	Funny     bool     `json:"funny,omitempty"`
	Notes     string   `json:"notes,omitempty"`
	OwnedByMe bool     `json:"ownedByMe,omitempty"`
	Wishlist  bool     `json:"wishlist,omitempty"`
}

type DataFetcher interface {
	GetCards(ctx context.Context) ([]Card, error)
	GetDecks(ctx context.Context) ([]Deck, error)
}
