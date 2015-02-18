package runner

import (
	"testing"

	"github.com/mgutz/dat"
	"github.com/stretchr/testify/assert"
)

func TestDeleteReal(t *testing.T) {
	s := createRealSessionWithFixtures()

	var id int64

	// Insert a Barack
	b := dat.InsertInto("dbr_people").Columns("name", "email").
		Values("Barack", "barack@whitehouse.gov").
		Returning("id")
	s.QueryScan(b, &id)

	// Delete Barack
	res, err := s.Exec(dat.DeleteFrom("dbr_people").Where("id = $1", id))
	assert.NoError(t, err)

	// Ensure we only reflected one row and that the id no longer exists
	rowsAff, err := res.RowsAffected()
	assert.NoError(t, err)
	assert.Equal(t, rowsAff, 1)

	var count int64
	err = s.QueryScan(
		dat.Select("count(*)").
			From("dbr_people").
			Where("id = $1", id),
		&count,
	)
	assert.NoError(t, err)
	assert.Equal(t, count, 0)
}
