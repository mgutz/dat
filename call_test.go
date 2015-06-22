package dat

import (
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"
)

func TestCallSql(t *testing.T) {
	sql, args := Call("foo", 1, "two").ToSQL()
	assert.Equal(t, "SELECT * FROM foo($1,$2)", sql)
	assert.Exactly(t, []interface{}{1, "two"}, args)
}

func TestCallNoArgsSql(t *testing.T) {
	sql, args := Call("foo").ToSQL()
	assert.Equal(t, "SELECT * FROM foo()", sql)
	assert.Nil(t, args)
}
