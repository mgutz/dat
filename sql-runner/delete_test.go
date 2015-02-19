package runner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteReal(t *testing.T) {
	s := createRealSessionWithFixtures()

	var id int64

	// Insert a Barack
	s.InsertInto("people").Columns("name", "email").
		Values("Barack", "barack@whitehouse.gov").
		Returning("id").
		QueryScan(&id)

	// Delete Barack
	res, err := s.DeleteFrom("people").Where("id = $1", id).Exec()
	assert.NoError(t, err)

	// Ensure we only reflected one row and that the id no longer exists
	rowsAff, err := res.RowsAffected()
	assert.NoError(t, err)
	assert.Equal(t, rowsAff, 1)

	var count int64
	err = s.Select("count(*)").
		From("people").
		Where("id = $1", id).
		QueryScan(&count)
	assert.NoError(t, err)
	assert.Equal(t, count, 0)
}
