package dat

import (
	"database/sql/driver"
	"testing"
	"time"

	"github.com/lib/pq"

	"gopkg.in/stretchr/testify.v1/assert"
)

func BenchmarkInterpolate(b *testing.B) {
	// Do some allocations outside the loop so they don't affect the results
	argEq1 := Eq{"f": 2, "x": "hi"}
	argEq2 := map[string]interface{}{"g": 3}
	argEq3 := Eq{"h": []int{1, 2, 3}}
	sq, args := Select("a", "b", "z", "y", "x").
		Distinct().
		From("c").
		Where("d = $1 OR e = $2", 1, "wat").
		Where(argEq1).
		Where(argEq2).
		Where(argEq3).
		GroupBy("i").
		GroupBy("ii").
		GroupBy("iii").
		Having("j = k").
		Having("jj = $1", 1).
		Having("jjj = $1", 2).
		OrderBy("l").
		OrderBy("l").
		OrderBy("l").
		Limit(7).
		Offset(8).
		ToSQL()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Interpolate(sq, args)
	}
}

func TestInterpolateNil(t *testing.T) {
	args := []interface{}{nil}

	str, _, err := Interpolate("SELECT * FROM x WHERE a = $1", args)
	assert.NoError(t, err)
	assert.Equal(t, str, "SELECT * FROM x WHERE a = NULL")
}

func TestInterpolateInts(t *testing.T) {
	args := []interface{}{
		int(1),
		int8(-2),
		int16(3),
		int32(4),
		int64(5),
		uint(6),
		uint8(7),
		uint16(8),
		uint32(9),
		uint64(10),
	}

	str, _, err := Interpolate("SELECT * FROM x WHERE a = $1 AND b = $2 AND c = $3 AND d = $4 AND e = $5 AND f = $6 AND g = $7 AND h = $8 AND i = $9 AND j = $1", args)
	assert.NoError(t, err)
	assert.Equal(t, str, "SELECT * FROM x WHERE a = 1 AND b = -2 AND c = 3 AND d = 4 AND e = 5 AND f = 6 AND g = 7 AND h = 8 AND i = 9 AND j = 1")
}

func TestInterpolateBools(t *testing.T) {
	args := []interface{}{true, false}

	str, _, err := Interpolate("SELECT * FROM x WHERE a = $1 AND b = $2", args)
	assert.NoError(t, err)
	assert.Equal(t, str, "SELECT * FROM x WHERE a = 't' AND b = 'f'")
}

func TestInterpolateFloats(t *testing.T) {
	args := []interface{}{float32(0.15625), float64(3.14159)}

	str, _, err := Interpolate("SELECT * FROM x WHERE a = $1 AND b = $2", args)
	assert.NoError(t, err)
	assert.Equal(t, str, "SELECT * FROM x WHERE a = 0.15625 AND b = 3.14159")
}

func TestInterpolateEscapeStrings(t *testing.T) {
	args := []interface{}{"hello", "\"pg's world\" \\\b\f\n\r\t\x1a"}

	str, _, err := Interpolate("SELECT * FROM x WHERE a = $1 AND b = $2", args)
	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM x WHERE a = 'hello' AND b = '\"pg''s world\" \\\b\f\n\r\t\x1a'", str)
}

func TestInterpolateSlices(t *testing.T) {
	args := []interface{}{[]int{1}, []int{1, 2, 3}, []uint32{5, 6, 7}, []string{"wat", "ok"}}

	str, _, err := Interpolate("SELECT * FROM x WHERE a = $1 AND b = $2 AND c = $3 AND d = $4", args)
	assert.NoError(t, err)
	assert.Equal(t, str, "SELECT * FROM x WHERE a = (1) AND b = (1,2,3) AND c = (5,6,7) AND d = ('wat','ok')")
}

type myString struct {
	Present bool
	Val     string
}

func (m myString) Value() (driver.Value, error) {
	if m.Present {
		return m.Val, nil
	}
	return nil, nil
}

func TestIntepolatingValuers(t *testing.T) {
	args := []interface{}{myString{true, "wat"}, myString{false, "fry"}}

	str, _, err := Interpolate("SELECT * FROM x WHERE a = $1 AND b = $2", args)
	assert.NoError(t, err)
	assert.Equal(t, str, "SELECT * FROM x WHERE a = 'wat' AND b = NULL")
}

func TestInterpolatingUnsafeStrings(t *testing.T) {
	args := []interface{}{NOW, DEFAULT, UnsafeString(`hstore`)}
	str, _, err := Interpolate("SELECT * FROM x WHERE one=$1 AND two=$2 AND three=$3", args)
	assert.NoError(t, err)
	assert.Equal(t, str, "SELECT * FROM x WHERE one=NOW() AND two=DEFAULT AND three=hstore")
}

