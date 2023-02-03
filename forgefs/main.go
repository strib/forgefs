package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/strib/forgefs/fsutil"
	"github.com/strib/forgefs/fusefs"
	"github.com/strib/forgefs/net"
	"github.com/strib/forgefs/storage"
)

const (
	defaultDoKAddr       = "https://decksofkeyforge.com"
	defaultSkyJAddr      = "https://tts.skyj.io"
	defaultDBFile        = ".forgefs.sqlite"
	defaultMountpoint    = "ffs"
	defaultImageCacheDir = ".forgefs_images"
)

var defaultConfigFile = filepath.Join(os.Getenv("HOME"), ".forgefs_config.json")

func doMain() (err error) {
	// Start with built-in defaults.
	config := fusefs.Config{
		Debug:         false,
		DoKAddr:       defaultDoKAddr,
		SkyJAddr:      defaultSkyJAddr,
		DBFile:        defaultDBFile,
		Mountpoint:    defaultMountpoint,
		ImageCacheDir: defaultImageCacheDir,
	}

	// Load default config file, if it exists, to provide default
	// config values.
	configData, err := ioutil.ReadFile(defaultConfigFile)
	switch err {
	case nil:
		err = json.Unmarshal(configData, &config)
		if err != nil {
			return err
		}
	default:
		if !os.IsNotExist(err) {
			return err
		}
	}

	// Get flag values.  If present, these override the config file.
	flag.StringVar(
		&config.DoKAPIKey, "api-key", config.DoKAPIKey,
		"Your decksofkeyforge API key")
	flag.StringVar(
		&config.DoKAddr, "dok-addr", config.DoKAddr,
		"The decksofkeyforge API host address")
	flag.StringVar(
		&config.SkyJAddr, "skyj-addr", config.SkyJAddr,
		"The skyjedi API host address")
	flag.StringVar(
		&config.DBFile, "db-file", config.DBFile, "Local database file")
	flag.StringVar(
		&config.Mountpoint, "mountpoint", config.Mountpoint,
		"Mountpoint for forgefs")
	flag.StringVar(
		&config.ImageCacheDir, "image-cache-dir", config.ImageCacheDir,
		"image cache directory")
	var configFile = flag.String(
		"config-file", "",
		fmt.Sprintf("Custom config file location (default %s)",
			defaultConfigFile))
	flag.Parse()

	if configFile != nil && *configFile != "" {
		configData, err := ioutil.ReadFile(defaultConfigFile)
		if err != nil {
			return err
		}
		err = json.Unmarshal(configData, &config)
		if err != nil {
			return err
		}
	}
	if config.DoKAPIKey == "" {
		return errors.New("No API key given")
	}

	ctx := context.Background()

	s, err := storage.NewSQLiteStorage(ctx, config.DBFile)
	if err != nil {
		return err
	}
	defer func() {
		serr := s.Shutdown()
		if err == nil {
			err = serr
		}
	}()

	count, err := s.GetCardsCount(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("Found %d cards\n", count)

	da := net.NewDoKAPI(config.DoKAddr, config.DoKAPIKey)
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

	imageCache, err := storage.NewDirImageCache(config.ImageCacheDir)
	if err != nil {
		return err
	}
	cardFetcher := &net.CardFetcher{}
	deckFetcher := net.NewSkyJAPI(config.SkyJAddr)
	im := fsutil.NewImageManager(cardFetcher, deckFetcher, imageCache)

	fmt.Printf("Mounting at %s\n", config.Mountpoint)
	root := fusefs.NewFSRoot(s, da, im)
	server, err := fs.Mount(config.Mountpoint, root, &fs.Options{
		MountOptions: fuse.MountOptions{
			Debug: config.Debug,
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
