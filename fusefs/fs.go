package fusefs

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/strib/forgefs"
	"github.com/strib/forgefs/filter"
	"github.com/strib/forgefs/fsutil"
)

const (
	numCardsInHouse = 12
)

// FSCard is a fuse inode representing a card's directory.
type FSCard struct {
	fs.Inode
	s  forgefs.Storage
	id string
	im *fsutil.ImageManager
}

var _ fs.InodeEmbedder = (*FSCard)(nil)
var _ fs.NodeLookuper = (*FSCard)(nil)
var _ fs.NodeReaddirer = (*FSCard)(nil)

// Lookup implements the fs.NodeLookuper interface.
func (c *FSCard) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (
	*fs.Inode, syscall.Errno) {
	n := c.GetChild(name)
	if n != nil {
		return n, 0
	}

	switch name {
	case fsutil.CardJSONFilename:
		card, err := c.s.GetCard(ctx, c.id)
		if err != nil {
			return nil, fs.ToErrno(err)
		}
		cardJSON, err := json.MarshalIndent(card, "", "\t")
		if err != nil {
			return nil, fs.ToErrno(err)
		}
		cardJSON = append(cardJSON, '\n')

		out.Size = uint64(len(cardJSON))
		n = c.NewInode(ctx, &fs.MemRegularFile{
			Data: cardJSON,
		}, fs.StableAttr{})
	default:
		imageURL, err := c.s.GetCardImageURL(ctx, c.id)
		if err != nil {
			return nil, fs.ToErrno(err)
		}
		if !strings.HasPrefix(name, fsutil.CardImagePrefix) {
			return nil, syscall.ENOENT
		}
		data, err := c.im.GetCardImage(ctx, c.id, imageURL)
		if err != nil {
			return nil, fs.ToErrno(err)
		}
		out.Size = uint64(len(data))
		n = c.NewInode(ctx, &fs.MemRegularFile{
			Data: data,
		}, fs.StableAttr{})
	}

	ok := c.AddChild(name, n, false)
	if !ok {
		return nil, syscall.EIO
	}
	return n, 0
}

// Readdir implements the fs.NodeReaddirer interface.
func (c *FSCard) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	imageURL, err := c.s.GetCardImageURL(ctx, c.id)
	if err != nil {
		return nil, fs.ToErrno(err)
	}
	split := strings.Split(imageURL, ".")
	suffix := ""
	if len(split) > 1 {
		suffix = split[len(split)-1]
	}

	entries := []fuse.DirEntry{
		{
			Name: fsutil.CardJSONFilename,
		},
		{
			Name: fsutil.CardImagePrefix + suffix,
		},
	}
	return fs.NewListDirStream(entries), 0
}

// FSCardsDir represents the directory containing all the card names
// as subdirectories.
type FSCardsDir struct {
	fs.Inode
	s  forgefs.Storage
	im *fsutil.ImageManager

	cards map[string]string
}

// NewFSCardsDir creates a new FSCardsDir instance.
func NewFSCardsDir(
	ctx context.Context, s forgefs.Storage, im *fsutil.ImageManager) (
	*FSCardsDir, error) {
	cd := &FSCardsDir{
		s:     s,
		im:    im,
		cards: make(map[string]string),
	}
	titles, err := cd.s.GetCardTitles(ctx)
	if err != nil {
		return nil, fs.ToErrno(err)
	}
	for id, title := range titles {
		cd.cards[title] = id
	}
	return cd, nil
}

var _ fs.InodeEmbedder = (*FSCardsDir)(nil)
var _ fs.NodeLookuper = (*FSCardsDir)(nil)
var _ fs.NodeReaddirer = (*FSCardsDir)(nil)

// Lookup implements the fs.NodeLookuper interface.
func (cd *FSCardsDir) Lookup(
	ctx context.Context, name string, out *fuse.EntryOut) (
	*fs.Inode, syscall.Errno) {
	n := cd.GetChild(name)
	if n != nil {
		return n, 0
	}

	id, ok := cd.cards[name]
	if !ok {
		return nil, syscall.ENOENT
	}

	n = cd.NewInode(ctx, &FSCard{
		s:  cd.s,
		id: id,
		im: cd.im,
	}, fs.StableAttr{
		Mode: syscall.S_IFDIR,
	})

	ok = cd.AddChild(name, n, false)
	if !ok {
		return nil, syscall.EIO
	}
	return n, 0
}

