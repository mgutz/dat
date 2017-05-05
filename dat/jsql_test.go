package dat

import (
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"
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
				(SELECT array_agg(dat__hobbies.*) FROM (SELECT * FROM hobbies WHERE id = $2) AS dat__hobbies) AS "hobbies",
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

func TestJSQLVector(t *testing.T) {
	sql, args, err := JSQL("SELECT name FROM users WHERE id=$1", 42).
		Many("hobbies", SQL("SELECT * FROM hobbies WHERE id=$1", 100)).
		Vector("child_ages", SQL("SELECT age FROM users WHERE parent_id=$1", 142)).
		ToSQL()
	assert.NoError(t, err)

	expected := `
		SELECT row_to_json(dat__item.*)
		FROM (
			SELECT
				(SELECT array_agg(dat__hobbies.*) FROM (SELECT * FROM hobbies WHERE id = $2) AS dat__hobbies) AS "hobbies",
				(SELECT array_agg(dat__child_ages.dat__scalar) FROM (SELECT age FROM users WHERE parent_id=$3) AS dat__child_ages (dat__scalar)) AS "child_ages",
				name
			FROM
				users
			WHERE
				id=$1
		) as dat__item
	`

	assert.Equal(t, stripWS(expected), stripWS(sql))
	assert.Equal(t, []interface{}{42, 100, 142}, args)
}
