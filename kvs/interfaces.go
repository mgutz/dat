package kvs

import (
	"errors"
	"hash/fnv"
	"strconv"
	"time"
)

// KeyValueStore represents simple key value storage.
type KeyValueStore interface {
	Set(key, value string, ttl time.Duration) error
	Get(key string) (string, error)
	Del(key string) error
	FlushDB() error
}

// TTLNever means do not expire a key
const TTLNever time.Duration = -1

// NanosecondsPerMillisecond is used to convert between ns and ms.
const NanosecondsPerMillisecond = 1000000

// ErrNotFound is returned when an entry is not found in memory database.
var ErrNotFound = errors.New("Key not found")

// Hash returns the hash value of a string. The returned value is useful
// as a key.
func Hash(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	return strconv.FormatUint(uint64(h.Sum32()), 16)
}
