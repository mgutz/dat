# dat

[GoDoc](https://godoc.org/github.com/mgutz/dat)

`dat` (Data Access Toolkit) is a fast, lightweight and intuitive Postgres
library for Go. `dat` likes SQL.

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

*   Performant

    -   ordinal placeholder logic has been optimized to be nearly as fast as `?`
        placeholders
    -   `dat` can interpolate queries locally before sending to server

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

    // set this to enable interpolation
    dat.EnableInterpolation = true
    // set to log SQL, etc
    dat.SetVerbose(false)
    // set to check things like sessions closing.
    // Should be disabled in production/release builds.
    dat.Strict = false
    conn = runner.NewConnection(db, "postgres")
}

type Post struct {
    ID        int64         `db:"id"`
    Title     string        `db:"title"`
    Body      string        `db:"body"`
    UserID    int64         `db:"user_id"`
    State     string        `db:"state"`
    UpdatedAt dat.Nulltime  `db:"updated_at"`
    CreatedAt dat.NullTime  `db:"created_at"`
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
var posts []*Post
err := conn.
    Select("title", "body").
    From("posts").
    Where("created_at > $1", someTime).
    OrderBy("id ASC").
    Limit(10).
    QueryStructs(&posts)
```

Plain SQL

```go
conn.SQL(`
    SELECT title, body
    FROM posts WHERE created_at > $1
    ORDER BY id ASC LIMIT 10`,
    someTime,
).QueryStructs(&posts)
```

Note: `dat` does not clean the SQL string, thus any extra whitespace is
transmitted to the database.

In practice, SQL is easier to write with backticks. Indeed, the reason this
library exists is my dissatisfaction with other SQL builders introducing
another domain language or AST-like expressions.

Query builders shine when dealing with data transfer objects,
records (input structs).


### Fetch Data Simply

Query then scan result to struct(s)

```go
var post Post
err := sess.
    Select("id, title, body").
    From("posts").
    Where("id = $1", id).
    QueryStruct(&post)

var posts []*Post
err = sess.
    Select("id, title, body").
    From("posts").
    Where("id > $1", 100).
    QueryStructs(&posts)
```

Query scalar values or a slice of values

```go
var n int64
conn.SQL("SELECT count(*) FROM posts WHERE title=$1", title).QueryScalar(&n)

var ids []int64
conn.SQL("SELECT id FROM posts", title).QuerySlice(&ids)
```

### Blacklist and Whitelist

Control which columns get inserted or updated when processing external data

```go
// userData came in from http.Handler, prevent them from setting protected fields
conn.InsertInto("payments").
    Blacklist("id", "updated_at", "created_at").
    Record(userData).
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
b := conn.SQL("SELECT * FROM posts WHERE id IN $1", ids)
b.MustInterpolate() == "SELECT * FROM posts WHERE id IN (10,20,30,40,50)"
```

### Runners

`dat` was designed to have clear separation between SQL builders and Query execers.
This is why the runner is in its own package.

*   `sqlx-runner` - based on [sqlx](https://github.com/jmoiron/sqlx)

## CRUD

### Create

Use `Returning` and `QueryStruct` to insert and update struct fields in one
trip

```go
post := Post{Title: "Swith to Postgres", State: "open"}

err := conn.
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

err := conn.
    InsertInto("posts").
    Blacklist("id", "user_id", "created_at", "updated_at").
    Record(post).
    Returning("id", "created_at", "updated_at").
    QueryStruct(&post)

// use wildcard to include all columns
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
b := conn.InsertInto("posts").Columns("title")

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

### Read

```go
var other Post
err = conn.
    Select("id, title").
    From("posts").
    Where("id = $1", post.ID).
    QueryStruct(&other)
```

### Update

Use `Returning` to fetch columns updated by triggers. For example,
an update trigger on "updated\_at" column

```go
err = conn.
    Update("posts").
    Set("title", "My New Title").
    Set("body", "markdown text here").
    Where("id = $1", post.ID).
    Returning("updated_at").
    QueryScalar(&post.UpdatedAt)
```

__applicable when dat.EnableInterpolation == true__

To reset columns to their default DDL value, use `DEFAULT`. For example,
to reset `payment\_type`

```go
res, err := conn.
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

err := conn.
    Update("payments").
    SetBlacklist(p, blacklist...)
    Where("id = $1", p.ID).
    Exec()
```

Use a map of attributes

``` go
attrsMap := map[string]interface{}{"name": "Gopher", "language": "Go"}
result, err := conn.
    Update("developers").
    SetMap(attrsMap).
    Where("language = $1", "Ruby").
    Exec()
```

### Delete

``` go
result, err = conn.
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
var conn *dat.Connection
conn = runner.NewConnection(db, "postgres")

err := conn.SQL(...).QueryStruct(&post)
```

For multiple operations, create a session. Note that session
is really a transaction due to `database/sql` connection pooling.
__`Session.AutoCommit() or Session.AutoRollback()` MUST be called__

```go

func PostsIndex(rw http.ResponseWriter, r *http.Request) {
    sess := conn.NewSession()
    defer sess.AutoRollback()

    // Do queries with the session
    var post Post
    err := sess.Select("id, title").
        From("posts").
        Where("id = $1", post.ID).
        QueryStruct(&post)
    )
    if err != nil {
    	// `defer AutoRollback()` is used, no need to rollback on error
    	r.WriteHeader(500)
    	return
    }

    // do more queries with session ...

    // MUST commit or AutoRollback() will rollback
    sess.Commit()
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
err := conn.
    Select("id", "user_id").From("posts").Limit(1).QueryScalar(&id, &userID)

var ids []int64
err = conn.Select("id").From("posts").QuerySlice(&ids)
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
err := conn.
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
tx, err := conn.Begin()
if err != nil {
    return err
}
// safe to call tx.Rollback() or tx.Commit() when deferring AutoCommit()
defer tx.AutoCommit()

// AutoRollback() is also available if you would rather Commit() at the end
// and not deal with Rollback on every error.

// Issue statements that might cause errors
res, err := tx.
    Update("posts").
    Set("state", "deleted").
    Where("deleted_at IS NOT NULL").
    Exec()

if err != nil {
    tx.Rollback()
    return err
}
```

### Local Interpolation

TL;DR: Interpolation avoids prepared statements and argument processing.

__Interpolation is DISABLED by default. Set `dat.EnableInterpolation = true`
to enable.__

`dat` can interpolate locally using a built-in escape function to inline
query arguments. What is interpolation? An interpolated statement has all
arguments inlined and often results in a single SQL statement with no arguments
sent to the DB:

```
"INSERT INTO (a, b, c, d) VALUES (1, 2, 3, 4)"
```

Non-interpolated statements use prepared statements underneath and
send statements with arguments to the DB:

```
"INSERT INTO (a, b, c, d) VALUES ($1, $2, $3, $4)",
[]interface{}[1, 2, 3, 4]
```

Some of the reasons you might want to use interpolation:

*   Interpolation can result in performance improvements
*   Debugging is simpler with interpolated SQL
*   Use SQL constants like `NOW` and `DEFAULT`
*   Expand placeholders with expanded slice values `$1 => (1, 2, 3)`

`[]byte`,  `[]*byte` and any unhandled values are passed through to the
driver when interpolating.

#### Interpolation Safety

As of Postgres 9.1, escaping is disabled by default. See
[String Constants with C-style Escapes](http://www.postgresql.org/docs/9.3/interactive/sql-syntax-lexical.html#SQL-SYNTAX-STRINGS-ESCAPE).

`dat` disallows **ALL** escape sequences when interpolating.

`dat` checks the Postgres database `standard_conforming_strings` setting value on a new connection when
`dat.EnableInterpolation == true`. If `standard_conforming_strings != "on"`
you should either set it to `"on"` or disable interpolation. `dat` will panic
if you try to use interpolation with an incorrect setting.

#### Why is Interpolation Faster?

Here is a comment from [lib/pq connection source](https://github.com/lib/pq/blob/master/conn.go),
which was prompted by me asking why was Python's psycopg2 so much
faster in my benchmarks a year or so back:

```go
// Check to see if we can use the "simpleExec" interface, which is
// *much* faster than going through prepare/exec
if len(args) == 0 {
    // ignore commandTag, our caller doesn't care
    r, _, err := cn.simpleExec(query)
    return r, err
}
```

That snippet bypasses the prepare/exec roundtrip to the database.

Keep in mind that prepared statements are only valid for the current
session and uless the same query will be executed *MANY* times in the
same session there is little benefit in using prepared statements.
One benefit of using prepared statements is they provide
safety against SQL injection by parameterizing queries.
See Interpolation Safety below.

Another benefit of interpolation is offloading dabatabase workload to your
application servers. There is less work and less network chatter when
interpolation is performed locally. It's usually much simpler to add application servers
than to vertically scale a database server.

#### Benchmarks

* Dat2 - mgutz/dat runner with 2 args
* Sql2 - database/sql with 2 args
* Sqx2 - jmoiron/sqlx with 2 args

Replace 2 with 4, 8 for variants of argument benchmarks. All source is under
sqlx-runner/benchmark\*

#### Interpolated v Non-Interpolated Queries

This benchmark compares the time to execute an interpolated SQL
statement with zero args against executing the same SQL statement with
args.

```
BenchmarkExecSQLDat2       5000   208345   ns/op  280   B/op  10  allocs/op
BenchmarkExecSQLSql2       5000   298789   ns/op  881   B/op  30  allocs/op
BenchmarkExecSQLSqx2       5000   296948   ns/op  881   B/op  30  allocs/op

BenchmarkExecSQLDat4       5000   210759   ns/op  296   B/op  10  allocs/op
BenchmarkExecSQLSql4       5000   306558   ns/op  978   B/op  35  allocs/op
BenchmarkExecSQLSqx4       5000   305569   ns/op  978   B/op  35  allocs/op
```

The logic is something like this

```go
// already interpolated
for i := 0; i < b.N; i++ {
    conn.Exec("INSERT INTO t (a, b, c, d) VALUES (1, 2, 3 4)")
}

// not interpolated
for i := 0; i < b.N; i++ {
    db.Exec("INSERT INTO t (a, b, c, d) VALUES ($1, $2, $3, $4)", 1, 2, 3, 4)
}
```

To be fair, this benchmark is not meaningful. It doesn't take into account
the time to perform the interpolation. It's only meant to show that
interpolated queries avoid the overhead of arguments and skip the prepare statement
logic in the underlying driver.

#### Interpolating then Execing

This benchmark compares the time to build and execute interpolated SQL
statement resulting in zero args against executing the same SQL statement with
args.

```
BenchmarkBuildExecSQLDat2  5000   215449   ns/op  832   B/op  21  allocs/op
BenchmarkBuildExecSQLSql2  5000   296281   ns/op  881   B/op  30  allocs/op
BenchmarkBuildExecSQLSqx2  5000   296259   ns/op  881   B/op  30  allocs/op

BenchmarkBuildExecSQLDat4  5000   221287   ns/op  1232  B/op  26  allocs/op
BenchmarkBuildExecSQLSql4  5000   305807   ns/op  978   B/op  35  allocs/op
BenchmarkBuildExecSQLSqx4  5000   305671   ns/op  978   B/op  35  allocs/op

BenchmarkBuildExecSQLDat8  5000   254252   ns/op  1480  B/op  33  allocs/op
BenchmarkBuildExecSQLSql8  5000   347407   ns/op  1194  B/op  44  allocs/op
BenchmarkBuildExecSQLSqx8  5000   346576   ns/op  1194  B/op  44  allocs/op
```

The logic is something like this

```go
// dat's SQL interpolates the statment then exececutes
for i := 0; i < b.N; i++ {
    conn.SQL("INSERT INTO (a, b, c, d) VALUES ($1, $2, $3, $4)", 1, 2, 3, 4).Exec()
}

// non-interpolated
for i := 0; i < b.N; i++ {
    db.Exec("INSERT INTO (a, b, c, d) VALUES ($1, $2, $3, $4)", 1, 2, 3, 4)
}
```

The results suggests that local interpolation is both faster and does less
allocation. Interpolation comes with a cost of more bytes used as it has
to inspect the args and splice them into the statement.

database/sql when presented with arguments prepares a
statement on the connection by sending it to the database then using the
prepared statement on the same connection to execute the query.
Keep in mind, these benchmarks are local so network latency is not a factor
which would favor interpolation even more.

### Interpolation and Transactions

This benchmark compares the performance of interpolation within a transaction on
"level playing field" with database/sql. As mentioned in a previous
section, prepared statements MUST be prepared and executed on the same
connection to utilize them.

```
BenchmarkTransactedDat2    10000  111959   ns/op  832   B/op  21  allocs/op
BenchmarkTransactedSql2    10000  173137   ns/op  881   B/op  30  allocs/op
BenchmarkTransactedSqx2    10000  175342   ns/op  881   B/op  30  allocs/op

BenchmarkTransactedDat4    10000  115383   ns/op  1232  B/op  26  allocs/op
BenchmarkTransactedSql4    10000  182626   ns/op  978   B/op  35  allocs/op
BenchmarkTransactedSqx4    10000  181641   ns/op  978   B/op  35  allocs/op

BenchmarkTransactedDat8    10000  145419   ns/op  1480  B/op  33  allocs/op
BenchmarkTransactedSql8    10000  221476   ns/op  1194  B/op  44  allocs/op
BenchmarkTransactedSqx8    10000  222460   ns/op  1194  B/op  44  allocs/op
```

The logic is something like this

```go
// dat: interpolate the statement then exececute as within the transactionn
tx := conn.Begin()
defer tx.Commit()
for i := 0; i < b.N; i++ {
	tx.SQL("INSERT INTO (a, b, c, d) VALUES ($1, $2, $3, $4)", 1, 2, 3, 4).Exec()
}

// non-interpolated
tx = db.Begin()
defer tx.Commit()
for i := 0; i < b.N; i++ {
	tx.Exec("INSERT INTO (a, b, c, d) VALUES ($1, $2, $3, $4)", 1, 2, 3, 4)
}
```

Again, interpolation seems faster with less allocations. The underlying driver
still has to process and send the arguments with the prepared statement name.
*I expected database/sql to better interpolation here. Still thinking
about this one.*

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
sql, args := builder.MustInterpolate()
if len(args) == 0 {
    rows, err = db.Query(sql)
} else {
    rows, err = db.Query(sql, args...)
}
```

## Running Tests and Benchmarks

Run the following inside project root

```sh
# install godo task runner
go get -u gopkg.in/godo.v1/cmd/godo

# install dependencies
cd tasks
go get -a

# back to root and run
cd ..

# create database
godo createdb

# run tests
godo test

# run benchmarks
godo bench
```

When createdb prompts for superuser, enter superuser like 'postgres' to create
the test database. On Mac + Postgress.app user your user name.

## TODO

* more tests
* hstore query suppport
* stored procedure support

## Inspiration

*   [mapper](https://github.com/mgutz/mapper)

    My SQL builder for node.js which has builder, interpolation and exec
    functionality.

*   [dbr](https://github.com/gocraft/dbr)

    used this as starting point instead of porting mapper from scratch
