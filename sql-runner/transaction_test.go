package runner

import (
	// "database/sql"
	"testing"

	"github.com/mgutz/dat"
	"github.com/stretchr/testify/assert"
)

func TestTransactionReal(t *testing.T) {
	s := createRealSessionWithFixtures()

	tx, err := s.Begin()
	assert.NoError(t, err)

	var id int64
	b := dat.InsertInto("dbr_people").Columns("name", "email").
		Values("Barack", "obama@whitehouse.gov").
		Returning("id")

	tx.QueryScan(b, &id)

	assert.True(t, id > 0)

	var person dbrPerson
	err = tx.QueryStruct(dat.Select("*").From("dbr_people").Where("id = $1", id), &person)
	assert.NoError(t, err)

	assert.Equal(t, person.ID, id)
	assert.Equal(t, person.Name, "Barack")
	assert.Equal(t, person.Email.Valid, true)
	assert.Equal(t, person.Email.String, "obama@whitehouse.gov")

	err = tx.Commit()
	assert.NoError(t, err)
}

func TestTransactionRollbackReal(t *testing.T) {
	// Insert by specifying values
	s := createRealSessionWithFixtures()

	tx, err := s.Begin()
	assert.NoError(t, err)

	var person dbrPerson
	err = tx.QueryStruct(dat.Select("*").From("dbr_people").Where("email = $1", "jonathan@uservoice.com"), &person)
	assert.NoError(t, err)
	assert.Equal(t, person.Name, "Jonathan")

	err = tx.Rollback()
	assert.NoError(t, err)
}
