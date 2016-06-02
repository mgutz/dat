package runner

import (
	"testing"

	"gopkg.in/mgutz/dat.v1"
	"gopkg.in/stretchr/testify.v1/assert"
)

func TestDeleteReal(t *testing.T) {
	s := beginTxWithFixtures()
	defer s.AutoRollback()

	var id int64
	// Insert a Barack
	err := s.InsertInto("people").
		Columns("name", "email").
		Values("Barack", "barack@whitehouse.gov").
		Returning("id").
		QueryScalar(&id)
	assert.NoError(t, err)

	// Delete Barack
	res, err := s.DeleteFrom("people").Where("id = $1", id).Exec()
	assert.NoError(t, err)

	// Ensure we only reflected one row and that the id no longer exists
	assert.EqualValues(t, res.RowsAffected, 1)

	var count int64
	err = s.Select("count(*)").
		From("people").
		Where("id = $1", id).
		QueryScalar(&count)
	assert.NoError(t, err)
	assert.EqualValues(t, count, 0)
}

func TestDeleteScope(t *testing.T) {
	s := beginTxWithFixtures()
	defer s.AutoRollback()

	var id int64
	// Insert a Barack
	err := s.InsertInto("people").
		Columns("name", "email").
		Values("Barack", "barack@whitehouse.gov").
		Returning("id").
		QueryScalar(&id)
	assert.NoError(t, err)

	scope := dat.NewScope("WHERE id = :id", dat.M{"id": 0})

	// Delete Barack
	res, err := s.
		DeleteFrom("people").
		ScopeMap(scope, dat.M{"id": id}).
		Exec()
	assert.NoError(t, err)

	// Ensure we only reflected one row and that the id no longer exists
	assert.EqualValues(t, res.RowsAffected, 1)

	var count int64
	err = s.Select("count(*)").
		From("people").
		Where("id = $1", id).
		QueryScalar(&count)
	assert.NoError(t, err)
	assert.EqualValues(t, count, 0)
}
