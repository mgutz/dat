package runner

import (
	"testing"

	"github.com/mgutz/dat"
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

func BenchmarkExecSQLDat4(b *testing.B) {
	benchmarkInsertDatN(b, 1, 4)
}

func benchmarkInsertDatN(b *testing.B, rows int, argc int) {
	benchReset()
	builder, err := benchInsertBuilder(rows, argc)
	if err != nil {
		b.Fatal(err)
	}

	dat.EnableInterpolation = true
	sql, _, err := builder.Interpolate()
	if err != nil {
		b.Fatal(err)
	}
	dat.EnableInterpolation = false

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err = conn.DB.Exec(sql)
		if err != nil {
			b.Error(err.Error())
		}
	}
}
