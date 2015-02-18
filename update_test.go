package dat

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

	assert.Equal(t, sql, quoteSQL("UPDATE a SET %s = $1, %s = $2", "b", "c"))
	assert.Equal(t, args, []interface{}{1, 2})
}

func TestUpdateSingleToSql(t *testing.T) {
	sql, args := Update("a").Set("b", 1).Set("c", 2).Where("id = $1", 1).ToSQL()

	assert.Equal(t, sql, quoteSQL("UPDATE a SET %s = $1, %s = $2 WHERE (id = $3)", "b", "c"))
	assert.Equal(t, args, []interface{}{1, 2, 1})
}

func TestUpdateSetMapToSql(t *testing.T) {
	sql, args := Update("a").SetMap(map[string]interface{}{"b": 1, "c": 2}).Where("id = $1", 1).ToSQL()

	if sql == quoteSQL("UPDATE a SET %s = $1, %s = $2 WHERE (id = $3)", "b", "c") {
		assert.Equal(t, args, []interface{}{1, 2, 1})
	} else {
		assert.Equal(t, sql, quoteSQL("UPDATE a SET %s = $1, %s = $2 WHERE (id = $3)", "c", "b"))
		assert.Equal(t, args, []interface{}{2, 1, 1})
	}
}

func TestUpdateSetExprToSql(t *testing.T) {
	sql, args := Update("a").Set("foo", 1).Set("bar", Expr("COALESCE(bar, 0) + 1")).Where("id = $1", 9).ToSQL()

	assert.Equal(t, sql, quoteSQL("UPDATE a SET %s = $1, %s = COALESCE(bar, 0) + 1 WHERE (id = $2)", "foo", "bar"))
	assert.Equal(t, args, []interface{}{1, 9})

	sql, args = Update("a").Set("foo", 1).Set("bar", Expr("COALESCE(bar, 0) + $1", 2)).Where("id = $2", 9).ToSQL()

	assert.Equal(t, sql, quoteSQL("UPDATE a SET %s = $1, %s = COALESCE(bar, 0) + $2 WHERE (id = $3)", "foo", "bar"))
	assert.Equal(t, args, []interface{}{1, 2, 9})
}

func TestUpdateTenStaringFromTwentyToSql(t *testing.T) {
	sql, args := Update("a").Set("b", 1).Limit(10).Offset(20).ToSQL()

	assert.Equal(t, sql, quoteSQL("UPDATE a SET %s = $1 LIMIT 10 OFFSET 20", "b"))
	assert.Equal(t, args, []interface{}{1})
}
