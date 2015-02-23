# dat

[GoDoc](https://godoc.org/github.com/mgutz/dat)

`dat` (Data Access Toolkit) is a fast, lightweight and intuitive Postgres
library for Go. `dat` likes SQL and so should you.

Highlights

*   Ordinal placeholders - friendlier than `?`

    ```go
    conn.SQL(`SELECT * FROM people WHERE state = $1`, "CA").Exec()
    ```

*   Intuitive - looks like SQL

    ```go
    err := conn.
        Select("id, user_name").
        From("users").
        Where("id = $1", id).
        QueryStruct(&user)
    ```

*   Multiple Runners - use `sqlx` or `database/sql`

*   Performant

    -   ordinal placeholder logic has been optimized to be nearly as fast as `?`
        placeholders
    -   `dat` can interpolate queries locally before sending to server which
        can speed things up

## Getting Started

```go
import (
    "database/sql"

    "github.com/mgutz/dat"
    "github.com/mgutz/dat/sqlx-runner"
    _ "github.com/lib/pq"
)

// global connection (pooling provided by SQL driver)
var connection *runner.Connection

func init() {
    // create a normal database connection through database/sql
    db, err := sql.Open("postgres", "dbname=dat_test user=dat password=!test host=localhost sslmode=disable")
    if err != nil {
        panic(err)
    }

    // set this to true to enable interpolation
    dat.EnableInterpolation = true
    conn = runner.NewConnection(db, "postgres")
}

type Post struct {
    ID        int64         `db:"id"`
    Title     string        `db:"title"`
    Body      string        `db:"body"`
    UserID    int64         `db:"user_id"`
    State     string        `db:"state"`
    UpdatedAt dat.Nulltime  `db:"updated_at"`
    CreatedAt dat.NullTime  `db:"creatd_at"`
}

func main() {
    var post Post
    err := conn.
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
var posts []Post
n, err := sess.
    Select("title", "body").
    From("posts").
    Where("created_at > $1", someTime).
    OrderBy("id ASC").
    Limit(10).
    QueryStructs(&posts)
```

Plain SQL

```go
sess.SQL(`
    SELECT title, body
    FROM posts WHERE created_at > $1
    ORDER BY id ASC LIMIT 10`,
    someTime,
).QueryStructs(&posts)
```

### Fetch Data Simply

Easily map results to structs

```go
var posts []struct {
    ID int64            `db:"id"`
    Title string        `db:"title"`
    Body dat.NullString `db:"body"`
}
err := sess.
    Select("id, title, body").
    From("posts").
    Where("id = $1", id).
    QueryStructs(&posts)
```

Query scalar values or a slice of values

```go
var n int64, ids []int64

sess.SQL("SELECT count(*) FROM posts WHERE title=$1", title).QueryScalar(&n)
sess.SQL("SELECT id FROM posts", title).QuerySlice(&ids)
```

### Blacklist and Whitelist

Control which columns get inserted or updated when processing external data

```go
// userData came in from http.Handler, prevent them from setting protected fields
conn.InsertInto("payments").
    SetBlacklist(userData, "id", "updated_at", "created_at").
    Returning("id").
    QueryScalar(&userData.ID)


// ensure session user can only update his information
conn.Update("users").
    SetWhitelist(user, "user_name", "avatar", "quote").
    Where("id = $1", session.UserID).
    Exec()
```

### IN queries

__applicable when dat.EnableInterpolation == true__

Simpler IN queries which expand correctly

```go
ids := []int64{10,20,30,40,50}
b := sess.SQL("SELECT * FROM posts WHERE id IN $1", ids)
b.MustInterpolate() == "SELECT * FROM posts WHERE id IN (10,20,30,40,50)"
```

### Runners

`dat` was designed to have clear separation between SQL builders and Query execers.
There are two runner implementations:

*   `sqlx-runner` - based on [sqlx](https://github.com/jmoiron/sqlx)
*   `sql-runner` - based on [dbr](https://github.com/gocraft/dbr)

    __sql-runner will not be supported in the future__ The database/sql logic is
    based on legacy code from the dbr project with some of my fixes and tweaks.
    I feel sqlx complements `dat` better since interpolation is disabled by default.

## CRUD

### Create

Use `Returning` and `QueryStruct` to insert and update struct fields in one
trip.

```go
post := Post{Title: "Swith to Postgres", State: "open"}

err := sess.
    InsertInto("posts").
    Columns("title", "state").
    Record(post).
    Returning("id", "created_at", "updated_at").
    QueryStruct(&post)
```

Use `Blacklist` and `Whitelist` to control which record columns get
inserted.

```go
post := Post{Title: "Swith to Postgres", State: "open"}

err := sess.
    InsertInto("posts").
    Blacklist("id", "user_id", "created_at", "updated_at").
    Record(post).
    Returning("id", "created_at", "updated_at").
    QueryStruct(&post)

// probably not safe but you get the idea
err := sess.
    InsertInto("posts").
    Whitelist("*").
    Record(post).
    Returning("id", "created_at", "updated_at").
    QueryStruct(&post)
```

Insert Multiple Records

```go
// create builder
b := sess.InsertInto("posts").Columns("title")

// add some new posts
for i := 0; i < 3; i++ {
	b.Record(&Post{Title: fmt.Sprintf("Article %s", i)})
}

// OR (this is more efficient as it does not do any reflection)
for i := 0; i < 3; i++ {
	b.Values(fmt.Sprintf("Article %s", i))
}

// Execute statement
_, err := b.Exec()
```

### Read

```go
var other Post
err = sess.
    Select("id, title").
    From("posts").
    Where("id = $1", post.ID).
    QueryStruct(&other)
```

### Update

Use `Returning` to fetch columns updated by triggers. For example,
there might be an update trigger on "updated\_at" column

```go
err = sess.
    Update("posts").
    Set("title", "My New Title").
    Set("body", "markdown text here").
    Where("id = $1", post.ID).
    Returning("updated_at").
    QueryScalar(&post.UpdatedAt)
```

To reset columns to their default value, use `DEFAULT`. For example,
to reset `payment\_type` to its default value from DDL

__applicable when dat.EnableInterpolation == true__

```go
res, err := sess.
    Update("payments").
    Set("payment_type", dat.DEFAULT).
    Where("id = $1", 1).
    Exec()
```

Use `Blacklist` and `Whitelist` to control which columns get updated.

```go
// create blacklists for each of your structs
blacklist := []string{"id", "created_at"}
p := paymentStructFromHandler

err := sess.
    Update("payments").
    SetBlacklist(p, blacklist...)
    Where("id = $1", p.ID).
    Exec()
```

Use a map of attributes

``` go
attrsMap := map[string]interface{}{"name": "Gopher", "language": "Go"}
result, err := sess.
    Update("developers").
    SetMap(attrsMap).
    Where("language = $1", "Ruby").
    Exec()
```

### Delete

``` go
result, err = sess.
    DeleteFrom("posts").
    Where("id = $1", otherPost.ID).
    Limit(1).
    Exec()
```

### Create a Session

All queries are made in the context of a session which are acquired
from the underlying SQL driver's pool

For one-off operations, use a `Connection` directly

```go
// a global connection usually created in `init`
conn = runner.NewConnection(db, "postgres")

err := conn.SQL(...).QueryStruct(&post)
```

For multiple operations, create a session

```go

func PostsIndex(rw http.ResponseWriter, r *http.Request) {
    sess := conn.NewSession()

    // Do queries with the session
    var post Post
    err := sess.Select("id, title").
        From("posts").
        Where("id = $1", post.ID).
        QueryStruct(&post)
    )

    // do more queries with session
}
```

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
conn.SQL("UPDATE table SET updated_at = $1", CURRENT_TIMESTAMP)
```

`UnsafeString` is exactly that, **UNSAFE**. If you must use it, create a
constant and **NEVER** use `UnsafeString` directly as an argument. This
is asking for a SQL injection attack

```go
conn.SQL("UPDATE table SET updated_at = $1", dat.UnsafeString(someVar))
```

### Primitive Values

Load scalar and slice values.

```go
var id int64
var userID string
err := sess.
    Select("id", "user_id").From("posts").Limit(1).QueryScalar(&id, &userID)

var ids []int64
err = sess.Select("id").From("posts").QuerySlice(&ids)
```

### Embedded structs

```go
// Columns are mapped to fields breadth-first
type Post struct {
    ID        int64      `db:"id"`
    Title     string     `db:"title"`
    User      *struct {
        ID int64         `db:"user_id"`
    }
}

var post Post
err := sess.
    Select("id, title, user_id").
    From("posts").
    Limit(1).
    QueryStruct(&post)
```

### JSON encoding of Null\* types

```go
// dat.Null* types serialize to JSON properly
post := Post{ID: 1, Title: "Test Title"}
jsonBytes, err := json.Marshal(&post)
fmt.Println(string(jsonBytes)) // {"id":1,"title":"Test Title","created_at":null}
```

### Transactions

```go
// Start transaction
tx, err := sess.Begin()
if err != nil {
    return err
}

// tx.Rollback() may also be called manually
defer tx.RollbackUnlessCommitted()

// Issue statements that might cause errors
res, err := tx.
    Update("posts").
    Set("state", "deleted").
    Where("deleted_at IS NOT NULL").
    Exec()

if err != nil {
    return err
}

// Commit the transaction
if err := tx.Commit(); err != nil {
	return err
}
```

### Local Interpolation

`dat` can interpolate locally using a built-in escape function to inline
query arguments. Some of the reasons you might want to use interpolation:

*   Interpolation can result in perfomance improvements.
*   Debugging is simpler too hen looking at the interpolated SQL in your logs.
*   Enhanced features like use of dat.NOW and data.DEFAULT, inling slice
    args ....

__Interpolation is DISABLED by default__ Set `dat.EnableInterpolation = true`
to enable.

Is interpolation safe? As of Postgres 9.1, escaping is disabled by default. See
[String Constants with C-style Escapes](http://www.postgresql.org/docs/9.3/interactive/sql-syntax-lexical.html#SQL-SYNTAX-STRINGS-ESCAPE).
The built-in interpolation func disallows **ALL** escape sequences.

`dat` checks the value of `standard_conforming_strings` on a new connection if
`data.EnableInterpolation == true`. If `standard_conforming_strings != "on"`
you should either set it to "on" or disable interpolation. `dat` will panic
if you try to use interpolation with an incorrect setting.

### Benchmarks

TODO Add interpolation  benchmarks


### Use With Other Libraries

```go
import "github.com/mgutz/dat"

builder := dat.Select("*").From("posts").Where("user_id = $1", 1)

// Get builder's SQL and arguments
sql, args := builder.ToSQL()
fmt.Println(sql)    // SELECT * FROM posts WHERE (user_id = $1)
fmt.Println(args)   // [1]

// Use raw database/sql for actual query
rows, err := db.Query(sql, args...)

// Alternatively build the interpolated sql statement
sql := builder.MustInterpolate()
rows, err := db.Query(sql)
```

## TODO

* more tests
* hstore query suppport

## Inspiration

*   [mapper](https://github.com/mgutz/mapper)

    My SQL builder for node.js which has builder, interpolation and exec
    functionality.

*   [dbr](https://github.com/gocraft/dbr)

    used this as starting point instead of porting mapper from scratch
