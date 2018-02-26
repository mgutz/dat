package runner

import (
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"

	"github.com/mgutz/dat/dat"
)

func TestRole(t *testing.T) {
	tx, err := testDB.Begin()
	if err != nil {
		t.Error(err)
	}
	defer tx.AutoRollback()

	t.Run("AppliedExprHappy", func(t *testing.T) {
		_, err := tx.ExecExpr(dat.AppliedExpr(
			`update people set name = $1 where id = $2`,
			[][]interface{}{
				{"red", 1},
				{"blue", 2},
			},
		))
		assert.NoError(t, err)

		var names []string
		err = tx.SQL("select name from people where id in (1,2)").QuerySlice(&names)
		assert.NoError(t, err)
		assert.EqualValues(t, names, []string{"red", "blue"})
		//assert.Equal(t, result.RowsAffected, 2)
	})

}