func TestInterpolatingPointers(t *testing.T) {
	var one int32 = 1000
	var two int64 = 2000
	var three float32 = 3
	var four float64 = 4
	var five = "five"
	var six = true

	args := []interface{}{&one, &two, &three, &four, &five, &six}
	str, _, err := Interpolate("SELECT * FROM x WHERE one=$1 AND two=$2 AND three=$3 AND four=$4 AND five=$5 AND six=$6", args)
	assert.NoError(t, err)
	assert.Equal(t, str, "SELECT * FROM x WHERE one=1000 AND two=2000 AND three=3 AND four=4 AND five='five' AND six='t'")
}

func TestInterpolatingNulls(t *testing.T) {
	var one *int32
	var two *int64
	var three *float32
	var four *float64
	var five *string
	var six *bool

	args := []interface{}{one, two, three, four, five, six}
	str, _, err := Interpolate("SELECT * FROM x WHERE one=$1 AND two=$2 AND three=$3 AND four=$4 AND five=$5 AND six=$6", args)
	assert.NoError(t, err)
	assert.Equal(t, str, "SELECT * FROM x WHERE one=NULL AND two=NULL AND three=NULL AND four=NULL AND five=NULL AND six=NULL")
}

func TestInterpolatingTime(t *testing.T) {
	var ptim *time.Time
	tim2 := time.Date(2004, time.January, 1, 1, 1, 1, 1, time.UTC)
	tim := time.Time{}

	args := []interface{}{ptim, tim, &tim2}

	str, _, err := Interpolate("SELECT * FROM x WHERE a = $1 AND b = $2 AND c = $3", args)
	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM x WHERE a = NULL AND b = '0001-01-01T00:00:00Z' AND c = '2004-01-01T01:01:01.000000001Z'", str)
}

func TestInterpolateErrors(t *testing.T) {
	_, _, err := Interpolate("SELECT * FROM x WHERE a = $1 AND b = $2", []interface{}{1})
	assert.Equal(t, err, ErrArgumentMismatch)

	// no harm, no foul
	if Strict {
		_, _, err = Interpolate("SELECT * FROM x WHERE", []interface{}{1})
		assert.Equal(t, err, ErrArgumentMismatch)
	}

	_, _, err = Interpolate("SELECT * FROM x WHERE a = $1", []interface{}{string([]byte{0x34, 0xFF, 0xFE})})
	assert.Equal(t, err, ErrNotUTF8)

	_, _, err = Interpolate("SELECT * FROM x WHERE a = $1", []interface{}{struct{}{}})
	assert.Equal(t, err, ErrInvalidValue)

	_, _, err = Interpolate("SELECT * FROM x WHERE a = $1", []interface{}{[]struct{}{{}, {}}})
	assert.Equal(t, err, ErrInvalidSliceValue)
}

func TestInterpolateJSON(t *testing.T) {
	j, _ := NewJSON([]int{1, 3, 10})
	sql, args, err := Interpolate("SELECT $1", []interface{}{j})
	assert.NoError(t, err)
	assert.Equal(t, "SELECT '[1,3,10]'", sql)
	assert.Equal(t, 0, len(args))
}

func TestInterpolateInvalidNullTime(t *testing.T) {
	invalid := NullTime{pq.NullTime{Valid: false}}

	sql, _, err := Interpolate("SELECT * FROM foo WHERE invalid = $1", []interface{}{invalid})
	assert.NoError(t, err)
	assert.Equal(t, stripWS("SELECT * FROM foo WHERE invalid=NULL"), stripWS(sql))
}

func TestInterpolateValidNullTime(t *testing.T) {
	now := time.Now()
	valid := NullTime{pq.NullTime{Time: now, Valid: true}}
	sql, _, err := Interpolate("SELECT * FROM foo WHERE valid = $1", []interface{}{valid})
	assert.NoError(t, err)

	assert.Equal(t, "SELECT * FROM foo WHERE valid = '"+valid.Time.Format(time.RFC3339Nano)+"'", sql)
}

func TestInterpolateNonPlaceholdersA(t *testing.T) {
	sql, _, err := Interpolate("$ $$ $aa $1 $", []interface{}{"value"})
	assert.NoError(t, err)
	assert.Equal(t, "$ $$ $aa 'value' $", sql)

	sql, _, err = Interpolate("$ $1$ $aa $1", []interface{}{"value"})
	assert.NoError(t, err)
	assert.Equal(t, "$ 'value'$ $aa 'value'", sql)
}

func TestInterpolateExpression(t *testing.T) {
	// the following case statement does not work with enums in straight SQL
	// but with Expression we can use composition
	//
	// case when $1 = '' then true else kind = $1 end
	exp := func(value string) *Expression {
		if value == "" {
			return Expr("true")
		}
		return Expr("kind = $1", "apple")
	}

	var nullExp *Expression

	sql, _, err := Interpolate("select * from fruits where $1 and $2 and $3", []interface{}{exp(""), exp("apple"), nullExp})
	assert.NoError(t, err)
	assert.Equal(t, "select * from fruits where true and kind = 'apple' and NULL", sql)
}
