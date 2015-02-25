package runner

import (
	"github.com/mgutz/dat"
	"gopkg.in/jackc/pgx.v2"
)

// Connection is a connection to the database with an EventReceiver
type Connection struct {
	DB *pgx.ConnPool
	*Queryable
}

// NewConnection instantiates a Connection for a given database/sql connection
func NewConnection(dsn string) *Connection {
	config := pgx.ConnPoolConfig{}
	config.Database = "dbr_test"
	config.User = "dbr"
	config.Password = "!test"
	config.Host = "localhost"
	//DAT_DSN="dbname=dbr_test user=dbr password=!test host=localhost sslmode=disable"

	db, err := pgx.NewConnPool(config)
	if err != nil {
		panic(err)
	}

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
