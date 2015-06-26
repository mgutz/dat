package runner

import (
	"testing"

	"gopkg.in/mgutz/dat.v1"
	"gopkg.in/stretchr/testify.v1/assert"
)

func TestInsect(t *testing.T) {
	// Insert by specifying values
	s := beginTxWithFixtures()
	defer s.AutoRollback()

	var id int64
	err := s.Insect("people").
		Columns("name", "email").
		Values("mario", "mgutz@mgutz.com").
		Returning("id").
		QueryScalar(&id)
	assert.NoError(t, err)
	assert.True(t, id > 0)

	// Insert by specifying a record (ptr to struct)
	person := Person{Name: "Barack"}
	person.Email.Valid = true
	person.Email.String = "obama1@whitehouse.gov"

	err = s.
		Insect("people").
		Columns("name", "email").
		Record(&person).
		Returning("id", "created_at").
		QueryStruct(&person)
	assert.NoError(t, err)
	assert.True(t, person.ID > 0)
	assert.NotEqual(t, person.CreatedAt, dat.NullTime{})
}

// Insect should select existing record.
func TestInsectSelect(t *testing.T) {
	s := beginTxWithFixtures()
	defer s.AutoRollback()

	var id int64
	err := s.Insect("people").
		Columns("name", "email").
		Values("Mario", "mario@acme.com").
		Returning("id").
		QueryScalar(&id)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, id)
}

// Insect should select existing record without updating it (see Upsert)
func TestInsectSelectWhere(t *testing.T) {
	s := beginTxWithFixtures()
	defer s.AutoRollback()

	var p Person
	err := s.Insect("people").
		Columns("name", "email").
		Values("Foo", "bar@acme.com").
		Where("id = $1", 1).
		Returning("id", "name", "email").
		QueryStruct(&p)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, p.ID)
	assert.Equal(t, "Mario", p.Name)
	assert.Equal(t, "mario@acme.com", p.Email.String)
}
