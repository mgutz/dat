package runner

// Session represents a business unit of execution for some connection
type Session struct {
	cxn *Connection
	Runner
}

// NewSession instantiates a Session for the Connection
func (cxn *Connection) NewSession() *Session {
	return &Session{cxn: cxn, Runner: Runner{cxn.Db}}
}
