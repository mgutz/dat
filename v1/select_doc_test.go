package dat

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
		As("f", `SELECT g, h FROM f WHERE id= $1`, 4).
		As("x", `SELECT id, y, z FROM x`).
		From("a").
		Where("d=$1", 4).
		ToSQL()

	expected := `
	SELECT row_to_json(dat__item.*)
	FROM (
		SELECT
			b,
			c,
			(SELECT array_agg(dat__f.*) FROM (SELECT g,h FROM f WHERE id=$1) AS dat__f) AS f,
			(SELECT array_agg(dat__x.*) FROM (SELECT id,y,z FROM x) AS dat__x) AS x
		FROM a
		WHERE (d=$2)
	) as dat__item
	`
	assert.Equal(t, stripWS(expected), stripWS(sql))
	assert.Equal(t, []interface{}{4, 4}, args)
}

func TestSelectDocSQLInnerSQL(t *testing.T) {
	sql, args := SelectDoc("b", "c").
		As("f", `SELECT g, h FROM f WHERE id= $1`, 4).
		As("x", `SELECT id, y, z FROM x`).
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
			(SELECT array_agg(dat__f.*) FROM (SELECT g,h FROM f WHERE id=$1) AS dat__f) AS f,
			(SELECT array_agg(dat__x.*) FROM (SELECT id,y,z FROM x) AS dat__x) AS x
		FROM a
		WHERE d=$2
	) as dat__item
	`
	assert.Equal(t, stripWS(expected), stripWS(sql))
	assert.Equal(t, []interface{}{4, 4}, args)
}

// func TestSelectDocSQLReturning(t *testing.T) {
// 	sql, args := SelectDoc("tab").Columns("b", "c").Values(1, 2).Where("d=$1", 4).Returning("f", "g").ToSQL()
// 	expected := `
// 	WITH
// 		upd AS (
// 			UPDATE tab
// 			SET "b" = $1, "c" = $2
// 			WHERE (d=$3)
// 			RETURNING "f","g"
// 		), ins AS (
// 			INSERT INTO "tab"("b","c")
// 			SELECT $1,$2
// 			WHERE NOT EXISTS (SELECT 1 FROM upd)
// 			RETURNING "f","g"
// 		)
// 	SELECT * FROM ins UNION ALL SELECT * FROM upd
// 	`

// 	assert.Equal(t, stripWS(expected), stripWS(sql))
// 	assert.Equal(t, []interface{}{1, 2, 4}, args)
// }

// func TestSelectDocSQLRecord(t *testing.T) {
// 	var rec = struct {
// 		B int
// 		C int
// 	}{1, 2}
// 	sql, args := SelectDoc("tab").
// 		Columns("b", "c").
// 		Record(rec).
// 		Where("d=$1", 4).
// 		Returning("f", "g").
// 		ToSQL()

// 	expected := `
// 	WITH
// 		upd AS (
// 			UPDATE tab
// 			SET "b" = $1, "c" = $2
// 			WHERE (d=$3)
// 			RETURNING "f","g"
// 		), ins AS (
// 			INSERT INTO "tab"("b","c")
// 			SELECT $1,$2
// 			WHERE NOT EXISTS (SELECT 1 FROM upd)
// 			RETURNING "f","g"
// 		)
// 	SELECT * FROM ins UNION ALL SELECT * FROM upd
// 	`

// 	assert.Equal(t, stripWS(expected), stripWS(sql))
// 	assert.Equal(t, []interface{}{1, 2, 4}, args)
// }
