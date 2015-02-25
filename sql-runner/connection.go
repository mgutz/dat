package runner

import (
	"database/sql"

	"github.com/mgutz/dat"
)

// Connection is a connection to the database with an EventReceiver
type Connection struct {
	DB *sql.DB
	*Queryable
}

// NewConnection instantiates a Connection for a given database/sql connection
func NewConnection(db *sql.DB) *Connection {
	return &Connection{db, &Queryable{db}}
}

// ExecMulti executes group SQL statemetns in a string marked by a marker.
// The deault marker is "GO"
func (conn *Connection) ExecMulti(sql string) error {
	statements, err := dat.SQLSliceFromString(sql)
	if err != nil {
		return err
	}
	for _, sq := range statements {
		_, err := conn.DB.Exec(sq)
		if err != nil {
			return err
		}
	}
	return nil
}
