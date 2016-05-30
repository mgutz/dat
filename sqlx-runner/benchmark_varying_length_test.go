package runner

import (
	"bytes"
	"testing"

	"github.com/syreclabs/dat"
)

// These benchmarks compare the total cost of interpolating the SQL then
// executing the query on the same connection using a transaction.
// Both database/sql and jmoiron/sqlx can take advantage of a prepared
// statements.

func BenchmarkVaryingLengthDatBinary128(b *testing.B) {
	benchmarkVaryingLengthDatN(b, 128, 0)
}

func BenchmarkVaryingLengthSqlBinary128(b *testing.B) {
	benchmarkVaryingLengthSqlN(b, 128, 0)
}

func BenchmarkVaryingLengthDatBinary512(b *testing.B) {
	benchmarkVaryingLengthDatN(b, 512, 0)
}

func BenchmarkVaryingLengthSqlBinary512(b *testing.B) {
	benchmarkVaryingLengthSqlN(b, 512, 0)
}

func BenchmarkVaryingLengthDatBinary4K(b *testing.B) {
	benchmarkVaryingLengthDatN(b, 4096, 0)
}

func BenchmarkVaryingLengthSqlBinary4K(b *testing.B) {
	benchmarkVaryingLengthSqlN(b, 4096, 0)
}

func BenchmarkVaryingLengthDatBinary8K(b *testing.B) {
	benchmarkVaryingLengthDatN(b, 8*1024, 0)
}

func BenchmarkVaryingLengthSqlBinary8K(b *testing.B) {
	benchmarkVaryingLengthSqlN(b, 8*1024, 0)
}

func BenchmarkVaryingLengthDatText128(b *testing.B) {
	benchmarkVaryingLengthDatN(b, 0, 128)
}

func BenchmarkVaryingLengthSqlText128(b *testing.B) {
	benchmarkVaryingLengthSqlN(b, 0, 128)
}

func BenchmarkVaryingLengthDatText512(b *testing.B) {
	benchmarkVaryingLengthDatN(b, 0, 512)
}

func BenchmarkVaryingLengthSqlText512(b *testing.B) {
	benchmarkVaryingLengthSqlN(b, 0, 512)
}

func BenchmarkVaryingLengthDatText4K(b *testing.B) {
	benchmarkVaryingLengthDatN(b, 0, 4096)
}

func BenchmarkVaryingLengthSqlText4K(b *testing.B) {
	benchmarkVaryingLengthSqlN(b, 0, 4096)
}

func BenchmarkVaryingLengthDatText8K(b *testing.B) {
	benchmarkVaryingLengthDatN(b, 0, 8*1024)
}

func BenchmarkVaryingLengthSqlText8K(b *testing.B) {
	benchmarkVaryingLengthSqlN(b, 0, 8*1024)
}

func BenchmarkVaryingLengthDatText64K(b *testing.B) {
	benchmarkVaryingLengthDatN(b, 0, 64*1024)
}

func BenchmarkVaryingLengthSqlText64K(b *testing.B) {
	benchmarkVaryingLengthSqlN(b, 0, 64*1024)
}

func benchmarkVaryingLengthDatN(b *testing.B, maxBytes int, maxText int) {
	benchReset()
	builder, err := benchInsertVaryingLengthBuilder(maxBytes, maxText)
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
			b.Error(err.Error())
		}
	}
}

func benchmarkVaryingLengthSqlN(b *testing.B, maxBytes int, maxText int) {
	benchReset()
	builder, err := benchInsertVaryingLengthBuilder(maxBytes, maxText)
	if err != nil {
		b.Fatal(err)
	}
	sql, args := builder.ToSQL()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := testDB.Exec(sql, args...)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkVaryingLengthSqlxN(b *testing.B, maxBytes int, maxText int) {
	benchReset()
	builder, err := benchInsertVaryingLengthBuilder(maxBytes, maxText)
	if err != nil {
		b.Fatal(err)
	}
	sql, args := builder.ToSQL()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := testDB.DB.Exec(sql, args...)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// benchInsertVaryingLengthBuilder builds an INSERT statement with varying
// length data.
func benchInsertVaryingLengthBuilder(maxBytes, maxText int) (*dat.InsertBuilder, error) {
	var columns []string
	var values []interface{}

	if maxBytes > 0 {
		image := make([]byte, maxBytes)
		for i := 0; i < maxBytes; i++ {
			image[i] = byte(i % 256)
		}
		columns = append(columns, "image")
		values = append(values, image)
	}

	if maxText > 0 {
		var buf bytes.Buffer
		for i := 0; i < maxText; i++ {
			if i > 0 && i%1024 == 0 {
				// force escaping
				buf.WriteRune('\'')
			} else {
				buf.WriteRune('t')
			}
		}
		columns = append(columns, "name")
		values = append(values, buf.String())
	}

	builder := dat.
		NewInsertBuilder("benches").
		Columns(columns...).
		Values(values...)

	return builder, nil
}
