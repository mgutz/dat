package runner

import (
	"testing"

	"github.com/mgutz/dat/v1"
	"github.com/stretchr/testify/assert"
)

func TestDeleteReal(t *testing.T) {
	s := createRealSessionWithFixtures()
	defer s.Close()

	var id int64
	// Insert a Barack
	s.InsertInto("people").
		Columns("name", "email").
		Values("Barack", "barack@whitehouse.gov").
		Returning("id").
		QueryScalar(&id)

	// Delete Barack
	res, err := s.DeleteFrom("people").Where("id = $1", id).Exec()
	assert.NoError(t, err)

	// Ensure we only reflected one row and that the id no longer exists
	assert.Equal(t, res.RowsAffected, 1)

	var count int64
	err = s.Select("count(*)").
		From("people").
		Where("id = $1", id).
		QueryScalar(&count)
	assert.NoError(t, err)
	assert.Equal(t, count, 0)
}

func TestDeleteScope(t *testing.T) {
	s := createRealSessionWithFixtures()
	defer s.Close()

	var id int64
	// Insert a Barack
	s.InsertInto("people").
		Columns("name", "email").
		Values("Barack", "barack@whitehouse.gov").
		Returning("id").
		QueryScalar(&id)

	scope := dat.NewScope("WHERE id = :id", dat.M{"id": 0})

	// Delete Barack
	res, err := s.
		DeleteFrom("people").
		ScopeMap(scope, dat.M{"id": id}).
		Exec()
	assert.NoError(t, err)

	// Ensure we only reflected one row and that the id no longer exists
	assert.Equal(t, res.RowsAffected, 1)

	var count int64
	err = s.Select("count(*)").
		From("people").
		Where("id = $1", id).
		QueryScalar(&count)
	assert.NoError(t, err)
	assert.Equal(t, count, 0)
}
