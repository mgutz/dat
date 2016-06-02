package dat

import (
	"testing"
	"time"

	"github.com/mgutz/str"

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
			WHERE (m.game_id=$1
				AND (
					m.id > $3
					OR (m.id >= $2 AND m.id<=$3 AND m.updated_at>$4)
				))
		) as dat__item
	`

	assert.Equal(t, stripWS(expected), stripWS(sql))
	assert.Equal(t, []interface{}{100, 1, 2, now}, args)
}

func TestDocScopeWhere(t *testing.T) {
	published := `
		INNER JOIN posts p on (p.author_id = u.id)
		WHERE
			p.state = $1
	`
	sql, args := SelectDoc("u.*, p.*").
		From(`users u`).
		Scope(published, "published").
		Where(`u.id = $1`, 1).
		ToSQL()
	sql = str.Clean(sql)
	expected := `
		SELECT row_to_json(dat__item.*)
		FROM (
			SELECT u.*, p.*
			FROM users u
				INNER JOIN posts p on (p.author_id = u.id)
			WHERE (u.id = $1) AND ( p.state = $2 )
		) as dat__item
	`
	assert.Equal(t, stripWS(expected), stripWS(sql))
	assert.Exactly(t, args, []interface{}{1, "published"})
}

func TestDocDistinctOn(t *testing.T) {
	published := `
		INNER JOIN posts p on (p.author_id = u.id)
		WHERE
			p.state = $1
	`
	sql, args := SelectDoc("u.*, p.*").
		DistinctOn("aa", "bb").
		From(`users u`).
		Scope(published, "published").
		Where(`u.id = $1`, 1).
		ToSQL()
	expected := `
		SELECT row_to_json(dat__item.*)
		FROM (
			SELECT DISTINCT ON (aa, bb)
			u.*, p.*
			FROM users u
				INNER JOIN posts p on (p.author_id = u.id)
			WHERE (u.id = $1) AND ( p.state = $2 )
		) as dat__item
	`
	assert.Equal(t, stripWS(expected), stripWS(sql))
	assert.Exactly(t, args, []interface{}{1, "published"})
}

func TestNestedSelecDocWhere(t *testing.T) {
	user := SelectDoc("id", "user_name").
		Many("comments", `SELECT * FROM comments WHERE id = u.id`).
		From("users u").
		Where("u.id = $1", 1)

	sql, args := SelectDoc("id").
		One("user", user).
		From(`games`).
		Where(`id = $1`, 10).
		ToSQL()

	expected := `
		SELECT row_to_json(dat__item.*)
		FROM (
			SELECT id,
				(
					SELECT row_to_json(dat__user.*)
					FROM (
						SELECT id, user_name,
							(
								SELECT array_agg(dat__comments.*)
								FROM (SELECT * FROM comments WHERE id = u.id)
								AS dat__comments
							) AS "comments"
						FROM users u
						WHERE (u.id = $1)
					) AS dat__user
				) AS "user"
			FROM games
			WHERE (id = $2)
		) as dat__item
	`
	assert.Equal(t, stripWS(expected), stripWS(sql))
	assert.Exactly(t, args, []interface{}{1, 10})
}

func TestSelectDocColumns(t *testing.T) {
	sql, args := SelectDoc("id, user_name").
		From("users").
		Columns("created_at").
		ToSQL()
	assert.Equal(t, stripWS(`
		SELECT row_to_json(dat__item.*)
		FROM (
			SELECT id, user_name, created_at
			FROM users
		) as dat__item`), stripWS(sql))
	assert.Nil(t, args)
}

func TestSelectDocFor(t *testing.T) {
	sql, args := SelectDoc("id, user_name").
		From("users").
		Columns("created_at").
		For("UPDATE").
		ToSQL()
	assert.Equal(t, stripWS(`
		SELECT row_to_json(dat__item.*)
		FROM (
			SELECT id, user_name, created_at
			FROM users
			FOR UPDATE
		) as dat__item`), stripWS(sql))
	assert.Nil(t, args)
}
