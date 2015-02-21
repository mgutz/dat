package runner

import "github.com/jmoiron/sqlx"

// Session represents a business unit of execution for some connection
type Session struct {
	DB *sqlx.DB
	*Queryable
}

// NewSession instantiates a Session for the Connection
func (cxn *Connection) NewSession() *Session {
	return &Session{cxn.DB, &Queryable{cxn.DB}}
}
