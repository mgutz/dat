package runner

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mgutz/dat"
)

const (
	txPending = iota
	txCommitted
	txRollbacked
)

// Tx is a transaction for the given Session
type Tx struct {
	sync.Mutex
	*sqlx.Tx
	*Queryable
	state int
}

// Begin creates a transaction for the given session
func (cxn *Connection) Begin() (*Tx, error) {
	tx, err := cxn.DB.Beginx()
	if err != nil {
		return nil, dat.Events.EventErr("begin.error", err)
	}
	dat.Events.Event("begin")

	newtx, err := &Tx{Tx: tx, Queryable: &Queryable{tx}}, nil
	if dat.Strict {
		time.AfterFunc(1*time.Minute, func() {
			if newtx.state == txPending {
				panic("A database session was not closed!")
			}
		})
	}
	return newtx, err
}

// Commit finishes the transaction
func (tx *Tx) Commit() error {
	tx.Lock()
	defer tx.Unlock()

	if tx.state == txCommitted || tx.state == txRollbacked {
		return nil
	}
	err := tx.Tx.Commit()
	if err != nil {
		return dat.Events.EventErr("commit.error", err)
	}
	dat.Events.Event("commit")
	return nil
}

// Rollback cancels the transaction
func (tx *Tx) Rollback() error {
	tx.Lock()
	defer tx.Unlock()

	if tx.state == txCommitted {
		return dat.Events.EventErr("rollback", fmt.Errorf("Cannot rollback, transaction has already been commited"))
	}
	if tx.state == txRollbacked {
		return dat.Events.EventErr("rollback", fmt.Errorf("Cannot rollback, transaction has already been rollbacked"))
	}
	err := tx.Tx.Rollback()
	if err != nil {
		return dat.Events.EventErr("rollback", err)
	}
	tx.state = txRollbacked
	dat.Events.Event("rollback")
	return nil
}

// RollbackUnlessCommitted rollsback the transaction unless it has already been committed or rolled back.
// Useful to defer tx.RollbackUnlessCommitted() -- so you don't have to handle N failure cases
// Keep in mind the only way to detect an error on the rollback is via the event log.
func (tx *Tx) RollbackUnlessCommitted() {
	panic("RollbackUnlessCommitted has been obsoleted. Use AutoCommit")
	// err := tx.Tx.Rollback()
	// if err == sql.ErrTxDone {
	// 	// ok
	// } else if err != nil {
	// 	dat.Events.EventErr("rollback_unless_committed", err)
	// } else {
	// 	dat.Events.Event("rollback")
	// }
}

// AutoCommit closes a session.
func (tx *Tx) AutoCommit() error {
	tx.Lock()
	defer tx.Unlock()

	if tx.state == txRollbacked || tx.state == txCommitted {
		return nil
	}
	err := tx.Tx.Commit()
	if err != nil {
		if dat.Strict {
			log.Fatalf("Could not close session: %s\n", err.Error())
		}
		return dat.Events.EventErr("transaction.AutoCommit.commit_error", err)
	}
	dat.Events.Event("autocommit")
	return err
}

// Select creates a new SelectBuilder for the given columns.
// This disambiguates between Queryable.Select and sqlx's Select
func (tx *Tx) Select(columns ...string) *dat.SelectBuilder {
	return tx.Queryable.Select(columns...)
}
