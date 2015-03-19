package dat

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSnakeCaseID(t *testing.T) {
	assert.Equal(t, camelCaseToSnakeCaseID("ID"), "id")
	assert.Equal(t, camelCaseToSnakeCaseID("a"), "a")
	assert.Equal(t, camelCaseToSnakeCaseID("SomeId"), "some_id")
	assert.Equal(t, camelCaseToSnakeCaseID("SomeID"), "some_id")
}

func TestLoadSQLMap(t *testing.T) {
	s := `
--@foo
SELECT *
FROM foo;

--@bar
SELECT *
FROM bar;
`
	r := bytes.NewBufferString(s)
	m, err := SQLMapFromReader(r)
	assert.NoError(t, err)
	assert.Equal(t, "SELECT *\nFROM foo;\n\n", m["foo"])
	assert.Equal(t, "SELECT *\nFROM bar;\n", m["bar"])
}

func TestSQLSlice(t *testing.T) {
	s := `
SELECT *
FROM foo;
GO
SELECT *
FROM bar;`
	sli, err := SQLSliceFromString(s)
	assert.NoError(t, err)
	assert.Equal(t, "\nSELECT *\nFROM foo;\n", sli[0])
	assert.Equal(t, "\nSELECT *\nFROM bar;", sli[1])
}
