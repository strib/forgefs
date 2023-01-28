package forgefs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func getImageFile(ctx context.Context, imageURL string) (
	data []byte, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", imageURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		closeErr := resp.Body.Close()
		if err == nil {
			err = closeErr
		}
	}()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

type FSCard struct {
	fs.Inode
	s  *SQLiteStorage
	id string
}

var _ fs.InodeEmbedder = (*FSCard)(nil)
var _ fs.NodeLookuper = (*FSCard)(nil)
var _ fs.NodeReaddirer = (*FSCard)(nil)

func (c *FSCard) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (
	*fs.Inode, syscall.Errno) {
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
		return c.NewInode(ctx, &fs.MemRegularFile{
			Data: cardJSON,
		}, fs.StableAttr{}), 0
	default:
		imageURL, err := c.s.GetCardImageURL(ctx, c.id)
		if err != nil {
			return nil, fs.ToErrno(err)
		}
		if strings.HasPrefix(name, cardImagePrefix) {
			data, err := getImageFile(ctx, imageURL)
			if err != nil {
				return nil, fs.ToErrno(err)
			}
			out.Size = uint64(len(data))
			return c.NewInode(ctx, &fs.MemRegularFile{
				Data: data,
			}, fs.StableAttr{}), 0
		}

		return nil, syscall.ENOENT
	}
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

	// XXX: get image URL from DB, figure out suffix, and append to
	// the card image prefix for the name.
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

// FSRoot is the root of the file system.
type FSRoot struct {
	fs.Inode
	s *SQLiteStorage

	cards map[string]string
}

func NewFSRoot(s *SQLiteStorage) *FSRoot {
	return &FSRoot{
		s:     s,
		cards: make(map[string]string),
	}
}

var _ fs.InodeEmbedder = (*FSRoot)(nil)
var _ fs.NodeLookuper = (*FSRoot)(nil)
var _ fs.NodeReaddirer = (*FSRoot)(nil)

func (r *FSRoot) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (
	*fs.Inode, syscall.Errno) {
	id, ok := r.cards[name]
	if !ok {
		return nil, syscall.ENOENT
	}

	return r.NewInode(ctx, &FSCard{
		s:  r.s,
		id: id,
	}, fs.StableAttr{
		Mode: syscall.S_IFDIR,
	}), 0
}

func (r *FSRoot) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	titles, err := r.s.GetCardTitles(ctx)
	if err != nil {
		return nil, fs.ToErrno(err)
	}
	entries := make([]fuse.DirEntry, 0, len(titles))
	for id, title := range titles {
		entries = append(entries, fuse.DirEntry{
			Mode: syscall.S_IFDIR,
			Name: title,
		})
		r.cards[title] = id
	}

	return fs.NewListDirStream(entries), 0
}
