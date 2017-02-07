package runner

import (
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"
)

func TestVersion(t *testing.T) {
	// require at least 9.3+ for testing
	assert.True(t, testDB.Version > 90300)
}

func TestMissingDestinationColumns(t *testing.T) {
	type MissingColumns struct {
		ID int `db:"id"`
	}
	var row MissingColumns
	err := testDB.SQL(`SELECT * FROM posts LIMIT 1`).QueryStruct(&row)
	assert.Error(t, err, "Result had more columns than destination struct. Should have returned error")

	// sqlx will error when the result of a query has columns which are
	// not present in destination struct. This becomes problematic
	// when queries use SELECT *.
	err = testDB.Loose().SQL(`SELECT * FROM posts LIMIT 1`).QueryStruct(&row)
	assert.NoError(t, err, "In non-strict mode, having more result columns than destination should not error")

	var rows []*MissingColumns
	err = testDB.Loose().SQL(`SELECT * FROM posts LIMIT 2`).QueryStructs(&rows)
	assert.NoError(t, err, "In non-strict mode, having more result columns than destination should not error")
}
