package dat

import (
	"bytes"
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"
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

func TestParseReader(t *testing.T) {
	s := `
--@key=foo
SELECT *
FROM foo;

--@key=bar
SELECT *
FROM bar;

--@sproc
create function f_foo() as $$
begin
end; $$ language plpgsql;
`
	r := bytes.NewBufferString(s)
	a, err := PartitionKV(r, "--@", "=")
	assert.NoError(t, err)

	assert.Equal(t, "key", a[0]["_kind"])
	assert.Equal(t, "SELECT *\nFROM foo;\n\n", a[0]["_body"])
	assert.Equal(t, "foo", a[0]["key"])

	assert.Equal(t, "key", a[1]["_kind"])
	assert.Equal(t, "SELECT *\nFROM bar;\n\n", a[1]["_body"])
	assert.Equal(t, "bar", a[1]["key"])

	assert.Equal(t, "sproc", a[2]["_kind"])
	assert.Equal(t, "create function f_foo() as $$\nbegin\nend; $$ language plpgsql;\n", a[2]["_body"])
	assert.Equal(t, "", a[2]["sproc"])
}