// Readdir implements the fs.NodeReaddirer interface.
func (cd *FSCardsDir) Readdir(ctx context.Context) (
	fs.DirStream, syscall.Errno) {
	entries := make([]fuse.DirEntry, 0, len(cd.cards))
	for title := range cd.cards {
		entries = append(entries, fuse.DirEntry{
			Mode: syscall.S_IFDIR,
			Name: title,
		})
	}

	return fs.NewListDirStream(entries), 0
}

// FSDeckHouseDir represents a directory containing symlinks to all
// the cards for one house in a deck.
type FSDeckHouseDir struct {
	fs.Inode

	d     *forgefs.Deck
	house string
}

var _ fs.InodeEmbedder = (*FSDeckHouseDir)(nil)
var _ fs.NodeLookuper = (*FSDeckHouseDir)(nil)
var _ fs.NodeReaddirer = (*FSDeckHouseDir)(nil)

// Lookup implements the fs.NodeLookuper interface.
func (dh *FSDeckHouseDir) Lookup(
	ctx context.Context, name string, out *fuse.EntryOut) (
	*fs.Inode, syscall.Errno) {
	n := dh.GetChild(name)
	if n != nil {
		return n, 0
	}

	i, err := strconv.Atoi(name)
	if err != nil {
		return nil, syscall.ENOENT
	}

	var house forgefs.HouseInDeck
	for _, h := range dh.d.DeckInfo.Houses {
		if h.House == dh.house {
			house = h
		}
	}
	if i > len(house.Cards) || i < 0 {
		return nil, syscall.ENOENT
	}

	path := dh.Path(nil)
	backtrack := strings.Repeat("../", strings.Count(path, "/")+1)
	n = dh.NewInode(ctx, &fs.MemSymlink{
		Data: []byte(backtrack + "cards/" + house.Cards[i-1].CardTitle),
	}, fs.StableAttr{
		Mode: syscall.S_IFLNK,
	})

	ok := dh.AddChild(name, n, false)
	if !ok {
		return nil, syscall.EIO
	}
	return n, 0
}

// Readdir implements the fs.NodeReaddirer interface.
func (dh *FSDeckHouseDir) Readdir(ctx context.Context) (
	fs.DirStream, syscall.Errno) {
	entries := make([]fuse.DirEntry, 0, numCardsInHouse)
	for i := 1; i <= numCardsInHouse; i++ {
		entries = append(entries, fuse.DirEntry{
			Mode: syscall.S_IFLNK,
			Name: fmt.Sprintf("%02d", i),
		})
	}

	return fs.NewListDirStream(entries), 0
}

// FSDeckCardsDir represents a directory containing subdirectories for
// each house in a deck.
type FSDeckCardsDir struct {
	fs.Inode

	d *forgefs.Deck
}

var _ fs.InodeEmbedder = (*FSDeckCardsDir)(nil)
var _ fs.NodeLookuper = (*FSDeckCardsDir)(nil)
var _ fs.NodeReaddirer = (*FSDeckCardsDir)(nil)

// Lookup implements the fs.NodeLookuper interface.
func (dcd *FSDeckCardsDir) Lookup(
	ctx context.Context, name string, out *fuse.EntryOut) (
	*fs.Inode, syscall.Errno) {
	n := dcd.GetChild(name)
	if n != nil {
		return n, 0
	}

	found := false
	for _, house := range dcd.d.DeckInfo.Houses {
		if house.House == name {
			found = true
			break
		}
	}
	if !found {
		return nil, syscall.ENOENT
	}

	n = dcd.NewInode(ctx, &FSDeckHouseDir{
		d:     dcd.d,
		house: name,
	}, fs.StableAttr{
		Mode: syscall.S_IFDIR,
	})

	ok := dcd.AddChild(name, n, false)
	if !ok {
		return nil, syscall.EIO
	}
	return n, 0
}

// Readdir implements the fs.NodeReaddirer interface.
func (dcd *FSDeckCardsDir) Readdir(ctx context.Context) (
	fs.DirStream, syscall.Errno) {
	entries := make([]fuse.DirEntry, 0, len(dcd.d.DeckInfo.Houses))
	for _, house := range dcd.d.DeckInfo.Houses {
		entries = append(entries, fuse.DirEntry{
			Mode: syscall.S_IFDIR,
			Name: house.House,
		})
	}

	return fs.NewListDirStream(entries), 0
}

// FSDeck represents the directory containing info about one deck.
type FSDeck struct {
	fs.Inode
	s  forgefs.Storage
	da forgefs.DataFetcher
	im *fsutil.ImageManager
	id string
}

