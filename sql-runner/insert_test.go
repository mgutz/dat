package runner

import (
	"testing"

	"github.com/mgutz/dat"
	"github.com/stretchr/testify/assert"
)

func TestInsertKeywordColumnName(t *testing.T) {
	// Insert a column whose name is reserved
	s := createRealSessionWithFixtures()
	b := dat.InsertInto("dbr_people").Columns("name", "key").Values("Barack", "44")
	res, err := s.Exec(b)
	assert.NoError(t, err)

	rowsAff, err := res.RowsAffected()
	assert.NoError(t, err)
	assert.Equal(t, rowsAff, 1)
}

func TestInsertReal(t *testing.T) {
	// Insert by specifying values
	s := createRealSessionWithFixtures()
	var id int64
	b := dat.InsertInto("dbr_people").
		Columns("name", "email").
		Values("Barack", "obama0@whitehouse.gov").
		Returning("id")
	err := s.QueryScalar(&id, b)
	assert.NoError(t, err)
	assert.True(t, id > 0)

	// Insert by specifying a record (ptr to struct)
	person := dbrPerson{Name: "Barack"}
	person.Email.Valid = true
	person.Email.String = "obama1@whitehouse.gov"
	b = dat.InsertInto("dbr_people").
		Columns("name", "email").
		Record(&person).
		Returning("id", "created_at")
	err = s.QueryStruct(&person, b)
	assert.NoError(t, err)
	assert.True(t, person.ID > 0)
	assert.NotEqual(t, person.CreatedAt, dat.NullTime{})

	// Insert by specifying a record (struct)
	person2 := dbrPerson{Name: "Barack"}
	person2.Email.Valid = true
	person2.Email.String = "obama2@whitehouse.gov"
	b = dat.InsertInto("dbr_people").Columns("name", "email").
		Record(person2).
		Returning("id")
	err = s.QueryStruct(&person2, b)
	assert.NoError(t, err)
	assert.True(t, person2.ID > 0)
	assert.NotEqual(t, person.ID, person2.ID)
}

func TestInsertMultipleRecords(t *testing.T) {
	assert := assert.New(t)

	s := createRealSessionWithFixtures()
	b := dat.InsertInto("dbr_people").
		Columns("name", "email").
		Values("apple", "apple@fruits.local").
		Values("orange", "orange@fruits.local").
		Values("pear", "pear@fruits.local")
	res, err := s.Exec(b)
	assert.NoError(err)
	n, err := res.RowsAffected()
	assert.Equal(n, 3)

	person1 := dbrPerson{Name: "john_timr"}
	person2 := dbrPerson{Name: "jane_timr"}

	b = dat.InsertInto("dbr_people").
		Columns("name", "email").
		Record(&person1).
		Record(&person2)
	res, err = s.Exec(b)
	assert.NoError(err)
	n, err = res.RowsAffected()
	assert.NoError(err)
	assert.Equal(n, 2)

	people := []*dbrPerson{}
	n, err = s.QueryStructs(&people,
		dat.Select("name").From("dbr_people").Where("name like $1", "%_timr"),
	)
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
