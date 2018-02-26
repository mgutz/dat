package dat

import (
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"
)

func TestAppliedExpr(t *testing.T) {
	arrArgs := [][]interface{}{
		{"one", 1},
		{"two", 2},
	}

	// eg runner.ExecExpr(AppliedExpr("SELECT $1, $2", arrArgs))
	str, args, err := AppliedExpr("SELECT $1, $2", arrArgs).Expression()
	assert.NoError(t, err)
	assert.Nil(t, args)
	assert.Equal(t, str, "SELECT 'one', 1;SELECT 'two', 2;")
}
