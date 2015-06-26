package dat

import (
	"testing"
	"time"

	"gopkg.in/stretchr/testify.v1/assert"
)

func TestSelectDocSQLNoDocs(t *testing.T) {
	sql, args := SelectDoc("b", "c").From("a").Where("d=$1", 4).ToSQL()

	expected := `
		SELECT row_to_json(dat__item.*)
		FROM (
			SELECT b,c
			FROM a
			WHERE (d=$1)
		) as dat__item
	`

	assert.Equal(t, stripWS(expected), stripWS(sql))
	assert.Equal(t, []interface{}{4}, args)
}

func TestSelectDocSQLDocs(t *testing.T) {
	sql, args := SelectDoc("b", "c").
		Many("f", `SELECT g, h FROM f WHERE id= $1`, 4).
		Many("x", `SELECT id, y, z FROM x`).
		From("a").
		Where("d=$1", 4).
		ToSQL()

	expected := `
	SELECT row_to_json(dat__item.*)
	FROM (
		SELECT
			b,
			c,
			(SELECT array_agg(dat__f.*) FROM (SELECT g,h FROM f WHERE id=$1) AS dat__f) AS "f",
			(SELECT array_agg(dat__x.*) FROM (SELECT id,y,z FROM x) AS dat__x) AS "x"
		FROM a
		WHERE (d=$2)
	) as dat__item
	`
	assert.Equal(t, stripWS(expected), stripWS(sql))
	assert.Equal(t, []interface{}{4, 4}, args)
}

func TestSelectDocSQLInnerSQL(t *testing.T) {
	sql, args := SelectDoc("b", "c").
		Many("f", `SELECT g, h FROM f WHERE id= $1`, 4).
		Many("x", `SELECT id, y, z FROM x`).
		InnerSQL(`
			FROM a
			WHERE d = $1
		`, 4).
		ToSQL()

	expected := `
	SELECT row_to_json(dat__item.*)
	FROM (
		SELECT
			b,
			c,
			(SELECT array_agg(dat__f.*) FROM (SELECT g,h FROM f WHERE id=$1) AS dat__f) AS "f",
			(SELECT array_agg(dat__x.*) FROM (SELECT id,y,z FROM x) AS dat__x) AS "x"
		FROM a
		WHERE d=$2
	) as dat__item
	`
	assert.Equal(t, stripWS(expected), stripWS(sql))
	assert.Equal(t, []interface{}{4, 4}, args)
}

func TestSelectDocScope(t *testing.T) {
	now := NullTimeFrom(time.Now())

	sql, args := SelectDoc("e", "f").
		From("matches m").
		Scope(`
			WHERE m.game_id = $1
				AND (
					m.id > $3
					OR (m.id >= $2 AND m.id <= $3 AND m.updated_at > $4)
				)
		`, 100, 1, 2, now).
		ToSQL()

	expected := `
		SELECT row_to_json(dat__item.*)
		FROM (
			SELECT e, f
			FROM matches m
			WHERE m.game_id=$1
				AND (
					m.id > $3
					OR (m.id >= $2 AND m.id<=$3 AND m.updated_at>$4)
				)
		) as dat__item
	`

	assert.Equal(t, stripWS(expected), stripWS(sql))
	assert.Equal(t, []interface{}{100, 1, 2, now}, args)
}
