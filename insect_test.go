package dat

import (
	"regexp"
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"
)

var regexpWS = regexp.MustCompile(`\s`)

func stripWS(s string) string {
	return regexpWS.ReplaceAllLiteralString(s, "")
}

func TestInsectSqlSimple(t *testing.T) {
	sql, args := Insect("tab").Columns("b", "c").Values(1, 2).ToSQL()

	expected := `
		WITH
			sel AS (SELECT b, c FROM tab WHERE (b = $1) AND (c = $2)),
			ins AS (
				INSERT INTO "tab"("b","c")
				SELECT $1, $2
				WHERE NOT EXISTS (SELECT 1 FROM sel)
				RETURNING "b","c"
			)
		SELECT * FROM ins UNION ALL SELECT * FROM sel
	`
	assert.Equal(t, stripWS(expected), stripWS(sql))
	assert.Equal(t, args, []interface{}{1, 2})
}

func TestInsectSqlWhere(t *testing.T) {
	sql, args := Insect("tab").
		Columns("b", "c").
		Values(1, 2).
		Where("d = $1", 3).
		ToSQL()

	expected := `
	WITH
		sel AS (SELECT b, c FROM tab WHERE (d = $1)),
		ins AS (
			INSERT INTO "tab"("b","c")
			SELECT $2, $3
			WHERE NOT EXISTS (SELECT 1 FROM sel)
			RETURNING "b", "c"
		)
	SELECT * FROM ins UNION ALL SELECT * FROM sel
	`
	assert.Equal(t, stripWS(expected), stripWS(sql))
	assert.Equal(t, args, []interface{}{3, 1, 2})
}

func TestInsectSqlReturning(t *testing.T) {
	sql, args := Insect("tab").
		Columns("b", "c").
		Values(1, 2).
		Where("d = $1", 3).
		Returning("id", "f", "g").
		ToSQL()

	expected := `
	WITH
		sel AS (SELECT id, f, g FROM tab WHERE (d = $1)),
		ins AS (
			INSERT INTO "tab"("b","c")
			SELECT $2,$3
			WHERE NOT EXISTS (SELECT 1 FROM sel)
			RETURNING "id","f","g"
		)
	SELECT * FROM ins UNION ALL SELECT * FROM sel
	`
	assert.Equal(t, stripWS(expected), stripWS(sql))
	assert.Equal(t, args, []interface{}{3, 1, 2})
}

func TestInsectSqlRecord(t *testing.T) {
	type Embedded struct {
		D int `db:"d"`
	}
	var rec = struct {
		B int `db:"b"`
		C int `db:"c"`
		Embedded
	}{1, 2, Embedded{4}}

	sql, args := Insect("tab").
		Columns("b", "c", "d").
		Record(rec).
		Where("d = $1", 3).
		Returning("id", "f", "g").
		ToSQL()

	expected := `
	WITH
		sel AS (SELECT id, f, g FROM tab WHERE (d = $1)),
		ins AS (
			INSERT INTO "tab"("b","c","d")
			SELECT $2, $3, $4
			WHERE NOT EXISTS (SELECT 1 FROM sel)
			RETURNING "id","f","g"
		)
	SELECT * FROM ins UNION ALL SELECT * FROM sel
	`
	assert.Equal(t, stripWS(expected), stripWS(sql))
	assert.Equal(t, args, []interface{}{3, 1, 2, 4})
}