var _ fs.InodeEmbedder = (*FSDeck)(nil)
var _ fs.NodeLookuper = (*FSDeck)(nil)
var _ fs.NodeReaddirer = (*FSDeck)(nil)

func (d *FSDeck) getDeck(ctx context.Context) (*forgefs.Deck, error) {
	deck, err := d.s.GetDeck(ctx, d.id)
	if err != nil {
		return nil, fs.ToErrno(err)
	}

	// Lookup the houses if we don't have them yet, and cache that
	// in the DB.
	if len(deck.DeckInfo.Houses) == 0 || deck.SASVersion == 0 {
		newDeck, err := d.da.GetDeck(ctx, d.id)
		if err != nil {
			return nil, fs.ToErrno(err)
		}
		deck.DeckInfo.Houses = newDeck.DeckInfo.Houses
		deck.SASVersion = newDeck.SASVersion
		err = d.s.StoreDecks(ctx, []forgefs.Deck{*deck})
		if err != nil {
			return nil, fs.ToErrno(err)
		}
	}

	return deck, nil
}

// Lookup implements the fs.NodeLookuper interface.
func (d *FSDeck) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (
	*fs.Inode, syscall.Errno) {
	n := d.GetChild(name)
	if n != nil {
		return n, 0
	}

	switch name {
	case fsutil.DeckJSONFilename:
		deck, err := d.getDeck(ctx)
		if err != nil {
			return nil, fs.ToErrno(err)
		}

		deckJSON, err := json.MarshalIndent(deck, "", "\t")
		if err != nil {
			return nil, fs.ToErrno(err)
		}
		deckJSON = append(deckJSON, '\n')

		out.Size = uint64(len(deckJSON))
		n = d.NewInode(ctx, &fs.MemRegularFile{
			Data: deckJSON,
		}, fs.StableAttr{})
	case fsutil.DeckImageFilename:
		deckImage, err := d.im.GetDeckImage(ctx, d.id)
		if err != nil {
			return nil, fs.ToErrno(err)
		}

		out.Size = uint64(len(deckImage))
		n = d.NewInode(ctx, &fs.MemRegularFile{
			Data: deckImage,
		}, fs.StableAttr{})
	case fsutil.DeckCardsDir:
		deck, err := d.getDeck(ctx)
		if err != nil {
			return nil, fs.ToErrno(err)
		}

		n = d.NewInode(ctx, &FSDeckCardsDir{
			d: deck,
		}, fs.StableAttr{
			Mode: syscall.S_IFDIR,
		})
	default:
		return nil, syscall.ENOENT
	}

	ok := d.AddChild(name, n, false)
	if !ok {
		return nil, syscall.EIO
	}
	return n, 0
}

// Readdir implements the fs.NodeReaddirer interface.
func (d *FSDeck) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	entries := []fuse.DirEntry{
		{
			Name: fsutil.DeckJSONFilename,
		},
		{
			Name: fsutil.DeckImageFilename,
		},
		{
			Name: fsutil.DeckCardsDir,
			Mode: syscall.S_IFDIR,
		},
	}
	return fs.NewListDirStream(entries), 0
}

// FSMyDecksDir represents a directory containing subdirectory for
// each deck of the user running the program, optionally filtered with
// constraints.
type FSMyDecksDir struct {
	fs.Inode
	s  forgefs.Storage
	da forgefs.DataFetcher
	im *fsutil.ImageManager

	decks      map[string]string
	filterRoot *filter.Node
}

// NewFSMyDecksDir creates a new unfiltered FSMyDecksDir instance.
func NewFSMyDecksDir(
	ctx context.Context, s forgefs.Storage, da forgefs.DataFetcher,
	im *fsutil.ImageManager) (*FSMyDecksDir, error) {
	mdd := &FSMyDecksDir{
		s:     s,
		da:    da,
		im:    im,
		decks: make(map[string]string),
	}
	names, err := mdd.s.GetMyDeckNames(ctx)
	if err != nil {
		return nil, fs.ToErrno(err)
	}
	for id, name := range names {
		mdd.decks[name] = id
	}
	return mdd, nil
}

// NewFSMyDecksDirWithFilter creates a new FSMyDecksDir instance, with
// the deck list filtered by the given filter.
func NewFSMyDecksDirWithFilter(
	ctx context.Context, s forgefs.Storage, da forgefs.DataFetcher,
	im *fsutil.ImageManager, filterRoot *filter.Node) (
	*FSMyDecksDir, error) {
	mdd := &FSMyDecksDir{
		s:          s,
		da:         da,
		im:         im,
		decks:      make(map[string]string),
		filterRoot: filterRoot,
	}
	names, err := mdd.s.GetMyDeckNamesWithFilter(ctx, filterRoot)
	if err != nil {
		return nil, fs.ToErrno(err)
	}
	for id, name := range names {
		mdd.decks[name] = id
	}
	return mdd, nil
}

