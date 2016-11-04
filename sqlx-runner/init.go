package runner

import (
	"time"

	"github.com/syreclabs/dat"
	"github.com/syreclabs/dat/kvs"
	"github.com/syreclabs/dat/postgres"
	"github.com/syreclabs/logxi/v1"
)

var logger log.Logger

// LogQueriesThreshold is the threshold for logging "slow" queries
var LogQueriesThreshold time.Duration

func init() {
	dat.Dialect = postgres.New()
	logger = log.New("dat:sqlx")
}

// Cache caches query results.
var Cache kvs.KeyValueStore

// SetCache sets this runner's cache. The default cache is in-memory
// based. See cache.MemoryKeyValueStore.
func SetCache(store kvs.KeyValueStore) {
	Cache = store
}
