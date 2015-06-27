package kvs

import (
	"time"

	"github.com/garyburd/redigo/redis"
)

func newRedisPool(host, password string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", host)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

////////////////////////////////////////

// NewDefaultRedisStore instantiates Redis.
func NewDefaultRedisStore() (KeyValueStore, error) {
	return NewRedisStore("", ":6379", "")
}

// NewRedisStore creates a new instance of RedisTokenStore.
func NewRedisStore(ns string, host string, password string) (*RedisStore, error) {
	logger.Info("Creating redis pool", "ns", ns, "host", host, "usingPassword", password == "")
	pool := newRedisPool(host, password)
	return &RedisStore{ns: ns + ":", pool: pool}, nil
}

// RedisStore is a concrete implementation of KeyValueStore for Redis.
type RedisStore struct {
	pool *redis.Pool
	ns   string
}

// Set sets a key's value with TTL. Use cache.TTLNever to never expire.
func (rs *RedisStore) Set(key, value string, ttl time.Duration) error {
	conn := rs.pool.Get()
	defer conn.Close()
	var err error

	key = rs.ns + key

	if ttl == TTLNever {
		_, err = conn.Do("SET", key, value)
	} else {
		_, err = conn.Do("SET", key, value, "PX", ttl.Nanoseconds()/NanosecondsPerMillisecond)
	}
	return err
}

// Get gets
func (rs *RedisStore) Get(key string) (string, error) {
	conn := rs.pool.Get()
	defer conn.Close()

	key = rs.ns + key
	s, err := redis.String(conn.Do("GET", key))
	if err == redis.ErrNil {
		return "", ErrNotFound
	} else if err != nil {
		return "", err
	}
	return s, nil
}

// Del deletes a key
func (rs *RedisStore) Del(key string) error {
	conn := rs.pool.Get()
	defer conn.Close()
	key = rs.ns + key
	_, err := conn.Do("DEL", key)
	return err
}

// FlushDB removes all keys.
func (rs *RedisStore) FlushDB() error {
	conn := rs.pool.Get()
	defer conn.Close()
	_, err := conn.Do("FLUSHDB")
	return err
}
