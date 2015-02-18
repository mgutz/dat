package runner

import "database/sql"

// Tx is a transaction for the given Session
type Tx struct {
	*sql.Tx
	Runner
}

// Begin creates a transaction for the given session
func (sess *Session) Begin() (*Tx, error) {
	tx, err := sess.cxn.Db.Begin()
	if err != nil {
		return nil, events.EventErr("begin.error", err)
	}
	events.Event("begin")

	return &Tx{
		Tx:     tx,
		Runner: Runner{tx},
	}, nil
}

// Commit finishes the transaction
func (tx *Tx) Commit() error {
	err := tx.Tx.Commit()
	if err != nil {
		return events.EventErr("commit.error", err)
	}
	events.Event("commit")
	return nil
}

// Rollback cancels the transaction
func (tx *Tx) Rollback() error {
	err := tx.Tx.Rollback()
	if err != nil {
		return events.EventErr("rollback", err)
	}
	events.Event("rollback")
	return nil
}

// RollbackUnlessCommitted rollsback the transaction unless it has already been committed or rolled back.
// Useful to defer tx.RollbackUnlessCommitted() -- so you don't have to handle N failure cases
// Keep in mind the only way to detect an error on the rollback is via the event log.
func (tx *Tx) RollbackUnlessCommitted() {
	err := tx.Tx.Rollback()
	if err == sql.ErrTxDone {
		// ok
	} else if err != nil {
		events.EventErr("rollback_unless_committed", err)
	} else {
		events.Event("rollback")
	}
}
