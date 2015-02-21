package runner

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/mgutz/dat"
)

//
// Test helpers
//
var testConn *Connection

func createRealSession() *Session {
	testConn = NewConnection(realDb())
	return testConn.NewSession()
}

func createRealSessionWithFixtures() *Session {
	sess := createRealSession()
	installFixtures(sess.cxn.Db)
	return sess
}

func quoteColumn(column string) string {
	var buffer bytes.Buffer
	dat.Quoter.WriteQuotedColumn(column, &buffer)
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
	driver := os.Getenv("DBR_DRIVER")
	if driver == "" {
		log.Fatalln("env DBR_DRIVER is not set")
	}

	dsn := os.Getenv("DBR_DSN")
	if dsn == "" {
		log.Fatalln("env DBR_DSN is not set")
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		log.Fatalln("Database error ", err)
	}

	return db
}

type Person struct {
	ID        int64 `db:"id"`
	Name      string
	Foo       string
	Email     dat.NullString
	Key       dat.NullString
	Doc       dat.NullString
	CreatedAt dat.NullTime
}

func installFixtures(db *sql.DB) {
	createTablePeople := `
		CREATE TABLE people (
			id SERIAL PRIMARY KEY,
			name varchar(255) NOT NULL,
			email varchar(255),
			key varchar(255),
			doc hstore,
			foo varchar(255) default 'bar',
			created_at timestamptz default now()
		)
	`

	sqlToRun := []string{
		"DROP TABLE IF EXISTS people",
		createTablePeople,
		"INSERT INTO people (name,email) VALUES ('Jonathan', 'jonathan@uservoice.com')",
		"INSERT INTO people (name,email) VALUES ('Dmitri', 'zavorotni@jadius.com')",
	}

	for _, v := range sqlToRun {
		_, err := db.Exec(v)
		if err != nil {
			log.Fatalln("Failed to execute statement: ", v, " Got error: ", err)
		}
	}
}
