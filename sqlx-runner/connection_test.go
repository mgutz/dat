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
		QueryScalar(&id, &str)

	assert.NoError(t, err)
	assert.True(t, id > 0)
	assert.Equal(t, "---", str)

	dat.EnableInterpolation = true
	_, err = testConn.Update("people").
		Set("foo", dat.DEFAULT).
		Returning("foo").
		Exec()
	dat.EnableInterpolation = false
	assert.NoError(t, err)
}

func TestEscapeSequences(t *testing.T) {
	createRealSessionWithFixtures()

	dat.EnableInterpolation = true
	id := 0
	str := ""
	expect := "I said, \"a's \\ \b\f\n\r\t\x1a\""

	err := testConn.InsertInto("people").
		Columns("name", "foo").
		Values("conn1", expect).
		Returning("id", "foo").
		QueryScalar(&id, &str)

	assert.NoError(t, err)
	assert.True(t, id > 0)
	assert.Equal(t, expect, str)
	dat.EnableInterpolation = false
}
