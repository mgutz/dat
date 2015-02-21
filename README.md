# dat

[GoDoc](https://godoc.org/github.com/mgutz/dat)

`dat` (Data Access Toolkit) is a fast, lightweight and intuitive Postgres
library for Go. `dat` tries to make SQL more accessible.

Highlights

*   Ordinal placeholders - friendlier than `?`

    ```go
    conn.SQL(`SELECT * FROM people WHERE state = $1`, "CA").Exec()
    ```

*   Intuitive - it looks like SQL

    ```go
    err := conn.
        Select("id, user_name").
        From("users").
        Where("id = $1", id).
        QueryStruct(&user)
    ```

*   Multiple Runners - use `sqlx` or `database/sql`

*   Performant

    -   `dat` can interpolates queries locally before sending to server which can speed things up.
    -   ordinal placeholder logic has been optimized to be almost as fast as `?`
        placeholders

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

### Runners

`dat` was designed to have clear separation between SQL builders and Query execers.
There are two runner implementations:

* `sqlx-runner` - based on [sqlx](https://github.com/jmoiron/sqlx)
* `sql-runner` - based on [dbr](https://github.com/gocraft/dbr)

I recommend `sqlx-runner`. There are times when you need to use the driver directly,
for example when structs contain binary types, which cannot be interpolated efficiently
and therefore not supported. `sqlx` is much friendlier than plain `database/sql`
in those situations.

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

### IN queries

Simpler IN queries which expand correctly

```go
ids := []int64{10,20,30,40,50}
b := sess.SQL("SELECT * FROM posts WHERE id IN $1", ids)
b.MustInterpolate() == "SELECT * FROM posts WHERE id IN (10,20,30,40,50)"
```

### Local Interpolation

`dat` interpolates locally using a built-in escape function to inline
query arguments which can result in performance improvements. It's safe.
What is safe? It uses a more strict escape function than the `appendEscapedText`
functino in `https://github.com/lib/pq/blob/master/encode.go`.

__interpolation is disabled by default__, set `dat.EnableInterpolation = true`
to enable this feature. Keep it disabled if you are concerned about the
interpolation.

TODO Add benchmarks

## Usage Examples

### Create a Session

All queries are made in the context of a session which are acquired
from the pool in the underlying SQL driver

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

### CRUD

Create

```go
post := Post{Title: "Swith to Postgres", State: "open"}

// Use Returning() and QueryStruct to update ID and CreatedAt in one trip
err := sess.
    InsertInto("posts").
    Columns("title", "state").
    Record(post).
    Returning("id", "created_at", "updated_at").
    QueryStruct(&post)
```

Read

```go
var other Post
err = sess.
    Select("id, title").
    From("posts").
    Where("id = $1", post.ID).
    QueryStruct(&other)
```

Update

```go
err = sess.
    Update("posts").
    Set("title", "My New Title").
    Where("id = $1", post.ID).
    Returning("updated_at").
    QueryScalar(&post.UpdatedAt)

// To reset values to their default value, use DEFAULT
// eg, to reset payment_type to its default value in DDL
res, err := sess.
    Update("payments").
    Set("payment_type", dat.DEFAULT).
    Where("id = $1", 1).
    Exec()
```

Delete

``` go
response, err = sess.
    DeleteFrom("posts").
    Where("id = $1", otherPost.ID).
    Limit(1).
    Exec()
```

### Constants

`dat` provides often used constants in SQL statements

* dat.DEFAULT - inserts `DEFAULT`
* dat.NOW - inserts `NOW()`

**BEGIN DANGER ZONE**

_UnsafeStrings and constants will panic unless_ `dat.EnableInterpolation=true`

To define your own SQL constants, use `dat.UnsafeString`

```go
const CURRENT_TIMESTAMP = dat.UnsafeString("NOW()")
conn.SQL("UPDATE table SET updated_at = $1", CURRENT_TIMESTAMP)
```

`UnsafeString` is exactly that, **unsafe**. If you must use it, create a constant
and name it according to its SQL usage.

**END**

### Primitive Values

Load scalar and slice primitive values

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

### Inserting Multiple Records

```go
// Start bulding an INSERT statement
b := sess.InsertInto("posts").Columns("title")

// Add some new posts.
for i := 0; i < 3; i++ {
	b.Record(&Post{Title: fmt.Sprintf("Article %s", i)})
}

// OR (this is more efficient)
for i := 0; i < 3; i++ {
	b.Values(fmt.Sprintf("Article %s", i))
}

// Execute statement
_, err := b.Exec()
```


### Updating Records

```go
// Update any rubyists to gophers
result, err := sess.
    Update("posts").
    Set("name", "Gopher").
    Set("language", "Go").
    Where("language = $1", "Ruby").
    Exec()

// Alternatively use a map of attributes
attrsMap := map[string]interface{}{"name": "Gopher", "language": "Go"}
result, err := sess.
    Update("developers").
    SetMap(attrsMap).
    Where("language = $1", "Ruby").
    Exec()
```

### Transactions

```go
// Start transaction
tx, err := sess.Begin()
if err != nil {
    return err
}

// Rollback unless we're successful. tx.Rollback() may also be called manually
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

// Alternatively build the interpolated sql statement for better performance
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
