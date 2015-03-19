package runner

import (
	"testing"

	"github.com/mgutz/dat/v1"
	"github.com/stretchr/testify/assert"
)

func TestUpsertReal(t *testing.T) {
	// Insert by specifying values
	s := createRealSessionWithFixtures()
	defer s.Close()
	var id int64
	err := s.Upsert("people").
		Columns("name", "email").
		Values("mario", "mgutz@mgutz.com").
		Where("name = $1", "mario").
		Returning("id").
		QueryScalar(&id)
	assert.NoError(t, err)
	assert.True(t, id > 0)

	var id2 int64
	err = s.Upsert("people").
		Columns("name", "email").
		Values("mario", "mario@foo.com").
		Where("name = $1", "mario").
		Returning("id").
		QueryScalar(&id2)
	assert.NoError(t, err)
	assert.Equal(t, id, id2)

	// Insert by specifying a record (ptr to struct)
	person := Person{Name: "Barack"}
	person.Email.Valid = true
	person.Email.String = "obama1@whitehouse.gov"

	err = s.
		Upsert("people").
		Columns("name", "email").
		Record(&person).
		Where("name = $1", "Barack").
		Returning("id", "created_at").
		QueryStruct(&person)
	assert.NoError(t, err)
	assert.True(t, person.ID > 0)
	assert.NotEqual(t, person.CreatedAt, dat.NullTime{})
}
