package runner

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

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
	*sql.Tx
	*Queryable
	state int
}

// Begin creates a transaction for the given session
func (conn *Connection) Begin() (*Tx, error) {
	tx, err := conn.DB.Begin()
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

// AutoRollback rolls back transaction IF neither Commit or Rollback were called.
func (tx *Tx) AutoRollback() error {
	tx.Lock()
	defer tx.Unlock()

	if tx.state == txRollbacked || tx.state == txCommitted {
		return nil
	}
	err := tx.Tx.Rollback()
	if err != nil {
		if dat.Strict {
			log.Fatalf("Could not rollback session: %s\n", err.Error())
		}
		return dat.Events.EventErr("transaction.AutoRollback.rollback_error", err)
	}
	dat.Events.Event("autorollback")
	return err
}

// AutoCommit commits a transaction IF neither Commit or Rollback were called.
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
