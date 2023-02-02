package fs

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
)

const (
	numCardsInHouse = 12
)

type FSCard struct {
	fs.Inode
	s  forgefs.Storage
	id string
	im *ImageManager
}

var _ fs.InodeEmbedder = (*FSCard)(nil)
var _ fs.NodeLookuper = (*FSCard)(nil)
var _ fs.NodeReaddirer = (*FSCard)(nil)

func (c *FSCard) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (
	*fs.Inode, syscall.Errno) {
	n := c.GetChild(name)
	if n != nil {
		return n, 0
	}

	switch name {
	case cardJSONFilename:
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
		if !strings.HasPrefix(name, cardImagePrefix) {
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
			Name: cardJSONFilename,
		},
		{
			Name: cardImagePrefix + suffix,
		},
	}
	return fs.NewListDirStream(entries), 0
}

type FSCardsDir struct {
	fs.Inode
	s  forgefs.Storage
	im *ImageManager

	cards map[string]string
}

func NewFSCardsDir(
	ctx context.Context, s forgefs.Storage, im *ImageManager) (
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

type FSDeckHouseDir struct {
	fs.Inode

	d     *forgefs.Deck
	house string
}

var _ fs.InodeEmbedder = (*FSDeckHouseDir)(nil)
var _ fs.NodeLookuper = (*FSDeckHouseDir)(nil)
var _ fs.NodeReaddirer = (*FSDeckHouseDir)(nil)

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

func (mdd *FSDeckHouseDir) Readdir(ctx context.Context) (
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

type FSDeckCardsDir struct {
	fs.Inode

	d *forgefs.Deck
}

var _ fs.InodeEmbedder = (*FSDeckCardsDir)(nil)
var _ fs.NodeLookuper = (*FSDeckCardsDir)(nil)
var _ fs.NodeReaddirer = (*FSDeckCardsDir)(nil)

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

type FSDeck struct {
	fs.Inode
	s  forgefs.Storage
	da forgefs.DataFetcher
	im *ImageManager
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
	if len(deck.DeckInfo.Houses) == 0 {
		newDeck, err := d.da.GetDeck(ctx, d.id)
		if err != nil {
			return nil, fs.ToErrno(err)
		}
		deck.DeckInfo.Houses = newDeck.DeckInfo.Houses
		err = d.s.StoreDecks(ctx, []forgefs.Deck{*deck})
		if err != nil {
			return nil, fs.ToErrno(err)
		}
	}

	return deck, nil
}

func (d *FSDeck) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (
	*fs.Inode, syscall.Errno) {
	n := d.GetChild(name)
	if n != nil {
		return n, 0
	}

	switch name {
	case deckJSONFilename:
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
	case deckImageFilename:
		deckImage, err := d.im.GetDeckImage(ctx, d.id)
		if err != nil {
			return nil, fs.ToErrno(err)
		}

		out.Size = uint64(len(deckImage))
		n = d.NewInode(ctx, &fs.MemRegularFile{
			Data: deckImage,
		}, fs.StableAttr{})
	case deckCardsDir:
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

func (c *FSDeck) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	entries := []fuse.DirEntry{
		{
			Name: deckJSONFilename,
		},
		{
			Name: deckImageFilename,
		},
		{
			Name: deckCardsDir,
			Mode: syscall.S_IFDIR,
		},
	}
	return fs.NewListDirStream(entries), 0
}

type FSMyDecksDir struct {
	fs.Inode
	s  forgefs.Storage
	da forgefs.DataFetcher
	im *ImageManager

	decks      map[string]string
	filterRoot *filter.Node
}

func NewFSMyDecksDir(
	ctx context.Context, s forgefs.Storage, da forgefs.DataFetcher,
	im *ImageManager) (*FSMyDecksDir, error) {
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

func NewFSMyDecksDirWithFilter(
	ctx context.Context, s forgefs.Storage, da forgefs.DataFetcher,
	im *ImageManager, filterRoot *filter.Node) (
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
	im *ImageManager
}

func NewFSRoot(
	s forgefs.Storage, da forgefs.DataFetcher, im *ImageManager) *FSRoot {
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

func (r *FSRoot) OnAdd(ctx context.Context) {
	cdNode, err := r.getCardsDir(ctx)
	if err != nil {
		panic("Couldn't make cards dir")
	}
	ok := r.AddChild(cardsDir, cdNode, false)
	if !ok {
		panic("Couldn't add cards dir")
	}

	mddNode, err := r.getMyDecksDir(ctx)
	if err != nil {
		panic("Couldn't make my-decks dir")
	}
	ok = r.AddChild(myDecksDir, mddNode, false)
	if !ok {
		panic("Couldn't add my-decks dir")
	}
}
