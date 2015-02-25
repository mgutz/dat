package runner

import (
	"testing"

	"github.com/mgutz/dat"
	"github.com/stretchr/testify/assert"
)

func TestUpdateKeywordColumnName(t *testing.T) {
	s := createRealSessionWithFixtures()
	defer s.AutoCommit()

	// Insert a user with a key
	res, err := s.
		InsertInto("people").
		Columns("name", "email", "key").
		Values("Benjamin", "ben@whitehouse.gov", "6").
		Exec()
	assert.NoError(t, err)

	// Update the key
	res, err = s.Update("people").Set("key", "6-revoked").Where(dat.Eq{"key": "6"}).Exec()
	assert.NoError(t, err)

	// Assert our record was updated (and only our record)
	assert.NoError(t, err)
	assert.Equal(t, res.RowsAffected, 1)

	var person Person
	err = s.Select("*").From("people").Where(dat.Eq{"email": "ben@whitehouse.gov"}).QueryStruct(&person)
	assert.NoError(t, err)

	assert.Equal(t, person.Name, "Benjamin")
	assert.Equal(t, person.Key.String, "6-revoked")
}

func TestUpdateReal(t *testing.T) {
	s := createRealSessionWithFixtures()
	defer s.AutoCommit()

	var id int64
	// Insert a George
	s.InsertInto("people").Columns("name", "email").
		Values("George", "george@whitehouse.gov").
		Returning("id").
		QueryScalar(&id)

	// Rename our George to Barack
	_, err := s.Update("people").SetMap(map[string]interface{}{"name": "Barack", "email": "barack@whitehouse.gov"}).Where("id = $1", id).Exec()

	assert.NoError(t, err)

	var person Person
	err = s.Select("*").From("people").Where("id = $1", id).QueryStruct(&person)
	assert.NoError(t, err)

	assert.Equal(t, person.ID, id)
	assert.Equal(t, person.Name, "Barack")
	assert.Equal(t, person.Email.Valid, true)
	assert.Equal(t, person.Email.String, "barack@whitehouse.gov")
}
