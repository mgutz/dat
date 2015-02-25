package runner

import (
	"bytes"
	"fmt"
	"log"

	"github.com/mgutz/dat"
	"gopkg.in/jackc/pgx.v2"
)

var conn *Connection

func init() {
	conn = NewConnection("")
	dat.SetVerbose(false)
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

type Person struct {
	ID        int32           `db:"id"`
	Amount    pgx.NullFloat64 `db:"amount"`
	Doc       pgx.NullString  `db:"doc"`
	Email     pgx.NullString  `db:"email"`
	Foo       string          `db:"foo"`
	Image     []byte          `db:"image"`
	Key       pgx.NullString  `db:"key"`
	Name      string          `db:"name"`
	CreatedAt pgx.NullTime    `db:"created_at"`
}

func installFixtures() {
	db := conn.DB
	createTablePeople := `
		CREATE TABLE people (
			id SERIAL PRIMARY KEY,
			amount float8,
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
			log.Fatalln("Failed to execute statement: ", v, " Got error: ", err)
		}
	}
}
