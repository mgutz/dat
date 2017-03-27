# dat

[GoDoc](https://godoc.org/github.com/mgutz/dat)

`dat` (Data Access Toolkit) is a fast, lightweight Postgres library for Go.

*   Focused on Postgres. See `Insect`, `Upsert`, `SelectDoc`, `QueryJSON`

*   Built on a solid foundation [sqlx](https://github.com/jmoiron/sqlx)

    ```go
    // child DB is *sqlx.DB
    DB.DB.Queryx(`SELECT * FROM users`)
    ```

*   SQL and backtick friendly

    ```go
    DB.SQL(`SELECT * FROM people LIMIT 10`).QueryStructs(&people)
    ```

*   JSON Document retrieval (single trip to Postgres, requires Postgres 9.3+)

    ```go
    DB.SelectDoc("id", "user_name", "avatar").
        Many("recent_comments", `SELECT id, title FROM comments WHERE id = users.id LIMIT 10`).
        Many("recent_posts", `SELECT id, title FROM posts WHERE author_id = users.id LIMIT 10`).
        One("account", `SELECT balance FROM accounts WHERE user_id = users.id`).
        From("users").
        Where("id = $1", 4).
        QueryStruct(&obj) // obj must be agreeable with json.Unmarshal()
    ```

    results in

    ```json
    {
        "id": 4,
        "user_name": "mario",
        "avatar": "https://imgur.com/a23x.jpg",
        "recent_comments": [{"id": 1, "title": "..."}],
        "recent_posts": [{"id": 1, "title": "..."}],
        "account": {
            "balance": 42.00
        }
    }
    ```

*   JSON marshalable bytes (requires Postgres 9.3+)

    ```go
    var b []byte
    b, _ = DB.SQL(`SELECT id, user_name, created_at FROM users WHERE user_name = $1 `,
        "mario",
    ).QueryJSON()

    // straight into map
    var obj map[string]interface{}
    DB.SQL(`SELECT id, user_name, created_at FROM users WHERE user_name = $1 `,
        "mario",
    ).QueryObject(&obj)
    ```

*   Ordinal placeholders

    ```go
    DB.SQL(`SELECT * FROM people WHERE state = $1`, "CA").Exec()
    ```

*   SQL-like API

    ```go
    err := DB.
        Select("id, user_name").
        From("users").
        Where("id = $1", id).
        QueryStruct(&user)
    ```

*   Redis caching

    ```go
    // cache result for 30 seconds
    key := "user:" + strconv.Itoa(user.id)
    err := DB.
        Select("id, user_name").
        From("users").
        Where("id = $1", user.id).
        Cache(key, 30 * time.Second, false).
        QueryStruct(&user)
    ```

*   Nested transactions

*   Per query timeout with database cancellation logic `pg_cancel_backend`

*   SQL and slow query logging

*   Performant

    -   ordinal placeholder logic is optimized to be nearly as fast as using `?`
    -   `dat` can interpolate queries locally resulting in performance increase
        over plain database/sql and sqlx. [Benchmarks](https://github.com/mgutz/dat/wiki/Benchmarks)

## Getting Started

Get it

dat.v1 uses [glide](https://github.com/Masterminds/glide) package dependency manager. 
Earlier builds relied on gopkg.in which at the time was as good a solution as any.
Will move to `dep` once it is stable.

```sh
glide get gopkg.in/mgutz/dat.v1/sqlx-runner
```

Use it

```go
import (
    "database/sql"

    _ "github.com/lib/pq"
    "gopkg.in/mgutz/dat.v1"
    "gopkg.in/mgutz/dat.v1/sqlx-runner"
)

// global database (pooling provided by SQL driver)
var DB *runner.DB

func init() {
    // create a normal database connection through database/sql
    db, err := sql.Open("postgres", "dbname=dat_test user=dat password=!test host=localhost sslmode=disable")
    if err != nil {
        panic(err)
    }

    // ensures the database can be pinged with an exponential backoff (15 min)
    runner.MustPing(db)

    // set to reasonable values for production
    db.SetMaxIdleConns(4)
    db.SetMaxOpenConns(16)

    // set this to enable interpolation
    dat.EnableInterpolation = true

    // set to check things like sessions closing.
    // Should be disabled in production/release builds.
    dat.Strict = false

    // Log any query over 10ms as warnings. (optional)
    runner.LogQueriesThreshold = 10 * time.Millisecond

    DB = runner.NewDB(db, "postgres")
}

type Post struct {
    ID        int64         `db:"id"`
    Title     string        `db:"title"`
    Body      string        `db:"body"`
    UserID    int64         `db:"user_id"`
    State     string        `db:"state"`
    UpdatedAt dat.NullTime  `db:"updated_at"`
    CreatedAt dat.NullTime  `db:"created_at"`
}

func main() {
    var post Post
    err := DB.
        Select("id, title").
        From("posts").
        Where("id = $1", 13).
        QueryStruct(&post)
    fmt.Println("Title", post.Title)
}
```

## Feature highlights

### Use Builders or SQL

Query Builder

```go
var posts []*Post
err := DB.
    Select("title", "body").
    From("posts").
    Where("created_at > $1", someTime).
    OrderBy("id ASC").
    Limit(10).
    QueryStructs(&posts)
```

Plain SQL

```go
err = DB.SQL(`
    SELECT title, body
    FROM posts WHERE created_at > $1
    ORDER BY id ASC LIMIT 10`,
    someTime,
).QueryStructs(&posts)
```

Note: `dat` does not trim the SQL string, thus any extra whitespace is
transmitted to the database.

In practice, SQL is easier to write with backticks. Indeed, the reason this
library exists is most SQL builders introduce a DSL to insulate the user
from SQL.

Query builders shine when dealing with data transfer objects, structs.

### Fetch Data Simply

Query then scan result to struct(s)

```go
var post Post
err := DB.
    Select("id, title, body").
    From("posts").
    Where("id = $1", id).
    QueryStruct(&post)

var posts []*Post
err = DB.
    Select("id, title, body").
    From("posts").
    Where("id > $1", 100).
    QueryStructs(&posts)
```

Query scalar values or a slice of values

```go
var n int64
DB.SQL("SELECT count(*) FROM posts WHERE title=$1", title).QueryScalar(&n)

var ids []int64
DB.SQL("SELECT id FROM posts", title).QuerySlice(&ids)
```

### Field Mapping

**dat** DOES NOT map fields automatically like sqlx.
You must explicitly set `db` struct tags in your types.

Embedded fields are mapped breadth-first.

```go
type Realm struct {
    RealmUUID string `db:"realm_uuid"`
}
type Group struct {
    GroupUUID string `db:"group_uuid"`
    *Realm
}

g := &Group{Realm: &Realm{"11"}, GroupUUID: "22"}

sql, args := InsertInto("groups").Columns("group_uuid", "realm_uuid").Record(g).ToSQL()
expected := `
    INSERT INTO groups ("group_uuid", "realm_uuid")
    VALUES ($1, $2)
	`
```

### Blacklist and Whitelist

Control which columns get inserted or updated when processing external data

```go
// userData came in from http.Handler, prevent them from setting protected fields
DB.InsertInto("payments").
    Blacklist("id", "updated_at", "created_at").
    Record(userData).
    Returning("id").
    QueryScalar(&userData.ID)

// ensure session user can only update his information
DB.Update("users").
    SetWhitelist(user, "user_name", "avatar", "quote").
    Where("id = $1", session.UserID).
    Exec()
```

### IN queries

__applicable when dat.EnableInterpolation == true__

Simpler IN queries which expand correctly

```go
ids := []int64{10,20,30,40,50}
b := DB.SQL("SELECT * FROM posts WHERE id IN $1", ids)
b.MustInterpolate() == "SELECT * FROM posts WHERE id IN (10,20,30,40,50)"
```

### Tracing SQL

`dat` uses [logxi](https://github.com/mgutz/logxi) for logging. By default,
*logxi* logs all warnings and errors to the console. `dat` logs the
SQL and its arguments on any error. In addition, `dat` logs slow queries
as warnings if `runner.LogQueriesThreshold > 0`

To trace all SQL, set environment variable

```sh
LOGXI=dat* yourapp
```

## CRUD

### Create

Use `Returning` and `QueryStruct` to insert and update struct fields in one
trip

```go
var post Post

err := DB.
    InsertInto("posts").
    Columns("title", "state").
    Values("My Post", "draft").
    Returning("id", "created_at", "updated_at").
    QueryStruct(&post)
```

Use `Blacklist` and `Whitelist` to control which record (input struct) fields
are inserted.

```go
post := Post{Title: "Go is awesome", State: "open"}
err := DB.
    InsertInto("posts").
    Blacklist("id", "user_id", "created_at", "updated_at").
    Record(&post).
    Returning("id", "created_at", "updated_at").
    QueryStruct(&post)

// use wildcard to include all columns
err := DB.
    InsertInto("posts").
    Whitelist("*").
    Record(&post).
    Returning("id", "created_at", "updated_at").
    QueryStruct(&post)

```

Insert Multiple Records

```go
// create builder
b := DB.InsertInto("posts").Columns("title")

// add some new posts
for i := 0; i < 3; i++ {
    b.Record(&Post{Title: fmt.Sprintf("Article %s", i)})
}

// OR (this is more efficient as it does not do any reflection)
for i := 0; i < 3; i++ {
    b.Values(fmt.Sprintf("Article %s", i))
}

// execute statement
_, err := b.Exec()
```

Inserts if not exists or select in one-trip to database

```go
sql, args := DB.
    Insect("tab").
    Columns("b", "c").
    Values(1, 2).
    Where("d = $1", 3).
    Returning("id", "f", "g").
    ToSQL()

sql == `
WITH
    sel AS (SELECT id, f, g FROM tab WHERE (d = $1)),
    ins AS (
        INSERT INTO "tab"("b","c")
        SELECT $2,$3
        WHERE NOT EXISTS (SELECT 1 FROM sel)
        RETURNING "id","f","g"
    )
SELECT * FROM ins UNION ALL SELECT * FROM sel
`
```

### Read

```go
var other Post

err = DB.
    Select("id, title").
    From("posts").
    Where("id = $1", post.ID).
    QueryStruct(&other)

published := `
    WHERE user_id = $1
        AND state = 'published'
`

var posts []*Post
err = DB.
    Select("id, title").
    From("posts").
    Scope(published, 100).
    QueryStructs(&posts)
```

### Update

Use `Returning` to fetch columns updated by triggers. For example,
an update trigger on "updated\_at" column

```go
err = DB.
    Update("posts").
    Set("title", "My New Title").
    Set("body", "markdown text here").
    Where("id = $1", post.ID).
    Returning("updated_at").
    QueryScalar(&post.UpdatedAt)
```

Upsert - Update or Insert

```go
sql, args := DB.
    Upsert("tab").
    Columns("b", "c").
    Values(1, 2).
    Where("d=$1", 4).
    Returning("f", "g").
    ToSQL()

expected := `
WITH
    upd AS (
        UPDATE tab
        SET "b" = $1, "c" = $2
        WHERE (d=$3)
        RETURNING "f","g"
    ), ins AS (
        INSERT INTO "tab"("b","c")
        SELECT $1,$2
        WHERE NOT EXISTS (SELECT 1 FROM upd)
        RETURNING "f","g"
    )
SELECT * FROM ins UNION ALL SELECT * FROM upd
`
```

__applicable when dat.EnableInterpolation == true__

To reset columns to their default DDL value, use `DEFAULT`. For example,
to reset `payment\_type`

```go
res, err := DB.
    Update("payments").
    Set("payment_type", dat.DEFAULT).
    Where("id = $1", 1).
    Exec()
```

Use `SetBlacklist` and `SetWhitelist` to control which fields are updated.

```go
// create blacklists for each of your structs
blacklist := []string{"id", "created_at"}
p := paymentStructFromHandler

err := DB.
    Update("payments").
    SetBlacklist(p, blacklist...)
    Where("id = $1", p.ID).
    Exec()
```

Use a map of attributes

``` go
attrsMap := map[string]interface{}{"name": "Gopher", "language": "Go"}
result, err := DB.
    Update("developers").
    SetMap(attrsMap).
    Where("language = $1", "Ruby").
    Exec()
```

### Delete

``` go
result, err = DB.
    DeleteFrom("posts").
    Where("id = $1", otherPost.ID).
    Exec()
```

### Joins

Define JOINs in argument to `From`

``` go
err = DB.
    Select("u.*, p.*").
    From(`
        users u
        INNER JOIN posts p on (p.author_id = u.id)
    `).
    WHERE("p.state = 'published'").
    QueryStructs(&liveAuthors)
```

#### Scopes

Scopes predefine JOIN and WHERE conditions.
Scopes may be used with `DeleteFrom`, `Select` and `Update`.

As an example, a "published" scope might define published posts
by user.

```go
publishedPosts := `
    INNER JOIN users u on (p.author_id = u.id)
    WHERE
        p.state == 'published' AND
        p.deleted_at IS NULL AND
        u.user_name = $1
`

unpublishedPosts := `
    INNER JOIN users u on (p.author_id = u.id)
    WHERE
        p.state != 'published' AND
        p.deleted_at IS NULL AND
        u.user_name = $1
`

err = DB.
    Select("p.*").                      // must qualify columns
    From("posts p").
    Scope(publishedPosts, "mgutz").
    QueryStructs(&posts)
```

## Creating Connections

All queries are made in the context of a connection which is acquired
from the underlying SQL driver's pool

For one-off operations, use `DB` directly

```go
err := DB.SQL(sql).QueryStruct(&post)
```

For multiple operations, create a `Tx` transaction.
__`defer Tx.AutoCommit()` or `defer Tx.AutoRollback()` MUST be called__

```go
func PostsIndex(rw http.ResponseWriter, r *http.Request) {
    tx, _ := DB.Begin()
    defer tx.AutoRollback()

    // Do queries with the session
    var post Post
    err := tx.Select("id, title").
        From("posts").
        Where("id = $1", post.ID).
        QueryStruct(&post)
    )
    if err != nil {
        // `defer AutoRollback()` is used, no need to rollback on error
        r.WriteHeader(500)
        return
    }

    // do more queries with transaction ...

    // MUST commit or AutoRollback() will rollback
    tx.Commit()
}
```

`DB` and `Tx` implement `runner.Connection` interface to keep code DRY

```
func getUsers(conn runner.Connection) ([]*dto.Users, error) {
    sql := `
        SELECT *
        FROM users
    `
    var users []*dto.Users
    err := conn.SQL(sql).QueryStructs(&users)
    if err != nil {
        return err
    }
    return users
}
```

### Nested Transactions

Nested transaction logic is as follows:

*   If `Commit` is called in a nested transaction, the operation results in no operation (NOOP).
    Only the top level `Commit` commits the transaction to the database.

*   If `Rollback` is called in a nested transaction, then the entire
    transaction is rolled back. `Tx.IsRollbacked` is set to true.

*   Either `defer Tx.AutoCommit()` or `defer Tx.AutoRollback()` **MUST BE CALLED**
    for each corresponding `Begin`. The internal state of nested transactions is
    tracked in these two methods.

```go
func nested(conn runner.Connection) error {
    tx, err := conn.Begin()
    if err != nil {
        return err
    }
    defer tx.AutoRollback()

    _, err := tx.SQL(`INSERT INTO users (email) values $1`, "me@home.com").Exec()
    if err != nil {
        return err
    }
    // prevents AutoRollback
    tx.Commit()
}

func top() {
    tx, err := DB.Begin()
    if err != nil {
        logger.Fatal("Could not create transaction")
    }
    defer tx.AutoRollback()

    err := nested(tx)
    if err != nil {
        return
    }
    // top level commits the transaction
    tx.Commit()
}
```

### Timeouts

A timeout may be set on any `Query*` or `Exec` with the `Timeout` method. When a
timeout is set, the query is run in a separate goroutine and should a timeout
occur dat will cancel the query via Postgres' `pg_cancel_backend`.

```go
err := DB.Select("SELECT pg_sleep(1)").Timeout(1 * time.Millisecond).Exec()
err == dat.ErrTimedout
```

### Dates

Use `dat.NullTime` type to properly handle nullable dates
from JSON and Postgres.

### Constants

__applicable when dat.EnableInterpolation == true__

`dat` provides often used constants in SQL statements

* `dat.DEFAULT` - inserts `DEFAULT`
* `dat.NOW` - inserts `NOW()`

### Defining Constants

_UnsafeStrings and constants will panic unless_ `dat.EnableInterpolation == true`

To define SQL constants, use `UnsafeString`

```go
const CURRENT_TIMESTAMP = dat.UnsafeString("NOW()")
DB.SQL("UPDATE table SET updated_at = $1", CURRENT_TIMESTAMP)
```

`UnsafeString` is exactly that, **UNSAFE**. If you must use it, create a
constant and **NEVER** use `UnsafeString` directly as an argument like this

```go
DB.SQL("UPDATE table SET updated_at = $1", dat.UnsafeString(someVar))
```

### Primitive Values

Load scalar and slice values.

```go
var id int64
var userID string
err := DB.
    Select("id", "user_id").From("posts").Limit(1).QueryScalar(&id, &userID)

var ids []int64
err = DB.Select("id").From("posts").QuerySlice(&ids)
```

### Caching

dat implements caching backed by an in-memory or Redis store. The in-memory store
is not recommended for production use. Caching can cache any struct or primitive type that
can be marshaled/unmarshaled cleanly with the json package due to Redis being a string
value store.

Time is especially problematic as JavaScript, Postgres and Go
have different time formats. Use the type `dat.NullTime` if you are
getting `cannot parse time` errors.

Caching is performed before the database driver lessening the workload on
the database.

```go
// key-value store (kvs) package
import "gopkg.in/mgutz/dat.v1/kvs"

func init() {
    // Redis: namespace is the prefix for keys and should be unique
    store, err := kvs.NewRedisStore("namespace:", ":6379", "passwordOrEmpty")

    // Or, in-memory store provided by [go-cache](https://github.com/pmylund/go-cache)
    cleanupInterval := 30 * time.Second
    store = kvs.NewMemoryStore(cleanupInterval)

    runner.SetCache(store)
}

// Cache states query for a year using key "namespace:states"
b, err := DB.
    SQL(`SELECT * FROM states`).
    Cache("states", 365 * 24 * time.Hour, false).
    QueryJSON()

// Without a key, the checksum of the query is used as the cache key.
// In this example, the interpolated SQL  will contain their user_name
// (if EnableInterpolation is true) effectively caching each user.
//
// cacheID == checksum("SELECT * FROM users WHERE user_name='mario'")
b, err := DB.
    SQL(`SELECT * FROM users WHERE user_name = $1`, user).
    Cache("", 365 * 24 *  time.Hour, false).
    QueryJSON()

// Prefer using known unique IDs to avoid the computation cost
// of the checksum key.
key = "user" + user.UserName
b, err := DB.
    SQL(`SELECT * FROM users WHERE user_name = $1`, user).
    Cache(key, 15 * time.Minute, false).
    QueryJSON()

// Set invalidate to true to force setting the key
statesUpdated := true
b, err := DB.
    SQL(`SELECT * FROM states`).
    Cache("states", 365 * 24 *  time.Hour, statesUpdated).
    QueryJSON()

// Clears the entire cache
runner.Cache.FlushDB()

runner.Cache.Del("fookey")
```

### SQL Interpolation

__Interpolation is DISABLED by default. Set `dat.EnableInterpolation = true`
to enable.__

`dat` can interpolate locally to inline query arguments. For example,
this statement

go
```
db.Exec(
    "INSERT INTO (a, b, c, d) VALUES ($1, $2, $3, $4)",
    []interface{}[1, 2, 3, 4],
)
```

is sent to the database with inlined args bypassing prepared statement logic in
the lib/pq layer

```sql
"INSERT INTO (a, b, c, d) VALUES (1, 2, 3, 4)"
```

Interpolation provides these benefits:

*   Performance improvements
*   Debugging/tracing is simpler with interpolated SQL
*   May use safe SQL constants like `dat.NOW` and `dat.DEFAULT`
*   Expand placeholders with slice values `$1 => (1, 2, 3)`

Read [SQL Interpolation](https://github.com/mgutz/dat/wiki/Local-Interpolation) in wiki
for more details and SQL injection.

## LICENSE

[The MIT License (MIT)](https://github.com/mgutz/dat/blob/master/LICENSE)
