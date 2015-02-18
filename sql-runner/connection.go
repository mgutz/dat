package runner

import "database/sql"

// Connection is a connection to the database with an EventReceiver
// to send events, errors, and timings to
type Connection struct {
	Db *sql.DB
}

// NewConnection instantiates a Connection for a given database/sql connection
// and event receiver
func NewConnection(db *sql.DB) *Connection {
	return &Connection{Db: db}
}
