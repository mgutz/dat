package runner

import (
	// "database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransactionReal(t *testing.T) {
	installFixtures()

	tx, err := conn.Begin()
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

	tx, err := conn.Begin()
	assert.NoError(t, err)

	var person Person
	err = tx.Select("*").From("people").Where("email = $1", "john@acme.com").QueryStruct(&person)
	assert.NoError(t, err)
	assert.Equal(t, person.Name, "John")

	err = tx.Rollback()
	assert.NoError(t, err)
}
