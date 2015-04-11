package runner

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"

	"github.com/mgutz/dat/v1"
	"github.com/mgutz/dat/v1/postgres"
)

var conn *Connection
var db *sql.DB

func init() {
	dat.Dialect = postgres.New()
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
