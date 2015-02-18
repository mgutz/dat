package runner

import (
	"testing"

	"github.com/mgutz/dat"
	"github.com/stretchr/testify/assert"
)

func TestConnectionExec(t *testing.T) {

	createRealSessionWithFixtures()

	b := dat.InsertInto("dbr_people").
		Columns("name", "foo").
		Values("conn1", "---").
		Returning("id", "foo")
	id := 0
	str := ""
	err := testConn.QueryScan(b, &id, &str)
	assert.NoError(t, err)
	assert.True(t, id > 0)
	assert.Equal(t, "---", str)

	ub := dat.Update("dbr_people").
		Set("foo", dat.DEFAULT).
		Returning("foo")
	_, err = testConn.Exec(ub)
	assert.NoError(t, err)
}
