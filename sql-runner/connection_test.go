package runner

import (
	"testing"

	"github.com/mgutz/dat"
	"github.com/stretchr/testify/assert"
)

func TestConnectionExec(t *testing.T) {
	createRealSessionWithFixtures()

	id := 0
	str := ""
	err := testConn.InsertInto("people").
		Columns("name", "foo").
		Values("conn1", "---").
		Returning("id", "foo").
		QueryScan(&id, &str)
	assert.NoError(t, err)
	assert.True(t, id > 0)
	assert.Equal(t, "---", str)

	_, err = testConn.Update("people").
		Set("foo", dat.DEFAULT).
		Returning("foo").
		Exec()
	assert.NoError(t, err)
}
