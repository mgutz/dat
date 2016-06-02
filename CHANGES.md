## v1.next

Timeouts per query. If a timeout occurs, then the query will be cancelled through
`pg_cancel_backend`

```go
err := DB.Select("SELECT pg_sleep(1)").Timeout(1 * time.Millisecond).Exec()
err ==  dat.ErrTimedout
```


## v1.1.0

*   [Caching](https://github.com/mgutz/dat#caching) - caching with Redis or (in-memory for testing)
*   [LogQueriesThreshold](https://github.com/mgutz/dat#tracing-sql) - log slow queries
*   dat.Null* creators
*   fix resource cleanup
*   fix duplicate error logging
*   include RFC339Nano in NullTime parsing
*   HUGE BUG in remapPlaceholders

## v1.0.0

*   Original dat moved to legacy branch.

*   Move to gopkg.in for API stability.

*   Legacy `Connection` renamed to `DB` to be consistent with `database/sql`

*   `Connection` is now the interface for `DB` and `Tx`. Use `Connection` to
    receive either a `DB` or `Tx`.

*   Support for nested transactions. **Needs user testing and feedback**.

    In a nested transaction *only* the top-most commit commits to the
    database if it has not been rollbacked. Any rollback in nested
    funtion results in entire transaction being rollbacked and leaves the transaction
    in `tx.IsRollbacked()` state.

``` go

    func nested(conn runner.Connection) error {
        tx, err := conn.Begin()
        if err != nil {
            return err
        }
        defer tx.AutoRollback()

        _, err := tx.SQL('...').Exec()
        if err != nil {
            return err
        }
        return tx.Commit()
    }

    func fromDB() error {
        return nested(DB)
    }

    func fromTx() error {
        tx, err := DB.Begin()
        if err != nil {
            return err
        }
        defer tx.AutoRollback()

        err := nested(tx)
        if err ! = nil {
            return logger.Error("Failed in nested", err)
        }
        // if Rollback was called, Commit returns an error
        return tx.Commit()
    }

```

*   `SelectDoc.HasMany` and `SelectDoc.HasOne` renamed to `Many` and `One` for
    retrieving hierarchical JSON documents. BTW, `SelectDoc` itself can be used
    in `Many` and `One` to build N-deep hierarchies.

```go

    DB.SelectDoc("id", "user_name", "avatar").
        Many("recent_comments", `SELECT id, title FROM comments WHERE id = users.id LIMIT 10`).
        Many("recent_posts", `SELECT id, title FROM posts WHERE author_id = users.id LIMIT 10`).
        One("account", `SELECT balance FROM accounts WHERE user_id = users.id`).
        From("users").
        Where("id = $1", 4).
        QueryStruct(&obj) // obj must be agreeable with json.Unmarshal()

```

*   Session obsoleted. Go's db library does not support transaction-less
    sessions. Use `Tx` if you need a session, otherwise use `DB` directly.

*   Fixes to dat.NullTime to properly work with JSON timestamps from HTTP
    handlers and `timestamptz`.

