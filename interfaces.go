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

type DataFetcher interface {
	GetCards(ctx context.Context) ([]Card, error)
}
