package kvs

import (
	"time"

	gocache "github.com/pmylund/go-cache"
)

// MemoryKeyValueStore is an in-memory cache implementation of KeyValueStore.
type MemoryKeyValueStore struct {
	Cache           *gocache.Cache
	cleanupInterval time.Duration
}

// NewDefaultMemoryStore creates an instance of MemoryKeyValueStore
// with default settings.
func NewDefaultMemoryStore() KeyValueStore {
	return NewMemoryKeyValueStore(30 * time.Second)
}

// NewMemoryKeyValueStore creates an instance of MemoryKeyValueStore
// with given backing.
func NewMemoryKeyValueStore(cleanupInterval time.Duration) *MemoryKeyValueStore {
	cache := gocache.New(gocache.NoExpiration, cleanupInterval)
	store := &MemoryKeyValueStore{
		Cache:           cache,
		cleanupInterval: cleanupInterval,
	}
	return store
}

// Set sets a key with time-to-live.
func (store *MemoryKeyValueStore) Set(key, value string, ttl time.Duration) error {
	if ttl < store.cleanupInterval {
		logger.Warn("The cleanupInterval setting for in-memory key-value store is longer than the TTL of this operation, which means its effective TTL is based on the cleanupInterval")
	}
	store.Cache.Set(key, value, ttl)
	return nil
}

// Get retrieves a value given key.
func (store *MemoryKeyValueStore) Get(key string) (string, error) {
	val, found := store.Cache.Get(key)
	if !found {
		return "", nil
	}
	return val.(string), nil
}

// Del deletes value given key.
func (store *MemoryKeyValueStore) Del(key string) error {
	store.Cache.Delete(key)
	return nil
}

// FlushDB clears all keys
func (store *MemoryKeyValueStore) FlushDB() error {
	store.Cache = gocache.New(gocache.NoExpiration, store.cleanupInterval)
	return nil
}
