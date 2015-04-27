package runner

// Session represents a business unit of execution for some connection
type Session struct {
	*Tx
}

// NewSession instantiates a Session for the Connection
func (cxn *Connection) NewSession() (*Session, error) {
	tx, err := cxn.Begin()
	if err != nil {
		return nil, err
	}
	return &Session{tx}, nil
}

// Close closes the session.
func (sess *Session) Close() error {
	err := sess.Tx.AutoCommit()
	if err != nil {
		logger.Error("session.close", "err", err)
	}
	return err
}
