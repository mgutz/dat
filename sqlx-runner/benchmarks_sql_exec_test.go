package runner

import (
	"testing"

	"github.com/syreclabs/dat"
)

// These benchmarks compare the time to excute an interpolated SQL
// statement with zero args to that of executing a SQL statement with
// uninterpolated args.
//
// This test isn't that meaningful. It doesn't take into account
// the time to perform the interpolation. It's only meant to show that
// interpolated queries skip the prepare statement logic in the underlying
// driver.

func BenchmarkExecSQLDat2(b *testing.B) {
	benchmarkInsertDatN(b, 1, 2)
}

func BenchmarkExecSQLSql2(b *testing.B) {
	benchmarkInsertSqlN(b, 1, 2)
}

func BenchmarkExecSQLDat4(b *testing.B) {
	benchmarkInsertDatN(b, 1, 4)
}

func BenchmarkExecSQLSql4(b *testing.B) {
	benchmarkInsertSqlN(b, 1, 4)
}

func benchmarkInsertDatN(b *testing.B, rows int, argc int) {
	benchReset()
	builder, err := benchInsertBuilder(rows, argc)
	if err != nil {
		b.Fatal(err)
	}
	sql, args := builder.ToSQL()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sql2, args2, err := dat.Interpolate(sql, args)
		if err != nil {
			b.Fatal(err)
		}
		_, err = testDB.Exec(sql2, args2...)
		if err != nil {
			b.Error(err)
		}
	}
}

func benchmarkInsertSqlN(b *testing.B, rows int, argc int) {
	benchReset()
	builder, err := benchInsertBuilder(rows, argc)
	if err != nil {
		b.Fatal(err)
	}
	sql, args := builder.ToSQL()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := testDB.Exec(sql, args...)
		if err != nil {
			b.Error(err.Error())
		}
	}
}

func benchmarkInsertSqlxN(b *testing.B, rows int, argc int) {
	benchReset()
	builder, err := benchInsertBuilder(rows, argc)
	if err != nil {
		b.Fatal(err)
	}

	sql, args := builder.ToSQL()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := testDB.DB.Exec(sql, args...)
		if err != nil {
			b.Error(err.Error())
		}
	}
}

// benchInsertBuilders builds an insert statement with
// many values.
//
// INSERT INTO(benches)
// VALUES (row0), (row1), ... (rown-1)
func benchInsertBuilder(rows int, argc int) (*dat.InsertBuilder, error) {
	if argc > 4 {
		panic("args must be <= 4")
	}

	columns := []string{"amount", "name", "n", "is_ok"}
	values := []interface{}{42.0, "foo", 42, "true"}
	builder := dat.
		NewInsertBuilder("benches").
		Whitelist(columns[0:argc]...)

	// fill image with random bytes
	maxImage := 256
	image := make([]byte, maxImage)
	for i := 0; i < maxImage; i++ {
		image[i] = byte(i % 256)
	}

	for i := 0; i < rows; i++ {
		builder.Values(values[0:argc]...)
	}
	return builder, nil
}
