package dat

import (
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"
)

func TestUpsertSQLMissingWhere(t *testing.T) {
	assert.Panics(t, func() {
		Upsert("tab").Columns("b", "c").Values(1, 2).ToSQL()
	})
}

func TestUpsertSQLWhere(t *testing.T) {
	sql, args := Upsert("tab").Columns("b", "c").Values(1, 2).Where("d=$1", 4).ToSQL()
	expected := `
	WITH
		upd AS (
			UPDATE "tab"
			SET "b" = $1, "c" = $2
			WHERE (d=$3)
			RETURNING "b","c"
		), ins AS (
			INSERT INTO "tab"("b","c")
			SELECT $1,$2
			WHERE NOT EXISTS (SELECT 1 FROM upd)
			RETURNING "b","c"
		)
	SELECT * FROM ins UNION ALL SELECT * FROM upd
	`

	assert.Equal(t, stripWS(expected), stripWS(sql))
	assert.Equal(t, []interface{}{1, 2, 4}, args)
}

func TestUpsertSQLReturning(t *testing.T) {
	sql, args := Upsert("tab").Columns("b", "c").Values(1, 2).Where("d=$1", 4).Returning("f", "g").ToSQL()
	expected := `
	WITH
		upd AS (
			UPDATE "tab"
			SET "b" = $1, "c" = $2
			WHERE (d=$3)
			RETURNING "f","g"
		), ins AS (
			INSERT INTO "tab"("b","c")
			SELECT $1,$2
			WHERE NOT EXISTS (SELECT 1 FROM upd)
			RETURNING "f","g"
		)
	SELECT * FROM ins UNION ALL SELECT * FROM upd
	`

	assert.Equal(t, stripWS(expected), stripWS(sql))
	assert.Equal(t, []interface{}{1, 2, 4}, args)
}

func TestUpsertSQLRecord(t *testing.T) {
	var rec = struct {
		B int `db:"b"`
		C int `db:"c"`
	}{1, 2}

	sql, args := Upsert("tab").
		Columns("b", "c").
		Record(rec).
		Where("d=$1", 4).
		Returning("f", "g").
		ToSQL()

	expected := `
	WITH
		upd AS (
			UPDATE "tab"
			SET "b" = $1, "c" = $2
			WHERE (d=$3)
			RETURNING "f","g"
		), ins AS (
			INSERT INTO "tab"("b","c")
			SELECT $1,$2
			WHERE NOT EXISTS (SELECT 1 FROM upd)
			RETURNING "f","g"
		)
	SELECT * FROM ins UNION ALL SELECT * FROM upd
	`

	assert.Equal(t, stripWS(expected), stripWS(sql))
	assert.Equal(t, []interface{}{1, 2, 4}, args)
}
