package runner

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mgutz/dat/v1"
)

const (
	txPending = iota
	txCommitted
	txRollbacked
)

func logError(msg string, err error) error {
	logger.Error(msg, "err", err)
	return err
}

// Tx is a transaction for the given Session
type Tx struct {
	sync.Mutex
	*sqlx.Tx
	*Queryable
	state int
}

// WrapSqlxTx creates a Tx from a sqlx.Tx
func WrapSqlxTx(tx *sqlx.Tx) *Tx {
	newtx := &Tx{Tx: tx, Queryable: &Queryable{tx}}
	if dat.Strict {
		time.AfterFunc(1*time.Minute, func() {
			if newtx.state == txPending {
				panic("A database session was not closed!")
			}
		})
	}
	return newtx
}

// Begin creates a transaction for the given session
func (conn *Connection) Begin() (*Tx, error) {
	tx, err := conn.DB.Beginx()
	if err != nil {
		return nil, logError("begin.error", err)
	}
	logger.Debug("begin")
	return WrapSqlxTx(tx), nil
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
		return logError("commit.error", err)
	}
	logger.Debug("commit")
	tx.state = txCommitted
	return nil
}

// Rollback cancels the transaction
func (tx *Tx) Rollback() error {
	tx.Lock()
	defer tx.Unlock()

	if tx.state == txCommitted {
		return logError("rollback", fmt.Errorf("Cannot rollback, transaction has already been commited"))
	}
	if tx.state == txRollbacked {
		return logError("rollback", fmt.Errorf("Cannot rollback, transaction has already been rollbacked"))
	}
	err := tx.Tx.Rollback()
	if err != nil {
		return logError("rollback", err)
	}
	logger.Debug("rollback")
	tx.state = txRollbacked
	return nil
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
		return logError("transaction.AutoCommit.commit_error", err)
	}
	logger.Debug("autocommit")
	tx.state = txCommitted
	return err
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
		return logError("transaction.AutoRollback.rollback_error", err)
	}
	logger.Debug("autorollback")
	tx.state = txRollbacked
	return err
}

// Select creates a new SelectBuilder for the given columns.
// This disambiguates between Queryable.Select and sqlx's Select
func (tx *Tx) Select(columns ...string) *dat.SelectBuilder {
	return tx.Queryable.Select(columns...)
}
