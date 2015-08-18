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

func TestCallInterpolate(t *testing.T) {
	sql, args, err := Call("foo", 1).SetIsInterpolated(true).Interpolate()
	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM foo(1)", sql)
	assert.Exactly(t, []interface{}(nil), args)

	sql, args, err = Call("foo", 1).Interpolate()
	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM foo($1)", sql)
	assert.Exactly(t, []interface{}{1}, args)
}
