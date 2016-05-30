package runner

import (
	"testing"

	"github.com/syreclabs/dat"
)

// These benchmarks compare the total cost of interpolating the SQL then
// executing the query on the same connection using a transaction.
// Both database/sql and jmoiron/sqlx can take advantage of a prepared
// statements.

func BenchmarkTransactedDat2(b *testing.B) {
	benchmarkTransactedDatN(b, 1, 2)
}

func BenchmarkTransactedSql2(b *testing.B) {
	benchmarkTransactedSqlN(b, 1, 2)
}

func BenchmarkTransactedDat4(b *testing.B) {
	benchmarkTransactedDatN(b, 1, 4)
}

func BenchmarkTransactedSql4(b *testing.B) {
	benchmarkTransactedSqlN(b, 1, 4)
}

func BenchmarkTransactedDat8(b *testing.B) {
	benchmarkTransactedDatN(b, 2, 4)
}

func BenchmarkTransactedSql8(b *testing.B) {
	benchmarkTransactedSqlN(b, 2, 4)
}

func BenchmarkTransactedDat64(b *testing.B) {
	benchmarkTransactedDatN(b, 16, 4)
}

func BenchmarkTransactedSql64(b *testing.B) {
	benchmarkTransactedSqlN(b, 16, 4)
}

func benchmarkTransactedDatN(b *testing.B, rows int, argc int) {
	benchReset()
	builder, err := benchInsertBuilder(rows, argc)
	if err != nil {
		b.Fatal(err)
	}
	sql, args := builder.ToSQL()

	tx, err := testDB.Begin()
	if err != nil {
		b.Fatal(err)
	}
	defer tx.Commit()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sql2, args2, err := dat.Interpolate(sql, args)
		if err != nil {
			b.Fatal(err)
		}
		_, err = tx.Exec(sql2, args2...)
		if err != nil {
			//fmt.Println(builder)
			b.Fatal(err)
		}
	}
}

func benchmarkTransactedSqlN(b *testing.B, rows int, argc int) {
	benchReset()
	builder, err := benchInsertBuilder(rows, argc)
	if err != nil {
		b.Fatal(err)
	}

	sql, args := builder.ToSQL()

	tx, err := testDB.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tx.Exec(sql, args...)
		if err != nil {
			b.Error(err.Error())
		}
	}
}

func benchmarkTransactedSqlxN(b *testing.B, rows int, argc int) {
	benchReset()
	builder, err := benchInsertBuilder(rows, argc)
	if err != nil {
		b.Fatal(err)
	}

	sql, args := builder.ToSQL()

	tx, err := testDB.DB.Beginx()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tx.Exec(sql, args...)
		if err != nil {
			b.Error(err.Error())
		}
	}
}
