package runner

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/mgutz/logxi"
	"gopkg.in/mgutz/dat.v1"
	"gopkg.in/mgutz/dat.v1/kvs"
	"gopkg.in/mgutz/dat.v1/postgres"
)

var logger logxi.Logger

// LogQueriesThreshold is the threshold for logging "slow" queries
var LogQueriesThreshold time.Duration

// LogErrNoRows tells runner to log `sql.ErrNoRows`
var LogErrNoRows bool

func init() {
	dat.Dialect = postgres.New()
	logger = logxi.New("dat:sqlx")
}

// Cache caches query results.
var Cache kvs.KeyValueStore

// SetCache sets this runner's cache. The default cache is in-memory
// based. See cache.MemoryKeyValueStore.
func SetCache(store kvs.KeyValueStore) {
	Cache = store
}

// MustPing pings a database with an exponential backoff. The
// function panics if the database cannot be pinged after the specified duration.
// If no duration is specified it defaults to 15 minutes.
func MustPing(db *sql.DB, timeoutOrNil ...time.Duration) {
	var timeout time.Duration
	if len(timeoutOrNil) > 0 {
		timeout = timeoutOrNil[0]
	}

	err := pingDB(db, timeout)
	if err != nil {
		panic(err.Error())
	}
}

// ShouldPing pings a database with an exponential backoff. The
// function returns an error if the database cannot be pinged after the specified duration.
// If no duration is specified it defaults to 15 minutes.
func ShouldPing(db *sql.DB, timeoutOrNil ...time.Duration) error {
	var timeout time.Duration
	if len(timeoutOrNil) > 0 {
		timeout = timeoutOrNil[0]
	}

	return pingDB(db, timeout)
}

func pingDB(db *sql.DB, timeout time.Duration) error {
	var err error
	b := backoff.NewExponentialBackOff()
	if timeout > 0 {
		b.MaxElapsedTime = timeout
	}
	ticker := backoff.NewTicker(b)

	// Ticks will continue to arrive when the previous operation is still running,
	// so operations that take a while to fail could run in quick succession.
	for range ticker.C {
		if err = db.Ping(); err != nil {
			logger.Info("pinging database...", err.Error())
			continue
		}

		ticker.Stop()
		return nil
	}

	return fmt.Errorf("Could not ping database!")
}
