package runner

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// Connection is a connection to the database with an EventReceiver
type Connection struct {
	DB *sqlx.DB
	*Queryable
}

// NewConnection instantiates a Connection for a given database/sql connection
func NewConnection(db *sql.DB, driverName string) *Connection {
	DB := sqlx.NewDb(db, driverName)
	return &Connection{DB, &Queryable{DB}}
}
