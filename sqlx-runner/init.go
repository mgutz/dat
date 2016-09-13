package runner

import (
	"database/sql"
	"regexp"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/mgutz/logxi/v1"
	"gopkg.in/mgutz/dat.v3/dat"
	"gopkg.in/mgutz/dat.v3/kvs"
	"gopkg.in/mgutz/dat.v3/postgres"
)

// Logger is the internal logger interface
var Logger log.Logger

// LogQueriesThreshold is the threshold for logging "slow" queries
var LogQueriesThreshold time.Duration

// LogErrNoRows if set tells the runner to log no row errors. Defaults to false.
var LogErrNoRows bool

// Cache caches query results.
var Cache kvs.KeyValueStore

// Statement separator for scripts having multiple statements.
var reScriptSeparator = regexp.MustCompile(`(?mis)^GO$`)

func init() {
	dat.Dialect = postgres.New()
	Logger = log.New("dat:sqlx")
}

// SetCache sets this runner's cache. The default cache is in-memory
// based. See cache.MemoryKeyValueStore.
func SetCache(store kvs.KeyValueStore) {
	Cache = store
}

// SetScriptSeparator sets the separator used when executing a script
func SetScriptSeparator(pattern string) {
	reScriptSeparator = regexp.MustCompile(pattern)
}

// MustPing pings a database with an exponential backoff. The
// function panics if the database cannot be pinged after 15 minutes
func MustPing(db *sql.DB) {
	var err error
	b := backoff.NewExponentialBackOff()
	ticker := backoff.NewTicker(b)

	// Ticks will continue to arrive when the previous operation is still running,
	// so operations that take a while to fail could run in quick succession.
	for range ticker.C {
		if err = db.Ping(); err != nil {
			Logger.Info("pinging database...", err.Error())
			continue
		}

		ticker.Stop()
		return
	}

	panic("Could not ping database!")
}
