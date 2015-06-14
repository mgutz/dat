** The [legacy](https://github.com/mgutz/dat/tree/legacy) branch contains
the original dat **

# dat

[GoDoc](https://godoc.org/github.com/mgutz/dat)

`dat` (Data Access Toolkit) is a fast, lightweight and intuitive Postgres
library for Go.

How it is different than other toolkits:

*   Focused on Postgres. See `Insect`, `Upsert`, `SelectDoc`, `QueryJSON`.

*   Light layer over [sqlx](https://github.com/jmoiron/sqlx)

*   SQL and backtick friendly.

    ```go
    DB.SQL(`SELECT * FROM people LIMIT 10`).QueryStructs(&people)
    ```

*   Intuitive JSON Document retrieval (single trip to database!)

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

*   Simpler JSON retrieval from database (requires Postgres 9.3+)

    ```go
    var json []byte
    json, _ = DB.SQL(`SELECT id, user_name, created_at FROM users WHERE user_name = $1 `,
        "mario",
    ).QueryJSON()

    // straight into map
    var obj map[string]interface{}
    DB.SQL(`SELECT id, user_name, created_at FROM users WHERE user_name = $1 `,
        "mario",
    ).QueryObject(&obj)
    ```

    both result in

    ```json
    {
        "id": 1,
        "user_name": "mario",
        "created_at": "2015-03-01T14:23"
    }
    ```

*   Ordinal placeholders - friendlier than `?`

    ```go
    DB.SQL(`SELECT * FROM people WHERE state = $1`, "CA").Exec()
    ```

*   Minimal API Surface. No AST-like language to learn.

    ```go
    err := DB.
        Select("id, user_name").
        From("users").
        Where("id = $1", id).
        QueryStruct(&user)
    ```

*   Performant

    -   ordinal placeholder logic is optimized to be nearly as fast as using `?`
    -   `dat` can interpolate queries locally resulting in performance increase
        over plain database/sql and sqlx. [Benchmarks](https://github.com/mgutz/dat/wiki/Benchmarks)

## Getting Started

Get it

```sh
go get -u gopkg.in/mgutz/dat.v1
```

Use it

```go
import (
    "database/sql"

    _ "github.com/lib/pq"
    "gopkg.in/mgutz/data.v1"
    "gopkg.in/mgutz/data.v1/sqlx-runner"
)

// global database (pooling provided by SQL driver)
var DB *runner.DB

func init() {
    // create a normal database connection through database/sql
    db, err := sql.Open("postgres", "dbname=dat_test user=dat password=!test host=localhost sslmode=disable")
    if err != nil {
        panic(err)
    }

    // set this to enable interpolation
    dat.EnableInterpolation = true
    // set to check things like sessions closing.
    // Should be disabled in production/release builds.
    dat.Strict = false
    DB = runner.NewDB(db, "postgres")
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
library exists is my dissatisfaction with other SQL builders introducing
another domain language or AST-like expressions.

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

### Runners

`dat` was designed to have clear separation between SQL builders and Query execers.
This is why the runner is in its own package.

*   `sqlx-runner` - based on [sqlx](https://github.com/jmoiron/sqlx)


### Tracing SQL

`dat` uses [logxi](https://github.com/mgutz/logxi) for logging. To trace SQL
set environment variable

    LOGXI=dat* yourapp

## CRUD

### Create

Use `Returning` and `QueryStruct` to insert and update struct fields in one
trip

```go
post := Post{Title: "Swith to Postgres", State: "open"}

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
    Record(post).
    Returning("id", "created_at", "updated_at").
    QueryStruct(&post)

// use wildcard to include all columns
err := DB.
    InsertInto("posts").
    Whitelist("*").
    Record(post).
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

published := dat.NewScope(
    "WHERE user_id = :userID AND state = 'published'",
    dat.M{"userID": 0},
)

var posts []*Post
err = DB.
    Select("id, title").
    From("posts").
    ScopeMap(published, dat.M{"userID": 100})
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
    Limit(1).
    Exec()
```

### Scopes

Scopes predefine JOIN and WHERE conditions so they may be reused.
Scopes may be used with `DeleteFrom`, `Select` and `Update`.

As an example, a "published" scoped might define published posts
by user. The definition might look something like this with
joins

```go
// :TABLE is the table name of the builder to which this scope is applied.
publishedByUser := `
    INNER JOIN users U on (:TABLE.user_id = U.id)
    WHERE
        :TABLE.state = 'published' AND
        :TABLE.deleted_at IS NULL AND
        U.user_name = $1
`

err = DB.
    Select("posts.*").                  // must qualify columns
    From("posts").
    Scope(publishedByUser, "mgutz").
    QueryStructs(&posts)
```

If you need to predefine values for parameters, then use a MapScope.

```go
// creates a MapScope
publishedByUser := dat.NewScope(`
    INNER JOIN users U on (:TABLE.user_id = U.id)
    WHERE
        :TABLE.state = :state AND
        :TABLE.deleted_at IS NULL AND
        U.user_name = :user`,
    dat.M{"user": "unknown", "state": "published"},
)
```

First, it does not use ordinal placeholders. Instead it uses struct field
names in the SQL. The example above defines default values for fields `"user"`
and `"state"`. When the scope is applied, the scope is first cloned then
new values replace default values.

```go
err = DB.
    Select("posts.*").
    From("posts").
    ScopeMap(publishedByUser, dat.M{"user": "mgutz"}).
    QueryStructs(&posts)
```

`MapScope` provides flexibility but it inefficient compared to the first
approach.



## Creating Connections

All queries are made in the context of a connection which are acquired
from the underlying SQL driver's pool

For one-off operations, use `DB` directly

```go
// a global connection usually created in `init`
DB := runner.NewDB(db, "postgres")

err := DB.SQL(...).QueryStruct(&post)
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

`DB` and `Tx` implement `runner.Connection` interface

```
func getUsers(conn runner.Connection) ([]*dto.Users, error) {
    sql := `
        SELECT *
        FROM users
    `
    var users dto.Users
    err := conn.SQL(sql).QueryStructs(&users)
    if err != nil {
        return err
    }
    return users
}
```

### Dates

Dates are a pain in go. Use `dat.NullTime` type to properly
handle nullable dates from JSON and Postgres.

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
constant and **NEVER** use `UnsafeString` directly as an argument. This
is asking for a SQL injection attack

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
err := DB.
    Select("id, title, user_id").
    From("posts").
    Limit(1).
    QueryStruct(&post)
```

### Transactions

```go
// Start transaction
tx, err := DB.Begin()
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

#### Nested Transactions

Simple nested transaction support works as follows:

*   If `Commit` is called in a nested transaction, the operation results in a NOOP.
    Only the top level `Commit` commits a transaction to the database.

*   If `Rollback` is called in a nested transaction, then the entire
    transaction is rollbacked. `Tx.IsRollbacked` is set to true.

*   Either `defer Tx.AutoCommit()` or `defer Tx.AutoRollback()` **MUST BE CALLED**
    for each corresponding `Begin`. The internal state of nested transaction is
    cleaned up in these two methods.

```go
func nested(conn runner.Connection) error {
    tx, err := conn.Begin()
    if err != nil {
        return err
    }
    defer tx.AutoRollback()

    _, err := tx.SQL(`INSERT INTO users (email) values $1`, 'me@home.com').Exec()
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

### Local Interpolation

__Interpolation is DISABLED by default. Set `dat.EnableInterpolation = true`
to enable.__

`dat` can interpolate locally to inline query arguments. Let's start with a
normal SQL statements with arguments

```
db.Exec(
    "INSERT INTO (a, b, c, d) VALUES ($1, $2, $3, $4)",
    []interface{}[1, 2, 3, 4],
)
```

When the statement agove gets executed:

1. The driver checks if this SQL has been prepared previously on the current connection, using the SQL as the key
1. If not, the driver sends the SQL statement to the database to prepare the statement
2. The prepared statement is assigned to the connection
3. The prepared satement is executed along with arguments
4. Received data is sent back to the caller

In contrast, `dat` can interpolate the statement locally resulting in
a SQL statement with often no arguments. The code above results in
this interpolated SQL

```
"INSERT INTO (a, b, c, d) VALUES (1, 2, 3, 4)"
```

When the statement agove gets executed:

1. The statement is treated as simple exec and sent with args to database, since len(args) == 0
2. Received data is sent back to the caller

#### Interpolation Safety

As of Postgres 9.1, the database does not process escape sequence by default. See
[String Constants with C-style Escapes](http://www.postgresql.org/docs/current/interactive/sql-syntax-lexical.html#SQL-SYNTAX-STRINGS-ESCAPE).
In short, all backslashes are treated literally.

`dat` escapes single quotes (apostrophes) on
small strings, otherwise it uses Postgres' [dollar
quotes](http://www.postgresql.org/docs/current/interactive/sql-syntax-lexical.html#SQL-SYNTAX-DOLLAR-QUOTING)
to escape strings. The dollar quote tag is randomized at init. If a string contains the
dollar quote tag, the tag is randomized again and if the string still contains the tag, then
single quote escaping is used.

As an added safety measure, `dat` checks the Postgres database
`standard_conforming_strings` setting value on a new connection when
`dat.EnableInterpolation == true`. If `standard_conforming_strings != "on"` then set set it to `"on"`
or disable interpolation. `dat` will panic if it the setting is incompatible.

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

Keep in mind that prepared statements are only valid for the current session
and unless the same query is be executed *MANY* times in the same session there
is little benefit in using prepared statements other than protecting against
SQL injections. See Interpolation Safety section above.


#### More Reasons to Use Interpolation

*   Performance improvement
*   Debugging is simpler with interpolated SQL
*   Use SQL constants like `NOW` and `DEFAULT`
*   Expand placeholders with expanded slice values `$1 => (1, 2, 3)`

`[]byte`,  `[]*byte` and any unhandled values are passed through to the
driver when interpolating.

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
sql, args, err := builder.Interpolate()
if len(args) == 0 {
    rows, err = db.Query(sql)
} else {
    rows, err = db.Query(sql, args...)
}
```

## Running Tests and Benchmarks

To setup the task runner and create database

```sh
# install godo task runner
go get -u gopkg.in/godo.v2/cmd/godo

# install dependencies
cd tasks
go get -a

# back to project
cd ..
```

Then run any task

```sh
# (re)create database
godo createdb

# run tests with traced SQL (optional)
LOGXI=dat* godo test

# run tests without logs
godo test

# run benchmarks
godo bench

# see other tasks
godo
```

When createdb prompts for superuser, enter superuser like 'postgres' to create
the test database. On Mac + Postgress.app use your own user name and password.

