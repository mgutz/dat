package runner

import (
	"testing"

	"github.com/mgutz/dat"
	"github.com/stretchr/testify/assert"
)

func TestUpdateKeywordColumnName(t *testing.T) {
	s := createRealSessionWithFixtures()

	// Insert a user with a key
	b := dat.InsertInto("dbr_people").Columns("name", "email", "key").Values("Benjamin", "ben@whitehouse.gov", "6")
	res, err := s.Exec(b)
	assert.NoError(t, err)

	// Update the key
	res, err = s.Exec(dat.Update("dbr_people").Set("key", "6-revoked").Where(dat.Eq{"key": "6"}))
	assert.NoError(t, err)

	// Assert our record was updated (and only our record)
	rowsAff, err := res.RowsAffected()
	assert.NoError(t, err)
	assert.Equal(t, rowsAff, 1)

	var person dbrPerson
	err = s.QueryStruct(dat.Select("*").From("dbr_people").Where(dat.Eq{"email": "ben@whitehouse.gov"}), &person)
	assert.NoError(t, err)

	assert.Equal(t, person.Name, "Benjamin")
	assert.Equal(t, person.Key.String, "6-revoked")
}

func TestUpdateReal(t *testing.T) {
	s := createRealSessionWithFixtures()

	var id int64
	// Insert a George
	b := dat.InsertInto("dbr_people").Columns("name", "email").
		Values("George", "george@whitehouse.gov").
		Returning("id")
	s.QueryScan(b, &id)

	// Rename our George to Barack
	_, err := s.Exec(dat.Update("dbr_people").SetMap(map[string]interface{}{"name": "Barack", "email": "barack@whitehouse.gov"}).Where("id = $1", id))

	assert.NoError(t, err)

	var person dbrPerson
	err = s.QueryStruct(dat.Select("*").From("dbr_people").Where("id = $1", id), &person)
	assert.NoError(t, err)

	assert.Equal(t, person.ID, id)
	assert.Equal(t, person.Name, "Barack")
	assert.Equal(t, person.Email.Valid, true)
	assert.Equal(t, person.Email.String, "barack@whitehouse.gov")
}
