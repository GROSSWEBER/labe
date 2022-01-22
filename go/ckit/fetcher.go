package ckit

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/slub/labe/go/ckit/tabutils"
)

var (
	// ErrBlobNotFound can be used for unfetchable blobs.
	ErrBlobNotFound   = errors.New("blob not found")
	ErrBackendsFailed = errors.New("all backends failed")
	client            = http.Client{
		// We use the client to fetch data from backends. Often, we request one
		// item after another and there will be a 5 second timeout per request,
		// not for the whole operation.
		Timeout: 5 * time.Second,
	}
)

// Pinger allows to perform a simple health check.
type Pinger interface {
	Ping() error
}

// Fetcher fetches one or more blobs given their identifiers.
type Fetcher interface {
	Fetch(id string) ([]byte, error)
}

// SqliteFetcher serves index documents from sqlite database with a fixed schema,
// as generated by the makta tool.
type SqliteFetcher struct {
	DB *sqlx.DB
}

// Fetch document.
func (b *SqliteFetcher) Fetch(id string) (p []byte, err error) {
	var s string // TODO: could we just get into a []byte?
	if err := b.DB.Get(&s, "SELECT v FROM map WHERE k = ?", id); err != nil {
		return nil, err
	}
	return []byte(s), nil
}

// Ping pings the database.
func (b *SqliteFetcher) Ping() error {
	return b.DB.Ping()
}

// FetchGroup allows to run a index data fetch operation in a cascade over a
// couple of backends. The result from the first database that contains a value
// for a given id is returned. Currently sequential, but could be made
// parallel, maybe.
type FetchGroup struct {
	Backends []Fetcher
}

// FromFiles sets up a fetch group from a list of sqlite3 database filenames.
func (g *FetchGroup) FromFiles(files ...string) error {
	for _, f := range files {
		// TODO: In theory, we can allow empty files as well.
		if _, err := os.Stat(f); os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", f)
		}
		db, err := sqlx.Open("sqlite3", tabutils.WithReadOnly(f))
		if err != nil {
			return fmt.Errorf("database: %w", err)
		}
		fetcher := &SqliteFetcher{DB: db}
		g.Backends = append(g.Backends, fetcher)
	}
	return nil
}

// Ping is a healthcheck.
func (g *FetchGroup) Ping() error {
	for _, v := range g.Backends {
		w, ok := v.(Pinger)
		if !ok {
			continue
		}
		if err := w.Ping(); err != nil {
			return err
		}
	}
	return nil
}

// Fetch constructs a URL from a template and retrieves the blob.
func (g *FetchGroup) Fetch(id string) ([]byte, error) {
	for _, v := range g.Backends {
		if p, err := v.Fetch(id); err != nil {
			// OK to miss.
			continue
		} else {
			return p, nil
		}
	}
	return nil, ErrBackendsFailed
}
