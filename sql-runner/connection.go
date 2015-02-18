package runner

import "database/sql"

// Connection is a connection to the database with an EventReceiver
type Connection struct {
	Db *sql.DB
	Runner
}

// NewConnection instantiates a Connection for a given database/sql connection
func NewConnection(db *sql.DB) *Connection {
	return &Connection{Db: db, Runner: Runner{db}}
}
