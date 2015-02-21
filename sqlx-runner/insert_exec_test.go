package runner

import (
	"testing"

	"github.com/mgutz/dat"
	"github.com/stretchr/testify/assert"
)

func TestInsertKeywordColumnName(t *testing.T) {
	// Insert a column whose name is reserved
	s := createRealSessionWithFixtures()
	res, err := s.
		InsertInto("people").
		Columns("name", "key").
		Values("Barack", "44").
		Exec()

	assert.NoError(t, err)

	rowsAff, err := res.RowsAffected()
	assert.NoError(t, err)
	assert.Equal(t, rowsAff, 1)
}

func TestInsertDefault(t *testing.T) {
	s := createRealSessionWithFixtures()

	var str string
	err := s.
		InsertInto("people").Columns("name", "foo").
		Values("Barack", "fool").
		Returning("foo").
		QueryScalar(&str)
	assert.NoError(t, err)
	assert.Equal(t, str, "fool")

	dat.EnableInterpolation = true
	err = s.
		Update("people").
		Set("foo", dat.DEFAULT).
		Returning("foo").
		QueryScalar(&str)
	dat.EnableInterpolation = false
	assert.NoError(t, err)
	assert.Equal(t, str, "bar")
}

func TestInsertReal(t *testing.T) {
	// Insert by specifying values
	s := createRealSessionWithFixtures()
	var id int64
	err := s.InsertInto("people").
		Columns("name", "email").
		Values("Barack", "obama0@whitehouse.gov").
		Returning("id").
		QueryScalar(&id)
	assert.NoError(t, err)
	assert.True(t, id > 0)

	// Insert by specifying a record (ptr to struct)
	person := Person{Name: "Barack"}
	person.Email.Valid = true
	person.Email.String = "obama1@whitehouse.gov"

	err = s.
		InsertInto("people").
		Columns("name", "email").
		Record(&person).
		Returning("id", "created_at").
		QueryStruct(&person)
	assert.NoError(t, err)
	assert.True(t, person.ID > 0)
	assert.NotEqual(t, person.CreatedAt, dat.NullTime{})

	// Insert by specifying a record (struct)
	person2 := Person{Name: "Barack"}
	person2.Email.Valid = true
	person2.Email.String = "obama2@whitehouse.gov"
	err = s.
		InsertInto("people").Columns("name", "email").
		Record(person2).
		Returning("id").
		QueryStruct(&person2)
	assert.NoError(t, err)
	assert.True(t, person2.ID > 0)
	assert.NotEqual(t, person.ID, person2.ID)
}

func TestInsertMultipleRecords(t *testing.T) {
	assert := assert.New(t)

	s := createRealSessionWithFixtures()
	res, err := s.
		InsertInto("people").
		Columns("name", "email").
		Values("apple", "apple@fruits.local").
		Values("orange", "orange@fruits.local").
		Values("pear", "pear@fruits.local").
		Exec()
	assert.NoError(err)
	n, err := res.RowsAffected()
	assert.Equal(n, 3)

	person1 := Person{Name: "john_timr"}
	person2 := Person{Name: "jane_timr"}

	res, err = s.InsertInto("people").
		Columns("name", "email").
		Record(&person1).
		Record(&person2).
		Exec()
	assert.NoError(err)
	n, err = res.RowsAffected()
	assert.NoError(err)
	assert.Equal(n, 2)

	people := []Person{}
	err = s.
		Select("name").
		From("people").
		Where("name like $1", "%_timr").
		QueryStructs(&people)
	assert.NoError(err)
	assert.Equal(len(people), 2)

	n = 0
	for _, person := range people {
		if person.Name == "john_timr" {
			n++
		}
		if person.Name == "jane_timr" {
			n++
		}
	}
	assert.Equal(n, 2)
}
