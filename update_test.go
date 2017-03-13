package dat

import (
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"
	"fmt"
	"strings"
)

func BenchmarkUpdateValuesSql(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Update("alpha").Set("something_id", 1).Where("id", 1).ToSQL()
	}
}

func BenchmarkUpdateValueMapSql(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Update("alpha").Set("something_id", 1).SetMap(map[string]interface{}{"b": 1, "c": 2}).Where("id", 1).ToSQL()
	}
}

func TestUpdateAllToSql(t *testing.T) {
	sql, args := Update("a").Set("b", 1).Set("c", 2).ToSQL()

	assert.Equal(t, quoteSQL(`UPDATE "a" SET %s = $1, %s = $2`, "b", "c"), sql)
	assert.Equal(t, []interface{}{1, 2}, args)
}

func TestUpdateSingleToSql(t *testing.T) {
	sql, args := Update("a").Set("b", 1).Set("c", 2).Where("id = $1", 1).ToSQL()

	assert.Equal(t, quoteSQL(`UPDATE "a" SET %s = $1, %s = $2 WHERE (id = $3)`, "b", "c"), sql)
	assert.Equal(t, []interface{}{1, 2, 1}, args)
}

func TestUpdateSetMapToSql(t *testing.T) {
	sql, args := Update("a").SetMap(map[string]interface{}{"b": 1, "c": 2}).Where("id = $1", 1).ToSQL()

	if sql == quoteSQL(`UPDATE "a" SET %s = $1, %s = $2 WHERE (id = $3)`, "b", "c") {
		assert.Equal(t, []interface{}{1, 2, 1}, args)
	} else {
		assert.Equal(t, quoteSQL(`UPDATE "a" SET %s = $1, %s = $2 WHERE (id = $3)`, "c", "b"), sql)
		assert.Equal(t, []interface{}{2, 1, 1}, args)
	}
}

func TestUpdateSetExprToSql(t *testing.T) {
	sql, args := Update("a").Set("foo", 1).Set("bar", Expr("COALESCE(bar, 0) + 1")).Where("id = $1", 9).ToSQL()

	assert.Equal(t, quoteSQL(`UPDATE "a" SET %s = $1, %s = COALESCE(bar, 0) + 1 WHERE (id = $2)`, "foo", "bar"), sql)
	assert.Equal(t, []interface{}{1, 9}, args)

	sql, args = Update("a").Set("foo", 1).Set("bar", Expr("COALESCE(bar, 0) + $1", 2)).Where("id = $1", 9).ToSQL()

	assert.Equal(t, quoteSQL(`UPDATE "a" SET %s = $1, %s = COALESCE(bar, 0) + $2 WHERE (id = $3)`, "foo", "bar"), sql)
	assert.Equal(t, []interface{}{1, 2, 9}, args)
}

func TestUpdateTenStaringFromTwentyToSql(t *testing.T) {
	sql, args := Update("a").Set("b", 1).Limit(10).Offset(20).ToSQL()

	assert.Equal(t, quoteSQL(`UPDATE "a" SET %s = $1 LIMIT 10 OFFSET 20`, "b"), sql)
	assert.Equal(t, []interface{}{1}, args)
}

func TestUpdateWhitelist(t *testing.T) {
	// type someRecord struct {
	// 	SomethingID int   `db:"something_id"`
	// 	UserID      int64 `db:"user_id"`
	// 	Other       bool  `db:"other"`
	// }
	sr := &someRecord{1, 2, false}
	sql, args := Update("a").
		SetWhitelist(sr, "user_id", "other").
		ToSQL()

	assert.Equal(t, quoteSQL(`UPDATE "a" SET %s = $1, %s = $2`, "user_id", "other"), sql)
	checkSliceEqual(t, []interface{}{2, false}, args)
}

func TestUpdateBlacklist(t *testing.T) {
	sr := &someRecord{1, 2, false}
	sql, args := Update("a").
		SetBlacklist(sr, "something_id").
		ToSQL()

	assert.Equal(t, quoteSQL(`UPDATE "a" SET %s = $1, %s = $2`, "user_id", "other"), sql)
	checkSliceEqual(t, []interface{}{2, false}, args)
}

func TestUpdateWhereExprSql(t *testing.T) {
	expr := Expr("id=$1", 100)
	sql, args := Update("a").Set("b", 10).Where(expr).ToSQL()
	assert.Equal(t, `UPDATE "a" SET "b" = $1 WHERE (id=$2)`, sql)
	assert.Exactly(t, []interface{}{10, 100}, args)
}

func TestUpdateBeyondMaxLookup(t *testing.T) {
	sqlBuilder := Update("a")
	setClauses := []string{}
	expectedArgs := []interface{}{}
	for i := 1; i < maxLookup + 1; i++ {
		sqlBuilder = sqlBuilder.Set("b", i)
		setClauses = append(setClauses, fmt.Sprintf(" %s = $%d", quoteSQL("%s", "b"), i))
		expectedArgs = append(expectedArgs, i)
	}
	sql, args := sqlBuilder.Where("id = $1", maxLookup + 1).ToSQL()
	expectedSQL := fmt.Sprintf(`UPDATE "a" SET%s WHERE (id = $%d)`, strings.Join(setClauses, ","), maxLookup + 1)
	expectedArgs = append(expectedArgs, maxLookup + 1)

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)

}
