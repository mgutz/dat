package dat

import (
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"
)

func TestUpsertSQLMissingWhere(t *testing.T) {
	_, _, err := Upsert("tab").Columns("b", "c").Values(1, 2).ToSQL()
	assert.Error(t, err)
}

func TestUpsertSQLWhere(t *testing.T) {
	sql, args, err := Upsert("tab").Columns("b", "c").Values(1, 2).Where("d=$1", 4).ToSQL()
	assert.NoError(t, err)
	expected := `
	WITH
		upd AS (
			UPDATE tab
			SET b = $1, c = $2
			WHERE (d=$3)
			RETURNING b,c
		), ins AS (
			INSERT INTO tab(b,c)
			SELECT $1,$2
			WHERE NOT EXISTS (SELECT 1 FROM upd)
			RETURNING b,c
		)
	SELECT * FROM ins UNION ALL SELECT * FROM upd
	`

	assert.Equal(t, stripWS(expected), stripWS(sql))
	assert.Equal(t, []interface{}{1, 2, 4}, args)
}

func TestUpsertSQLReturning(t *testing.T) {
	sql, args, err := Upsert("tab").Columns("b", "c").Values(1, 2).Where("d=$1", 4).Returning("f", "g").ToSQL()
	assert.NoError(t, err)
	expected := `
	WITH
		upd AS (
			UPDATE tab
			SET b = $1, c = $2
			WHERE (d=$3)
			RETURNING f,g
		), ins AS (
			INSERT INTO tab(b,c)
			SELECT $1,$2
			WHERE NOT EXISTS (SELECT 1 FROM upd)
			RETURNING f,g
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
	us := Upsert("tab").
		Columns("b", "c").
		Record(rec).
		Where("d=$1", 4).
		Returning("f", "g")
	sql, args, err := us.ToSQL()
	assert.NoError(t, err)

	expected := `
	WITH
		upd AS (
			UPDATE tab
			SET b = $1, c = $2
			WHERE (d=$3)
			RETURNING f,g
		), ins AS (
			INSERT INTO tab(b,c)
			SELECT $1,$2
			WHERE NOT EXISTS (SELECT 1 FROM upd)
			RETURNING f,g
		)
	SELECT * FROM ins UNION ALL SELECT * FROM upd
	`

	assert.Equal(t, stripWS(expected), stripWS(sql))
	assert.Equal(t, []interface{}{1, 2, 4}, args)
}

func TestUpsertSQLBlacklistRecord(t *testing.T) {
	var rec = struct {
		B int `db:"b"`
		C int `db:"c"`
	}{1, 2}
	us := Upsert("tab").
		Blacklist("c").
		Record(rec).
		Where("d=$1", 4).
		Returning("f", "g")
	us.ToSQL()
	sql, args, err := us.ToSQL()
	assert.NoError(t, err)

	expected := `
	WITH
		upd AS (
			UPDATE tab
			SET b = $1
			WHERE (d=$2)
			RETURNING f,g
		), ins AS (
			INSERT INTO tab(b)
			SELECT $1
			WHERE NOT EXISTS (SELECT 1 FROM upd)
			RETURNING f,g
		)
	SELECT * FROM ins UNION ALL SELECT * FROM upd
	`

	assert.Equal(t, stripWS(expected), stripWS(sql))
	assert.Equal(t, []interface{}{1, 4}, args)
}
