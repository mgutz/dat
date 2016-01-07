package dat

import (
	"strings"
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"
)

type missingDbTag struct {
	ID int64
}

type someRecord struct {
	SomethingID int   `db:"something_id"`
	UserID      int64 `db:"user_id"`
	Other       bool  `db:"other"`
}

func BenchmarkInsertValuesSql(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		InsertInto("alpha").Columns("something_id", "user_id", "other").Values(1, 2, true).ToSQL()
	}
}

func BenchmarkInsertRecordsSql(b *testing.B) {
	obj := someRecord{1, 99, false}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		InsertInto("alpha").Columns("something_id", "user_id", "other").Record(obj).ToSQL()
	}
}

func TestInsertSingleToSql(t *testing.T) {
	sql, args := InsertInto("a").Columns("b", "c").Values(1, 2).ToSQL()

	assert.Equal(t, sql, quoteSQL("INSERT INTO a (%s,%s) VALUES ($1,$2)", "b", "c"))
	assert.Equal(t, args, []interface{}{1, 2})
}

func TestDefaultValue(t *testing.T) {
	sql, args := InsertInto("a").Columns("b", "c").Values(1, DEFAULT).ToSQL()

	assert.Equal(t, sql, quoteSQL("INSERT INTO a (%s,%s) VALUES ($1,$2)", "b", "c"))
	assert.Equal(t, args, []interface{}{1, DEFAULT})
}

func TestInsertMultipleToSql(t *testing.T) {
	sql, args := InsertInto("a").Columns("b", "c").Values(1, 2).Values(3, 4).ToSQL()

	assert.Equal(t, sql, quoteSQL("INSERT INTO a (%s,%s) VALUES ($1,$2),($3,$4)", "b", "c"))
	assert.Equal(t, args, []interface{}{1, 2, 3, 4})
}

func TestInsertRecordsToSql(t *testing.T) {
	objs := []someRecord{{1, 88, false}, {2, 99, true}}
	sql, args := InsertInto("a").Columns("something_id", "user_id", "other").Record(objs[0]).Record(objs[1]).ToSQL()

	assert.Equal(t, sql, quoteSQL("INSERT INTO a (%s,%s,%s) VALUES ($1,$2,$3),($4,$5,$6)", "something_id", "user_id", "other"))
	checkSliceEqual(t, args, []interface{}{1, 88, false, 2, 99, true})
}

func TestInsertWhitelist(t *testing.T) {
	objs := []someRecord{{1, 88, false}, {2, 99, true}}
	sql, args := InsertInto("a").
		Whitelist("*").
		Record(objs[0]).
		Record(objs[1]).
		ToSQL()
	assert.Equal(t, sql, quoteSQL("INSERT INTO a (%s,%s,%s) VALUES ($1,$2,$3),($4,$5,$6)", "something_id", "user_id", "other"))
	checkSliceEqual(t, []interface{}{1, 88, false, 2, 99, true}, args)

	assert.Panics(t, func() {
		InsertInto("a").Whitelist("*").Values("foo").ToSQL()
	}, `must use "*" in conjunction with Record`)
}

func TestInsertBlacklist(t *testing.T) {
	objs := []someRecord{{1, 88, false}, {2, 99, true}}
	sql, args := InsertInto("a").
		Blacklist("something_id").
		Record(objs[0]).
		Record(objs[1]).
		ToSQL()
	// order is not guaranteed
	//assert.Equal(t, sql, `INSERT INTO a ("user_id","other") VALUES ($1,$2),($3,$4)`)
	assert.True(t, strings.Contains(sql, `"user_id"`))
	assert.True(t, strings.Contains(sql, `"other"`))
	checkSliceEqual(t, args, []interface{}{88, false, 99, true})

	assert.Panics(t, func() {
		// does not have any columns or record
		InsertInto("a").Blacklist("something_id").Values("foo").ToSQL()
	})
}

func TestInsertDuplicateColumns(t *testing.T) {
	type A struct {
		Status string `db:"status"`
	}

	type B struct {
		Status string `db:"status"`
		A
	}

	b := B{}
	b.Status = "open"
	b.A.Status = "closed"
	sql, args := InsertInto("a").Columns("status").Record(&b).ToSQL()
	assert.Equal(t, sql, `INSERT INTO a ("status") VALUES ($1)`)
	assert.Equal(t, args, []interface{}{"open"})
}
