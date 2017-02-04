package dat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSQLBasic(t *testing.T) {
	sql, args, err := JSQL("SELECT name FROM users WHERE id=$1", 42).ToSQL()
	assert.NoError(t, err)

	expected := `
		SELECT row_to_json(dat__item.*)
		FROM (
			SELECT name
			FROM users
			WHERE id=$1
		) as dat__item
	`

	assert.Equal(t, stripWS(expected), stripWS(sql))
	assert.Equal(t, []interface{}{42}, args)
}

func TestJSQLMany(t *testing.T) {
	sql, args, err := JSQL("SELECT name FROM users WHERE id=$1", 42).
		Many("hobbies", SQL("SELECT * FROM hobbies WHERE id=$1", 100)).
		ToSQL()
	assert.NoError(t, err)

	expected := `
		SELECT row_to_json(dat__item.*)
		FROM (
			SELECT
				(SELECT array_agg(dat__hobbies.*) FROM (SELECT * FROM hobbies WHERE id = $2) AS dat__hobbies) AS hobbies,
				name
			FROM
				users
			WHERE
				id=$1
		) as dat__item
	`

	assert.Equal(t, stripWS(expected), stripWS(sql))
	assert.Equal(t, []interface{}{42, 100}, args)
}
