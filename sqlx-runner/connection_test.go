package runner

import (
	"testing"

	"gopkg.in/mgutz/dat.v1"
	"gopkg.in/stretchr/testify.v1/assert"
)

func TestConnectionExec(t *testing.T) {
	installFixtures()

	id := 0
	str := ""
	err := testDB.InsertInto("people").
		Columns("name", "foo").
		Values("conn1", "---").
		Returning("id", "foo").
		QueryScalar(&id, &str)

	assert.NoError(t, err)
	assert.True(t, id > 0)
	assert.Equal(t, "---", str)

	dat.EnableInterpolation = true
	_, err = testDB.Update("people").
		Set("foo", dat.DEFAULT).
		Returning("foo").
		Exec()
	dat.EnableInterpolation = false
	assert.NoError(t, err)
}

func TestEscapeSequences(t *testing.T) {
	installFixtures()

	dat.EnableInterpolation = true
	id := 0
	str := ""
	expect := "I said, \"a's \\ \\\b\f\n\r\t\x1a\"你好'; DELETE FROM people"

	err := testDB.InsertInto("people").
		Columns("name", "foo").
		Values("conn1", expect).
		Returning("id", "foo").
		QueryScalar(&id, &str)

	assert.NoError(t, err)
	assert.True(t, id > 0)
	assert.Equal(t, expect, str)
	dat.EnableInterpolation = false
}
