package runner

import (
	"testing"

	"gopkg.in/mgutz/dat.v1"
)

// These benchmarks compare the cost of interpolating the SQL then executing
// the query against non interpolated queries using database/sql and jmoiron/sqlx.

func BenchmarkBuildExecSQLDat2(b *testing.B) {
	benchmarkBuildInsertDatN(b, 1, 2)
}

func BenchmarkBuildExecSQLSql2(b *testing.B) {
	benchmarkBuildInsertSQLN(b, 1, 2)
}

func BenchmarkBuildExecSQLDat4(b *testing.B) {
	benchmarkBuildInsertDatN(b, 1, 4)
}

func BenchmarkBuildExecSQLSql4(b *testing.B) {
	benchmarkBuildInsertSQLN(b, 1, 4)
}

func BenchmarkBuildExecSQLDat8(b *testing.B) {
	benchmarkBuildInsertDatN(b, 2, 4)
}

func BenchmarkBuildExecSQLSql8(b *testing.B) {
	benchmarkBuildInsertSQLN(b, 2, 4)
}

func BenchmarkBuildExecSQLDat64(b *testing.B) {
	benchmarkBuildInsertDatN(b, 16, 4)
}

func BenchmarkBuildExecSQLSql64(b *testing.B) {
	benchmarkBuildInsertSQLN(b, 16, 4)
}

func benchmarkBuildInsertDatN(b *testing.B, rows int, argc int) {
	benchReset()
	builder, err := benchInsertBuilder(rows, argc)
	if err != nil {
		b.Fatal(err)
	}
	sql, args := builder.ToSQL()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sql2, args2, err := dat.Interpolate(sql, args)
		if err != nil {
			b.Fatal(err)
		}
		_, err = testDB.Exec(sql2, args2...)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkBuildInsertSQLN(b *testing.B, rows int, argc int) {
	benchReset()
	builder, err := benchInsertBuilder(rows, argc)
	if err != nil {
		b.Fatal(err)
	}

	sql, args := builder.ToSQL()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := testDB.Exec(sql, args...)
		if err != nil {
			b.Error(err.Error())
		}
	}
}

func benchmarkBuildInsertSqlxN(b *testing.B, rows int, argc int) {
	benchReset()
	builder, err := benchInsertBuilder(rows, argc)
	if err != nil {
		b.Fatal(err)
	}
	sql, args := builder.ToSQL()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := testDB.DB.Exec(sql, args...)
		if err != nil {
			b.Error(err.Error())
		}
	}
}