var _ fs.InodeEmbedder = (*FSMyDecksDir)(nil)
var _ fs.NodeLookuper = (*FSMyDecksDir)(nil)
var _ fs.NodeReaddirer = (*FSMyDecksDir)(nil)

// Lookup implements the fs.NodeLookuper interface.
func (mdd *FSMyDecksDir) Lookup(
	ctx context.Context, name string, out *fuse.EntryOut) (
	*fs.Inode, syscall.Errno) {
	n := mdd.GetChild(name)
	if n != nil {
		return n, 0
	}

	id, ok := mdd.decks[name]
	if !ok {
		// See if it's a filter.
		filterRoot, err := filter.Parse(name)
		if err != nil {
			return nil, syscall.ENOENT
		}

		if mdd.filterRoot != nil {
			// AND this filter to this existing one.
			filterRoot = &filter.Node{
				Op:    filter.And{},
				Left:  filterRoot,
				Right: mdd.filterRoot,
			}
		}

		newMDD, err := NewFSMyDecksDirWithFilter(
			ctx, mdd.s, mdd.da, mdd.im, filterRoot)
		if err != nil {
			return nil, fs.ToErrno(err)
		}
		n = mdd.NewInode(ctx, newMDD, fs.StableAttr{
			Mode: syscall.S_IFDIR,
		})
	} else {
		n = mdd.NewInode(ctx, &FSDeck{
			s:  mdd.s,
			da: mdd.da,
			id: id,
			im: mdd.im,
		}, fs.StableAttr{
			Mode: syscall.S_IFDIR,
		})
	}

	ok = mdd.AddChild(name, n, false)
	if !ok {
		return nil, syscall.EIO
	}

	return n, 0
}

// Readdir implements the fs.NodeReaddirer interface.
func (mdd *FSMyDecksDir) Readdir(ctx context.Context) (
	fs.DirStream, syscall.Errno) {
	entries := make([]fuse.DirEntry, 0, len(mdd.decks))
	for name := range mdd.decks {
		entries = append(entries, fuse.DirEntry{
			Mode: syscall.S_IFDIR,
			Name: name,
		})
	}

	return fs.NewListDirStream(entries), 0
}

// FSRoot is the root of the file system.
type FSRoot struct {
	fs.Inode
	s  forgefs.Storage
	da forgefs.DataFetcher
	im *fsutil.ImageManager
}

// NewFSRoot creates a new `FSRoot` instance.
func NewFSRoot(
	s forgefs.Storage, da forgefs.DataFetcher,
	im *fsutil.ImageManager) *FSRoot {
	return &FSRoot{
		s:  s,
		da: da,
		im: im,
	}
}

var _ fs.InodeEmbedder = (*FSRoot)(nil)
var _ fs.NodeOnAdder = (*FSRoot)(nil)

func (r *FSRoot) getCardsDir(ctx context.Context) (*fs.Inode, error) {
	cd, err := NewFSCardsDir(ctx, r.s, r.im)
	if err != nil {
		return nil, err
	}
	cdNode := r.NewPersistentInode(ctx, cd, fs.StableAttr{
		Mode: syscall.S_IFDIR,
	})
	return cdNode, nil
}

func (r *FSRoot) getMyDecksDir(ctx context.Context) (*fs.Inode, error) {
	mdd, err := NewFSMyDecksDir(ctx, r.s, r.da, r.im)
	if err != nil {
		return nil, err
	}
	mddNode := r.NewPersistentInode(ctx, mdd, fs.StableAttr{
		Mode: syscall.S_IFDIR,
	})
	return mddNode, nil
}

// OnAdd implements the fs.NodeOnAdder interface.
func (r *FSRoot) OnAdd(ctx context.Context) {
	cdNode, err := r.getCardsDir(ctx)
	if err != nil {
		panic("Couldn't make cards dir")
	}
	ok := r.AddChild(fsutil.CardsDir, cdNode, false)
	if !ok {
		panic("Couldn't add cards dir")
	}

	mddNode, err := r.getMyDecksDir(ctx)
	if err != nil {
		panic("Couldn't make my-decks dir")
	}
	ok = r.AddChild(fsutil.MyDecksDir, mddNode, false)
	if !ok {
		panic("Couldn't add my-decks dir")
	}
}
