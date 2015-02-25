package runner

import (
	"testing"

	"github.com/mgutz/dat"
)

// These benchmarks compare the cost of interpolating the SQL then executing
// the query against non interpolated queries using database/sql and jmoiron/sqlx.

func BenchmarkBuildExecSQLDat2(b *testing.B) {
	benchmarkBuildInsertDatN(b, 1, 2)
}

func BenchmarkBuildExecSQLDat4(b *testing.B) {
	benchmarkBuildInsertDatN(b, 1, 4)
}

func BenchmarkBuildExecSQLDat8(b *testing.B) {
	benchmarkBuildInsertDatN(b, 2, 4)
}

func benchmarkBuildInsertDatN(b *testing.B, rows int, argc int) {
	benchReset()
	builder, err := benchInsertBuilder(rows, argc)
	if err != nil {
		b.Fatal(err)
	}

	dat.EnableInterpolation = true
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = conn.ExecBuilder(builder)
		if err != nil {
			b.Error(err.Error())
		}
	}
	dat.EnableInterpolation = false
}
