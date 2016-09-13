package runner

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecScriptSuccess(t *testing.T) {
	script := `
INSERT INTO people (name)
VALUES ('execScriptA%v');
GO
INSERT INTO people (name)
VALUES ('execScriptB%v');
	`

	// ensure extra GO's do not have side effects
	variations := []string{
		fmt.Sprintf(script, 0, 0),
		fmt.Sprintf("GO\n"+script, 1, 1),
		fmt.Sprintf(script+"\nGO", 2, 2),
	}

	for i, s := range variations {
		err := testDB.ExecScript(s)
		assert.NoError(t, err)
		//assert.EqualError(t, err, sql.ErrNoRows.Error())

		a := fmt.Sprintf("execScriptA%d", i)
		b := fmt.Sprintf("execScriptB%d", i)
		var name string

		err = testDB.SQL(`SELECT name FROM people WHERE name = $1`, a).QueryScalar(&name)
		assert.NoError(t, err)
		assert.Equal(t, a, name)

		err = testDB.SQL(`SELECT name FROM people WHERE name = $1`, b).QueryScalar(&name)
		assert.NoError(t, err)
		assert.Equal(t, b, name)

	}
}

func TestExecScriptError(t *testing.T) {
	script := `
IUNSERT -- should be invalid
GO
	`

	err := testDB.ExecScript(script)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "IUNSERT -- should be invalid")
}
