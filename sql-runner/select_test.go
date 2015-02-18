package runner

import (
	"testing"

	"github.com/mgutz/dat"
	"github.com/stretchr/testify/assert"
)

func TestSelectQueryStructs(t *testing.T) {
	s := createRealSessionWithFixtures()

	var people []*dbrPerson
	count, err := s.QueryStructs(
		dat.Select("id", "name", "email").
			From("dbr_people").
			OrderBy("id ASC"),
		&people,
	)

	assert.NoError(t, err)
	assert.Equal(t, count, 2)

	assert.Equal(t, len(people), 2)
	if len(people) == 2 {
		// Make sure that the Ids are set. It's possible (maybe?) that different DBs set ids differently so
		// don't assume they're 1 and 2.
		assert.True(t, people[0].ID > 0)
		assert.True(t, people[1].ID > people[0].ID)

		assert.Equal(t, people[0].Name, "Jonathan")
		assert.True(t, people[0].Email.Valid)
		assert.Equal(t, people[0].Email.String, "jonathan@uservoice.com")
		assert.Equal(t, people[1].Name, "Dmitri")
		assert.True(t, people[1].Email.Valid)
		assert.Equal(t, people[1].Email.String, "zavorotni@jadius.com")
	}

	// TODO: test map
}

func TestSelectQueryStruct(t *testing.T) {
	s := createRealSessionWithFixtures()

	// Found:
	var person dbrPerson
	err := s.QueryStruct(
		dat.Select("id", "name", "email").
			From("dbr_people").
			Where("email = $1", "jonathan@uservoice.com"),
		&person,
	)
	assert.NoError(t, err)
	assert.True(t, person.ID > 0)
	assert.Equal(t, person.Name, "Jonathan")
	assert.True(t, person.Email.Valid)
	assert.Equal(t, person.Email.String, "jonathan@uservoice.com")

	// Not found:
	var person2 dbrPerson
	err = s.QueryStruct(
		dat.Select("id", "name", "email").
			From("dbr_people").Where("email = $1", "dontexist@uservoice.com"),
		&person2,
	)
	assert.Equal(t, err, dat.ErrNotFound)
}

func TestSelectBySqlQueryStructs(t *testing.T) {
	s := createRealSessionWithFixtures()

	var people []*dbrPerson
	count, err := s.QueryStructs(
		dat.SQL("SELECT name FROM dbr_people WHERE email IN $1", []string{"jonathan@uservoice.com"}),
		&people,
	)

	assert.NoError(t, err)
	assert.Equal(t, count, 1)
	if len(people) == 1 {
		assert.Equal(t, people[0].Name, "Jonathan")
		assert.Equal(t, people[0].ID, 0)              // not set
		assert.Equal(t, people[0].Email.Valid, false) // not set
		assert.Equal(t, people[0].Email.String, "")   // not set
	}
}

func TestSelectQueryScalar(t *testing.T) {
	s := createRealSessionWithFixtures()

	var name string
	err := s.QueryScan(
		dat.Select("name").
			From("dbr_people").
			Where("email = 'jonathan@uservoice.com'"),
		&name,
	)

	assert.NoError(t, err)
	assert.Equal(t, name, "Jonathan")

	var id int64
	err = s.QueryScan(dat.Select("id").From("dbr_people").Limit(1), &id)

	assert.NoError(t, err)
	assert.True(t, id > 0)
}

func TestSelectQuerySlice(t *testing.T) {
	s := createRealSessionWithFixtures()

	var names []string
	count, err := s.QuerySlice(dat.Select("name").From("dbr_people"), &names)

	assert.NoError(t, err)
	assert.Equal(t, count, 2)
	assert.Equal(t, names, []string{"Jonathan", "Dmitri"})

	var ids []int64
	count, err = s.QuerySlice(dat.Select("id").From("dbr_people").Limit(1), &ids)

	assert.NoError(t, err)
	assert.Equal(t, count, 1)
	assert.Equal(t, ids, []int64{1})
}

func TestScalar(t *testing.T) {
	s := createRealSessionWithFixtures()

	var name string
	err := s.QueryScan(dat.Select("name").From("dbr_people").Where("email = 'jonathan@uservoice.com'"), &name)
	assert.NoError(t, err)
	assert.Equal(t, name, "Jonathan")

	var count int64
	err = s.QueryScan(dat.Select("COUNT(*)").From("dbr_people"), &count)
	assert.NoError(t, err)
	assert.Equal(t, count, 2)
}

// Series of tests that test mapping struct fields to columns
