package runner

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"

	"github.com/mgutz/dat"
)

var conn *Connection
var db *sql.DB

func init() {
	db = realDb()
	conn = NewConnection(db, "postgres")
	dat.Strict = false
}

func createRealSession() *Session {
	sess, err := conn.NewSession()
	if err != nil {
		panic(err)
	}
	return sess
}

func createRealSessionWithFixtures() *Session {
	installFixtures()
	sess := createRealSession()
	return sess
}

func quoteColumn(column string) string {
	var buffer bytes.Buffer
	dat.Dialect.WriteIdentifier(&buffer, column)
	return buffer.String()
}

func quoteSQL(sqlFmt string, cols ...string) string {
	args := make([]interface{}, len(cols))

	for i := range cols {
		args[i] = quoteColumn(cols[i])
	}

	return fmt.Sprintf(sqlFmt, args...)
}

func realDb() *sql.DB {
	driver := os.Getenv("DAT_DRIVER")
	if driver == "" {
		logger.Fatal("env DAT_DRIVER is not set")
	}

	dsn := os.Getenv("DAT_DSN")
	if dsn == "" {
		logger.Fatal("env DAT_DSN is not set")
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		logger.Fatal("Database error ", "err", err)
	}

	return db
}

type Person struct {
	ID        int64           `db:"id"`
	Amount    dat.NullFloat64 `db:"amount"`
	Doc       dat.NullString  `db:"doc"`
	Email     dat.NullString  `db:"email"`
	Foo       string          `db:"foo"`
	Image     []byte          `db:"image"`
	Key       dat.NullString  `db:"key"`
	Name      string          `db:"name"`
	CreatedAt dat.NullTime    `db:"created_at"`
}

func installFixtures() {
	db := conn.DB
	createTablePeople := `
		CREATE TABLE people (
			id SERIAL PRIMARY KEY,
			amount money,
			doc hstore,
			email text,
			foo text default 'bar',
			image bytea,
			key text,
			name text NOT NULL,
			created_at timestamptz default now()
		)
	`

	sqlToRun := []string{
		"DROP TABLE IF EXISTS people",
		createTablePeople,
		"INSERT INTO people (name,email) VALUES ('Jonathan', 'jonathan@acme.com')",
		"INSERT INTO people (name,email) VALUES ('Dmitri', 'zavorotni@jadius.com')",
	}

	for _, v := range sqlToRun {
		_, err := db.Exec(v)
		if err != nil {
			logger.Fatal("Failed to execute statement", "sql", v, "err", err)
		}
	}
}
