# dat

[GoDoc](https://godoc.org/github.com/mgutz/dat)

Package dat (Data Access Toolkit) is a fast and convenient
Postgres specific library for Go. `dat`'s goal is to make SQL
more accessible not hide it behind some cumbersome AST dialect.

TODO

* hstore query suppport

## Getting Started

```go
import (
    "database/sql"

    "github.com/mgutz/dat"
    "github.com/mgutz/dat/sql-runner" // use database/sql runner
    _ "github.com/lib/pq"
)

type Suggestion struct {
    ID        int64         `db:"id"`
    Title     string
    CreatedAt dat.NullTime  `db:"created_at"`
}

// create global connection at startup (pooling provided by SQL driver)
var connection *runner.Connection

func main() {
    db, _ := sql.Open("postgres", "dbname=dat_test user=dat password=!test host=localhost sslmode=disable")
    conn = runner.NewConnection(db)

    // fetch a record
    var suggestion Suggestion
    err := conn.
        Select("id, title").
        From("suggestions").
        Where("id = $1", 13).
        QueryStruct(&suggestion)
    fmt.Println("Title", suggestion.Title)
}
```

## Feature highlights

### Fetching Data

Easily map results to structs

```go
var posts []*struct {
    ID int64            `db:"id"`
    Title string
    Body dat.NullString
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

### Use Builders or SQL

Query Builder

```go
// Tip: must be slice to pointers
var posts []*Post
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

### IN queries

Simpler IN queries which expand correctly

```go
ids := []int64{10,20,30,40,50}
b := sess.SQL("SELECT * FROM posts WHERE id IN $1", ids)
b.MustInterpolate() == "SELECT * FROM posts WHERE id IN (10,20,30,40,50)"
```

### Faster Than Using database/sql

`database/sql`'s db.Query("SELECT ...") method via the SQL driver,
creates a prepared statement, executes it, then discards it.
This has performance costs.

`dat` interpolates locally using a built-in escape function to inline
query arguments. The result is less work on the database server,
no prep time and it's safe.

TODO Check out these [benchmarks](https://github.com/tyler-smith/golang-sql-benchmark).

### JSON Friendly

```go
type Foo {
    S1 dat.NullString `json:"str1"`
}
```

`dat.Null*` types marshal to JSON correctly

```json
{ "str1": "Hi!" }
```

## Usage Examples

### Create a Session

All queries are made in the context of a session.

For one-off operations, use `Connection` directly

```go
err := conn.SQL(...).QueryStruct(&suggestion)
```

For multiple operations, create a session

```go
conn = runner.NewConnection(db)

func SuggestionsIndex(rw http.ResponseWriter, r *http.Request) {
    sess := conn.NewSession()

    // Do queries with the session
    var suggestion Suggestion
    err := sess.Select("id, title").
        From("suggestions").
        Where("id = $1", suggestion.ID).
        QueryStruct(&suggestion)
    )

    // do more queries
}
```

### CRUD

Create

```go
suggestion := &Suggestion{Title: "Swith to Postgres", State: "open"}

// Use Returning() and QueryStruct to update ID and CreatedAt in one trip
response, err := sess.
    InsertInto("suggestions").
    Columns("title", "state").
    Record(suggestion).
    Returning("id", "created_at", "updated_at").
    QueryStruct(&suggestion)
```

Read

```go
var other Suggestion
err = sess.
    Select("id, title").
    From("suggestions").
    Where("id = $1", suggestion.ID).
    QueryStruct(&other)
```

Update

```go
err = sess.
    Update("suggestions").
    Set("title", "My New Title").
    Where("id = $1", suggestion.ID).
    Returning("updated_at").
    QueryScalar(&suggestion.UpdatedAt)

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
    DeleteFrom("suggestions").
    Where("id = $1", otherSuggestion.ID).
    Limit(1).
    Exec()
```

### Primitive Values

Load scalar and slice primitive values

```go
var id int64
var userID string
n, err := sess.
    Select("id", "user_id").From("suggestions").Limit(1).QueryScalar(&id, &userID)

var ids []int64
n, err := sess.Select("id").From("suggestions").QuerySlice(&ids)
```

### Overriding Column Names With Struct Tags

By default `dat` maps CamelCase field names to snake\_case column names.
The column name can be overridden with struct tags. Be careful of fields
named like `ID`, `UserID`, which in snake case are `i_d`, `user_i_d`.
If in doubt, use struct tags.

```go
type Suggestion struct {
    ID        int64           `db:"id"`
    UserID    dat.NullString  `db:"user_id"`
    UpdatedAt data.NullTime
    CreatedAt dat.NullTime
}
```

Tip: Set

### Embedded structs

```go
// Columns are mapped to fields breadth-first
type Suggestion struct {
    ID        int64         `db:"id"`
    Title     string
    User      *struct {
        ID int64 `db:"user_id"`
    }
}

var suggestion Suggestion
err := sess.
    Select("id, title, user_id").
    From("suggestions").
    Limit(1).
    QueryStruct(&suggestion)
```

### JSON encoding of Null\* types

```go
// dat.Null* types serialize to JSON properly
suggestion := &Suggestion{ID: 1, Title: "Test Title"}
jsonBytes, err := json.Marshal(&suggestion)
fmt.Println(string(jsonBytes)) // {"id":1,"title":"Test Title","created_at":null}
```

### Inserting Multiple Records

```go
// Start bulding an INSERT statement
b := sess.InsertInto("developers").Columns("name", "language", "employee_number")

// Add some new developers
for i := 0; i < 3; i++ {
	b.Record(&Dev{Name: "Gopher", Language: "Go", EmployeeNumber: i})
}

// Execute statement
_, err := b.Exec()
```

### Updating Records

```go
// Update any rubyists to gophers
result, err := sess.
    Update("developers").
    Set("name", "Gopher").
    Set("language", "Go").
    Where("language = $1", "Ruby").
    Exec()

// Alternatively use a map of attributes to update
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

// Rollback unless we're successful. tx.Rollback() may also be called manually.
defer tx.RollbackUnlessCommitted()

// Issue statements that might cause errors
res, err := tx.
    Update("suggestions").
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

### Use With Other Libraries (sqlx, ...)

Use the `github.com/mgutz/dat` package which contains the various
SQL builders.

```go
import "github.com/mgutz/dat"
b := dat.Select("*").From("suggestions").Where("subdomain_id = $1", 1)

// Get builder's SQL and arguments
sql, args := b.ToSQL()
fmt.Println(sql)    // SELECT * FROM suggestions WHERE (subdomain_id = $1)
fmt.Println(args)   // [1]

// Use raw database/sql for actual query
rows, err := db.Query(sql, args...)

// Alternatively build the interpolated sql statement for better performance
sql := builder.MustInterpolate()
rows, err := db.Query(sql)
```

## Inspiration

*   [mapper](https://github.com/mgutz/mapper)

    my SQL builder for node.js which has builder, interpolation and exec
    functionality

*   [dbr](https://github.com/gocraft/dbr)

    used this as starting point instead of porting mapper from scratch

