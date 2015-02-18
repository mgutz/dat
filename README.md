# dat

[GoDoc](https://godoc.org/github.com/mgutz/dat)

github.com/mgutz/dat is a Data Access Toolkit for Go built for speed and convenience.

## Getting Started

```go
import (
    "database/sql"
    _ "github.com/lib/pq"
    "github.com/mgutz/dat"
    "github.com/mgutz/dat/sql-runner" // use database/sql runner
)

type Suggestion struct {
    ID        int64         `db:"id"`
    Title     string
    CreatedAt dat.NullTime  `db:"created_at"`
}

// Hold a single global connection (pooling provided by sql driver)
var connection *runner.Connection

func main() {
    // Create the connection during application initialization
    db, _ := sql.Open("postgres", "dbname=dat_test user=dat password=!test host=localhost sslmode=disable")
    connection = runner.NewConnection(db)

    // Create a session for each unit of execution, eg. each http.Handler
    sess := connection.NewSession()

    // Get a record
    var suggestion Suggestion
    err := sess.QueryStruct(&suggestion,
        dat.Select("id, title").From("suggestions").Where("id = $1", 13),
    )

    if err != nil {
        panic(err.Error())
    }
    fmt.Println("Title", suggestion.Title)
}
```

## Feature highlights

### Fetching into Variables

Automatically map results to structs

```go
var posts []*struct {
    ID int64            `db:"id"`
    Title string
    Body dat.NullString
}
err := sess.QueryStructs(&post,
    dat.Select("id, title, body").
        From("posts").
        Where("id = $1", id)
)
```

Query a scalar value or slice of values

```go
var n int64, ids []int64

sess.QueryScalar(&n, dat.SQL("SELECT count(*) FROM posts WHERE title=$1", title))
sess.QuerySlice(&ids, dat.SQL("SELECT id FROM posts", title))
```

### Use Query Builders or Plain SQL

Query Builder

```go
b := dat.Select("title", "body").
    From("posts").
    Where("created_at > $1", someTime).
    OrderBy("id ASC").
    Limit(10)

// Tip: must be slice to pointers
var posts []*Post
n, err := sess.QueryStructs(&posts, b)
```

Plain SQL

```go
b := dat.SQL(`
    SELECT title, body
    FROM posts WHERE created_at > $1
    ORDER BY id ASC LIMIT 10`,
    someTime,
)
n, err := sess.QueryStructs(&posts, b)
```

### IN queries

Simpler IN queries

Traditional Way

```go
ids := []int64{1,2,3,4,5}
questionMarks := []string
for _, _ := range ids {
    questionMarks = append(questionMarks, "?")
}
query := fmt.Sprintf("SELECT * FROM posts WHERE id IN (%s)",
    strings.Join(questionMarks, ",")
```

The easy way with dat

```go
ids := []int64{10,20,30,40,50}
b := dat.SQL("SELECT * FROM posts WHERE id IN $1", ids)
b.MustInterpolate() == "SELECT * FROM posts WHERE id IN (10,20,30,40,50)"
```

### Instrumentation

Writing instrumented code is a first-class concern.

### Faster Than Using database/sql

Every time you call database/sql's db.Query("SELECT ...") method,
under the hood, the mysql driver will create a prepared statement,
execute it, and then throw it away. This has a big performance cost.

`dat` doesn't use prepared statements. Postgres' escape functionality is
built-in, which means all queries are interpolated on the current server.
The result is less work on the database server, no prep time and it's safe.

Check out these [benchmarks](https://github.com/tyler-smith/golang-sql-benchmark).

### JSON Friendly

```go
type Foo {
    S1 dat.NullString `json:"str1"`
    S2 dat.NullString `json:"str2"`
}
```

`dat.Null*` types marshal to JSON correctly

```json
{
    "str1": "Hi!",
    "str2": null
}
```

## Driver support

Currently only PostgreSQL has been tested.

## Usage Examples

### Making a session

All queries in mgutz/dat are made in the context of a session.

Here's an example web endpoint that makes a session:

```go
connection = runner.NewConnection(db)

func SuggestionsIndex(rw http.ResponseWriter, r *http.Request) {
    sess := connection.NewSession()

    // Do queries with the session
    var suggestion Suggestion
    err := sess.QueryStruct(&suggestion,
        dat.Select("id, title").
            From("suggestions").
            Where("id = $1", suggestion.ID),
    )

    // Render etc. Nothing else needs to be done with the sesssion.
}
```

### CRUD

```go
suggestion := &Suggestion{Title: "My Cool Suggestion", State: "open"}

// Create - QueryStruct updates ID and CreatedAt, taking advantage of
// Postgres' RETURNING clause
response, err := sess.QueryStruct(&suggestion,
    dat.InsertInto("suggestions").
        Columns("title", "state").
        Record(suggestion).
        Returning("id", "created_at"),
)

// Read
var otherSuggestion Suggestion
err = sess.QueryStruct(&otherSuggestion,
    dat.Select("id, title").
        From("suggestions").
        Where("id = $1", suggestion.Id),
)

// Update
response, err = sess.Exec(
    dat.Update("suggestions").
        Set("title", "My New Title").
        Where("id = $1", suggestion.ID),
)

// Delete
response, err = sess.Exec(
    dat.DeleteFrom("suggestions").
        Where("id = $1", otherSuggestion.Id).
        Limit(1),
)
```

### Primitive Values

Load scalar and slice primitive values

```go
var id int64
n, err := sess.QueryScalar(&id,
    dat.Select("id").From("suggestions").Limit(1)
)

var ids []int64
n, err := sess.QuerySlice(&ids,
    dat.Select("id").From("suggestions")
)
```

### Overriding Column Names With Struct Tags

```go
// By default dat converts CamelCase property names to snake_case column_names
// You can override this with struct tags, just like with JSON tags
// This is especially helpful while migrating from legacy systems

type Suggestion struct {
    ID        int64 `db:"id"`
    Title     dat.NullString `db:"subject"` // subjects are called titles now
    CreatedAt dat.NullTime
}
```

### Embedded structs

```go
// Columns are mapped to fields breadth-first
type Suggestion struct {
    Id        int64
    Title     string
    User      *struct {
        Id int64 `db:"user_id"`
    }
}

var suggestion Suggestion
err := sess.QueryStruct(&suggestion,
    dat.Select("id, title, user_id").From("suggestions").Limit(1),
)
```

### JSON encoding of Null\* types

```go
// dat.Null* types serialize to JSON properly
suggestion := &Suggestion{Id: 1, Title: "Test Title"}
jsonBytes, err := json.Marshal(&suggestion)
fmt.Println(string(jsonBytes)) // {"id":1,"title":"Test Title","created_at":null}
```

### Inserting Multiple Records

```go
// Start bulding an INSERT statement
b := dat.InsertInto("developers").
	Columns("name", "language", "employee_number")

// Add some new developers
for i := 0; i < 3; i++ {
	b.Record(&Dev{Name: "Gopher", Language: "Go", EmployeeNumber: i})
}

// Execute statement
_, err := sess.Exec(b)
```

### Updating Records

```go
// Update any rubyists to gophers
response, err := sess.Exec(
    dat.Update("developers").
        Set("name", "Gopher").
        Set("language", "Go").
        Where("language = $1", "Ruby")
)


// Alternatively use a map of attributes to update
attrsMap := map[string]interface{}{"name": "Gopher", "language": "Go"}
response, err := sess.Exec(
    dat.Update("developers").SetMap(attrsMap).Where("language = $1", "Ruby")
)
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
res, err := tx.Exec(
    dat.Update("suggestions").
        Set("state", "deleted").
        Where("deleted_at IS NOT NULL"),
)
if err != nil {
    return err
}

// Commit the transaction
if err := tx.Commit(); err != nil {
	return err
}
```

### Use With Other Runners (sqlx, pgx)

```go
b := dat.Select("*").From("suggestions").Where("subdomain_id = $1", 1)

// Get builder's SQL and arguments
sql, args := b.ToSQL()
fmt.Println(sql)    // SELECT * FROM suggestions WHERE (subdomain_id = $1)
fmt.Println(args)   // [1]

// Use raw database/sql for actual query
rows, err := db.Query(sql, args...)

// Alternatively build the interpolated sql statement
sql := builder.MustInterpolate()
rows, err := db.Query(sql)
```

## Thanks

Inspiration from

*  [mapper](https://github.com/mgutz/mapper) - my data access library for node, which predates dbr
*  [dbr](https://github.com/gocraft/dbr) - builder code

