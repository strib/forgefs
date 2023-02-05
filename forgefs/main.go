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
	"os/signal"
	"path/filepath"
	"syscall"

	sdDaemon "github.com/coreos/go-systemd/daemon"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/strib/forgefs/fsutil"
	"github.com/strib/forgefs/fusefs"
	"github.com/strib/forgefs/net"
	"github.com/strib/forgefs/storage"
	"github.com/strib/forgefs/util"
)

const (
	defaultDoKAddr  = "https://decksofkeyforge.com"
	defaultSkyJAddr = "https://tts.skyj.io"
)

var defaultMountpoint = filepath.Join(os.Getenv("HOME"), "ffs")
var defaultConfigFile = filepath.Join(os.Getenv("HOME"), ".forgefs_config.json")
var defaultDBFile = filepath.Join(
	os.Getenv("HOME"), ".local", "share", "forgefs", "forgefs.sqlite")
var defaultImageCacheDir = filepath.Join(
	os.Getenv("HOME"), ".local", "share", "forgefs", "forgefs_images")

func sigHandler(signal os.Signal, server *fuse.Server) error {
	switch signal {
	case syscall.SIGTERM, syscall.SIGINT:
		fmt.Println("Unmounting")
		return server.Unmount()
	}
	return nil
}

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
	var showAPIKey = flag.Bool(
		"show-api-key", false, "Print the API key and exit")
	var showMountpoint = flag.Bool(
		"show-mountpoint", false, "Print the mountpoint and exit")
	var showImageCacheDir = flag.Bool(
		"show-image-cache-dir", false, "Print the image cache dir and exit")
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

	if showMountpoint != nil && *showMountpoint {
		fmt.Println(config.Mountpoint)
		return nil
	}

	if showAPIKey != nil && *showAPIKey {
		fmt.Println(config.DoKAPIKey)
		return nil
	}

	if showImageCacheDir != nil && *showImageCacheDir {
		fmt.Println(config.ImageCacheDir)
		return nil
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

	da := net.NewDoKAPI(config.DoKAddr, config.DoKAPIKey)
	err = util.CheckSASVersion(ctx, da, s)
	if err != nil {
		// Not a fatal error.
		fmt.Printf("Could not check SAS version: %+v\n", err)
	}

	count, err := s.GetCardsCount(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("Found %d cards\n", count)

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

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh)

	go func() {
		for {
			s := <-sigCh
			err := sigHandler(s, server)
			if err != nil {
				fmt.Printf("Couldn't handle signal: %+v", err)
			}
		}
	}()

	_, _ = sdDaemon.SdNotify(false /* unsetEnv */, "READY=1")

	server.Wait()
	return nil
}

func main() {
	err := doMain()
	if err != nil {
		log.Fatal(err)
	}
}
