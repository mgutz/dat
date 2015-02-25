PASS
BenchmarkDeleteSql          1000000               1145     ns/op  440   B/op  11  allocs/op
BenchmarkInsertValuesSql    1000000               1893     ns/op  744   B/op  13  allocs/op
BenchmarkInsertRecordsSql   500000                3383     ns/op  888   B/op  21  allocs/op
BenchmarkInterpolate        300000                5818     ns/op  904   B/op  13  allocs/op
BenchmarkSelectBasicSql     500000                2739     ns/op  960   B/op  17  allocs/op
BenchmarkSelectFullSql      200000                6950     ns/op  2136  B/op  40  allocs/op
BenchmarkUpdateValuesSql    1000000               1409     ns/op  544   B/op  14  allocs/op
BenchmarkUpdateValueMapSql  500000                3212     ns/op  1248  B/op  24  allocs/op
ok                          github.com/mgutz/dat  12.534s
PASS
BenchmarkTransactedDat2    10000                             111408   ns/op  832   B/op  21  allocs/op
BenchmarkTransactedSql2    10000                             173383   ns/op  881   B/op  30  allocs/op
BenchmarkTransactedSqx2    10000                             175494   ns/op  881   B/op  30  allocs/op
BenchmarkTransactedDat4    10000                             115811   ns/op  1232  B/op  26  allocs/op
BenchmarkTransactedSql4    10000                             182645   ns/op  978   B/op  35  allocs/op
BenchmarkTransactedSqx4    10000                             182248   ns/op  978   B/op  35  allocs/op
BenchmarkTransactedDat8    10000                             145904   ns/op  1480  B/op  33  allocs/op
BenchmarkTransactedSql8    5000                              223762   ns/op  1194  B/op  44  allocs/op
BenchmarkTransactedSqx8    10000                             222488   ns/op  1194  B/op  44  allocs/op
BenchmarkBuildExecSQLDat2  5000                              214768   ns/op  832   B/op  21  allocs/op
BenchmarkBuildExecSQLSql2  5000                              302398   ns/op  881   B/op  30  allocs/op
BenchmarkBuildExecSQLSqx2  5000                              302595   ns/op  881   B/op  30  allocs/op
BenchmarkBuildExecSQLDat4  5000                              226969   ns/op  1232  B/op  26  allocs/op
BenchmarkBuildExecSQLSql4  5000                              312168   ns/op  978   B/op  35  allocs/op
BenchmarkBuildExecSQLSqx4  5000                              314042   ns/op  978   B/op  35  allocs/op
BenchmarkBuildExecSQLDat8  5000                              257797   ns/op  1480  B/op  33  allocs/op
BenchmarkBuildExecSQLSql8  5000                              351399   ns/op  1194  B/op  44  allocs/op
BenchmarkBuildExecSQLSqx8  5000                              348669   ns/op  1194  B/op  44  allocs/op
BenchmarkExecSQLDat2       10000                             210165   ns/op  280   B/op  10  allocs/op
BenchmarkExecSQLSql2       5000                              300423   ns/op  881   B/op  30  allocs/op
BenchmarkExecSQLSqx2       5000                              298474   ns/op  881   B/op  30  allocs/op
BenchmarkExecSQLDat4       5000                              214584   ns/op  296   B/op  10  allocs/op
BenchmarkExecSQLSql4       5000                              308715   ns/op  978   B/op  35  allocs/op
BenchmarkExecSQLSqx4       5000                              313335   ns/op  978   B/op  35  allocs/op
ok                         github.com/mgutz/dat/sqlx-runner  37.760s
bench 51719ms
