package runner

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/mgutz/logxi"

	"gopkg.in/mgutz/dat.v1"
	"gopkg.in/mgutz/dat.v1/kvs"
	"gopkg.in/mgutz/dat.v1/postgres"
)

var testDB *DB
var sqlDB *sql.DB

func init() {
	dat.Dialect = postgres.New()
	sqlDB = realDb()
	testDB = NewDB(sqlDB, "postgres")
	dat.Strict = false
	logxi.Suppress(true)

	Cache = kvs.NewMemoryKeyValueStore(1 * time.Second)
	//Cache, _ = kvs.NewDefaultRedisStore()
}

func beginTxWithFixtures() *Tx {
	installFixtures()
	c, err := testDB.Begin()
	if err != nil {
		panic(err)
	}
	return c
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
