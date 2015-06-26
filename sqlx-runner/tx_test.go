package runner

import (
	// "database/sql"
	"database/sql"
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"

	"github.com/mgutz/logxi/v1"
)

func TestTransactionReal(t *testing.T) {
	installFixtures()

	tx, err := testDB.Begin()
	assert.NoError(t, err)

	var id int64
	tx.InsertInto("people").Columns("name", "email").
		Values("Barack", "obama@whitehouse.gov").
		Returning("id").
		QueryScalar(&id)

	assert.True(t, id > 0)

	var person Person
	err = tx.
		Select("*").
		From("people").
		Where("id = $1", id).
		QueryStruct(&person)
	assert.NoError(t, err)

	assert.Equal(t, person.ID, id)
	assert.Equal(t, person.Name, "Barack")
	assert.Equal(t, person.Email.Valid, true)
	assert.Equal(t, person.Email.String, "obama@whitehouse.gov")

	err = tx.Commit()
	assert.NoError(t, err)
}

func TestTransactionRollbackReal(t *testing.T) {
	installFixtures()

	tx, err := testDB.Begin()
	assert.NoError(t, err)

	var person Person
	err = tx.Select("*").From("people").Where("email = $1", "john@acme.com").QueryStruct(&person)
	assert.NoError(t, err)
	assert.Equal(t, person.Name, "John")

	err = tx.Rollback()
	assert.NoError(t, err)
}

func nestedCommit(c Connection) error {
	tx, err := c.Begin()
	if err != nil {
		return err
	}
	defer tx.AutoRollback()

	// this will commit
	var id int64
	tx.InsertInto("people").Columns("name", "email").
		Values("Mario", "mario@mgutz.com").
		Returning("id").
		QueryScalar(&id)
	return tx.Commit()
}

func nestedNestedCommit(c Connection) error {
	tx, err := c.Begin()
	if err != nil {
		return err
	}
	defer tx.AutoRollback()

	// this will commit
	var id int64
	tx.InsertInto("people").Columns("name", "email").
		Values("Mario2", "mario2@mgutz.com").
		Returning("id").
		QueryScalar(&id)
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nestedCommit(tx)
}

func nestedRollback(c Connection) error {
	tx, err := c.Begin()
	if err != nil {
		return err
	}
	defer tx.AutoRollback()

	// this will commit
	var id int64
	tx.InsertInto("people").Columns("name", "email").
		Values("Mario", "mario@mgutz.com").
		Returning("id").
		QueryScalar(&id)
	return nil
}

func nestedNestedRollback(c Connection) error {
	tx, err := c.Begin()
	if err != nil {
		return err
	}
	defer tx.AutoRollback()

	// this will commit
	var id int64
	tx.InsertInto("people").Columns("name", "email").
		Values("Mario", "mario@mgutz.com").
		Returning("id").
		QueryScalar(&id)
	return nestedRollback(tx)
}

func TestRollbackWithNestedCommit(t *testing.T) {
	log.Suppress(true)
	defer log.Suppress(false)
	installFixtures()
	tx, err := testDB.Begin()
	assert.NoError(t, err)
	err = nestedCommit(tx)
	assert.NoError(t, err)
	err = tx.Rollback()
	assert.NoError(t, err)

	var person Person
	err = testDB.
		Select("*").
		From("people").
		Where("email = $1", "mario@mgutz.com").
		QueryStruct(&person)
	assert.Exactly(t, sql.ErrNoRows, err)
	assert.EqualValues(t, 0, person.ID)
}

func TestCommitWithNestedCommit(t *testing.T) {
	log.Suppress(true)
	defer log.Suppress(false)
	installFixtures()
	tx, err := testDB.Begin()
	assert.NoError(t, err)
	err = nestedCommit(tx)
	assert.NoError(t, err)
	err = tx.Commit()
	assert.NoError(t, err)

	var person Person
	err = testDB.
		Select("*").
		From("people").
		Where("email = $1", "mario@mgutz.com").
		QueryStruct(&person)
	assert.NoError(t, err)
	assert.True(t, person.ID > 0)
}

func TestCommitWithNestedNestedCommit(t *testing.T) {
	log.Suppress(true)
	defer log.Suppress(false)
	installFixtures()
	tx, err := testDB.Begin()
	assert.NoError(t, err)
	err = nestedNestedCommit(tx)
	assert.NoError(t, err)
	err = tx.Commit()
	assert.NoError(t, err)

	var person Person
	err = testDB.
		Select("*").
		From("people").
		Where("email = $1", "mario@mgutz.com").
		QueryStruct(&person)
	assert.NoError(t, err)
	assert.True(t, person.ID > 0)

	err = testDB.
		Select("*").
		From("people").
		Where("email = $1", "mario2@mgutz.com").
		QueryStruct(&person)
	assert.NoError(t, err)
	assert.True(t, person.ID > 0)
}

func TestRollbackWithNestedRollback(t *testing.T) {
	log.Suppress(true)
	defer log.Suppress(false)
	installFixtures()
	tx, err := testDB.Begin()
	assert.NoError(t, err)
	err = nestedRollback(tx)
	assert.NoError(t, err)
	err = tx.Rollback()
	assert.Exactly(t, ErrTxRollbacked, err)

	var person Person
	err = testDB.
		Select("*").
		From("people").
		Where("email = $1", "mario@mgutz.com").
		QueryStruct(&person)
	assert.Exactly(t, sql.ErrNoRows, err)
}

func TestCommitWithNestedRollback(t *testing.T) {
	log.Suppress(true)
	defer log.Suppress(false)
	installFixtures()
	tx, err := testDB.Begin()
	assert.NoError(t, err)
	err = nestedRollback(tx)
	assert.NoError(t, err)
	err = tx.Commit()
	assert.Exactly(t, ErrTxRollbacked, err)

	var person Person
	err = testDB.
		Select("*").
		From("people").
		Where("email = $1", "mario@mgutz.com").
		QueryStruct(&person)
	assert.Exactly(t, sql.ErrNoRows, err)
}

func TestCommitWithNestedNestedRollback(t *testing.T) {
	log.Suppress(true)
	defer log.Suppress(false)
	installFixtures()
	tx, err := testDB.Begin()
	assert.NoError(t, err)
	err = nestedNestedRollback(tx)
	assert.NoError(t, err)
	err = tx.Commit()
	assert.Exactly(t, ErrTxRollbacked, err)

	var person Person
	err = testDB.
		Select("*").
		From("people").
		Where("email = $1", "mario@mgutz.com").
		QueryStruct(&person)
	assert.Exactly(t, sql.ErrNoRows, err)
}

func TestErrorInBeginIfRollbacked(t *testing.T) {
	log.Suppress(true)
	defer log.Suppress(false)
	installFixtures()
	tx, err := testDB.Begin()
	assert.NoError(t, err)
	err = tx.Rollback()
	assert.NoError(t, err)

	_, err = tx.Begin()
	assert.Exactly(t, ErrTxRollbacked, err)
}
