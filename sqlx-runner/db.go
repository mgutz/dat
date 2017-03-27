package runner

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"gopkg.in/mgutz/dat.v1"
)

// DB represents an abstract database connection pool.
type DB struct {
	DB *sqlx.DB
	*Queryable
	Version int64
}

var standardConformingStrings string

// pgMustNotAllowEscapeSequence checks if Postgres treats backlashes
// literally in strings when dat.EnableInterpolation == true. If escape
// sequences are allowed, then it is unsafe to use interpolation and
// this function panics.
func pgMustNotAllowEscapeSequence(conn *DB) {
	if !dat.EnableInterpolation {
		return
	}

	if standardConformingStrings == "" {
		err := conn.
			SQL("select setting from pg_settings where name='standard_conforming_strings'").
			QueryScalar(&standardConformingStrings)
		if err != nil {
			panic(err)
		}
	}

	if standardConformingStrings != "on" {
		logger.Fatal("Database allows escape sequences. Cannot be used with interpolation. "+
			"standard_conforming_strings=%q\n"+
			"See http://www.postgresql.org/docs/9.3/interactive/sql-syntax-lexical.html#SQL-SYNTAX-STRINGS-ESCAPE",
			"standardConformingStrings", standardConformingStrings)
	}
}

func pgSetVersion(db *DB) {
	err := db.
		SQL("SHOW server_version_num").
		QueryScalar(&db.Version)
	if err != nil {
		logger.Fatal("Could not query Postgres version")
		return
	}
}

// NewDB instantiates a Connection for a given database/sql connection
func NewDB(db *sql.DB, driverName string) *DB {
	database := sqlx.NewDb(db, driverName)
	conn := &DB{DB: database, Queryable: &Queryable{database}}
	if driverName == "postgres" {
		pgMustNotAllowEscapeSequence(conn)
		pgSetVersion(conn)
		if dat.Strict {
			conn.SQL("SET client_min_messages to 'DEBUG';")
		}
	} else {
		panic("Unsupported driver: " + driverName)
	}
	return conn
}

// NewDBFromString instantiates a Connection from a given driver
// and connection string.
func NewDBFromString(driver string, connectionString string) *DB {
	db, err := sql.Open(driver, connectionString)
	if err != nil {
		logger.Fatal("Database error ", "err", err)
	}
	err = db.Ping()
	if err != nil {
		logger.Fatal("Could not ping database", "err", err)
	}
	return NewDB(db, driver)
}

// NewDBFromSqlx creates a new Connection object from existing Sqlx.DB.
func NewDBFromSqlx(dbx *sqlx.DB) *DB {
	conn := &DB{DB: dbx, Queryable: &Queryable{dbx}}
	pgMustNotAllowEscapeSequence(conn)
	pgSetVersion(conn)
	return conn
}
