    N         int           // The number of iterations.
    T         time.Duration // The total time taken.
    Bytes     int64         // Bytes processed in one iteration.
    MemAllocs uint64        // The total number of memory allocations.

› godo bench

```
BenchmarkTransactedDat2    10000        111709 ns/op         624 B/op         18 allocs/op
BenchmarkTransactedSql2    10000        173870 ns/op         881 B/op         30 allocs/op
BenchmarkTransactedDat4    10000        115597 ns/op         800 B/op         21 allocs/op
BenchmarkTransactedSql4    10000        184454 ns/op         978 B/op         35 allocs/op
BenchmarkTransactedDat8    10000        146118 ns/op         904 B/op         27 allocs/op
BenchmarkTransactedSql8     5000        221265 ns/op        1194 B/op         44 allocs/op
BenchmarkVaryingLengthDat128Binary      5000        305182 ns/op        2194 B/op         41 allocs/op
BenchmarkVaryingLengthSql128Binary      5000        300642 ns/op        1884 B/op         35 allocs/op
BenchmarkVaryingLengthDat512Binary      5000        329345 ns/op        7984 B/op         47 allocs/op
BenchmarkVaryingLengthSql512Binary      5000        325058 ns/op        7671 B/op         41 allocs/op
BenchmarkVaryingLengthDat4096Binary     3000        539341 ns/op       70626 B/op         55 allocs/op
BenchmarkVaryingLengthSql4096Binary     3000        532109 ns/op       70314 B/op         49 allocs/op
BenchmarkVaryingLengthDat8192Binary     2000        792137 ns/op      132132 B/op         58 allocs/op
BenchmarkVaryingLengthSql8192Binary     2000        817895 ns/op      131820 B/op         52 allocs/op
BenchmarkVaryingLengthDat128Text       10000        212843 ns/op        1072 B/op         16 allocs/op
BenchmarkVaryingLengthSql128Text        5000        294508 ns/op         896 B/op         27 allocs/op
BenchmarkVaryingLengthDat512Text        5000        230352 ns/op        3282 B/op         17 allocs/op
BenchmarkVaryingLengthSql512Text        5000        301830 ns/op        2304 B/op         28 allocs/op
BenchmarkVaryingLengthDat4096Text       3000        362475 ns/op       18904 B/op         17 allocs/op
BenchmarkVaryingLengthSql4096Text       3000        371247 ns/op        9474 B/op         28 allocs/op
BenchmarkVaryingLengthDat8192Text       2000        564513 ns/op       34270 B/op         17 allocs/op
BenchmarkVaryingLengthSql8192Text       3000        454148 ns/op       17412 B/op         28 allocs/op
BenchmarkBuildExecSQLDat2       5000        216065 ns/op         624 B/op         18 allocs/op
BenchmarkBuildExecSQLSql2       5000        298815 ns/op         881 B/op         30 allocs/op
BenchmarkBuildExecSQLDat4       5000        221711 ns/op         800 B/op         21 allocs/op
BenchmarkBuildExecSQLSql4       5000        306066 ns/op         978 B/op         35 allocs/op
BenchmarkBuildExecSQLDat8       5000        252755 ns/op         904 B/op         27 allocs/op
BenchmarkBuildExecSQLSql8       5000        346723 ns/op        1194 B/op         44 allocs/op
BenchmarkExecSQLDat2        5000        216172 ns/op         624 B/op         18 allocs/op
BenchmarkExecSQLSql2        5000        296267 ns/op         881 B/op         30 allocs/op
BenchmarkExecSQLDat4        5000        222349 ns/op         800 B/op         21 allocs/op
BenchmarkExecSQLSql4        5000        306460 ns/op         978 B/op         35 allocs/op
```
ok      github.com/mgutz/dat/sqlx-runner    48.405s


BenchmarkTransactedDat2            10000                             114017   ns/op  832     B/op  21  allocs/op
BenchmarkTransactedSql2            10000                             173774   ns/op  881     B/op  30  allocs/op
BenchmarkTransactedSqx2            10000                             178306   ns/op  881     B/op  30  allocs/op
BenchmarkTransactedDat4            10000                             118270   ns/op  1232    B/op  26  allocs/op
BenchmarkTransactedSql4            10000                             186538   ns/op  978     B/op  35  allocs/op
BenchmarkTransactedSqx4            10000                             182337   ns/op  978     B/op  35  allocs/op
BenchmarkTransactedDat8            10000                             152272   ns/op  1480    B/op  33  allocs/op
BenchmarkTransactedSql8            5000                              222522   ns/op  1194    B/op  44  allocs/op
BenchmarkTransactedSqx8            5000                              221790   ns/op  1194    B/op  44  allocs/op
BenchmarkVaryingLengthDat64        5000                              311759   ns/op  2213    B/op  45  allocs/op
BenchmarkVaryingLengthSql64        5000                              301019   ns/op  1435    B/op  35  allocs/op
BenchmarkVaryingLengthSqx64        5000                              301003   ns/op  1435    B/op  35  allocs/op
BenchmarkVaryingLengthDat512       3000                              360636   ns/op  10915   B/op  53  allocs/op
BenchmarkVaryingLengthSql512       5000                              337574   ns/op  10266   B/op  43  allocs/op
BenchmarkVaryingLengthSqx512       5000                              336681   ns/op  10266   B/op  43  allocs/op
BenchmarkVaryingLengthDat2048      2000                              517949   ns/op  43585   B/op  58  allocs/op
BenchmarkVaryingLengthSql2048      3000                              466612   ns/op  44217   B/op  48  allocs/op
BenchmarkVaryingLengthSqx2048      3000                              469630   ns/op  44217   B/op  48  allocs/op
BenchmarkVaryingLengthDat8192      2000                              1111741  ns/op  166084  B/op  65  allocs/op
BenchmarkVaryingLengthSql8192      2000                              974755   ns/op  168468  B/op  54  allocs/op
BenchmarkVaryingLengthSqx8192      2000                              977573   ns/op  168468  B/op  54  allocs/op
BenchmarkVaryingLengthDat128Text   10000                             213237   ns/op  1232    B/op  18  allocs/op
BenchmarkVaryingLengthSql128Text   5000                              292275   ns/op  896     B/op  27  allocs/op
BenchmarkVaryingLengthSqx128Text   5000                              292246   ns/op  896     B/op  27  allocs/op
BenchmarkVaryingLengthDat512Text   5000                              231574   ns/op  3442    B/op  19  allocs/op
BenchmarkVaryingLengthSql512Text   5000                              301583   ns/op  2304    B/op  28  allocs/op
BenchmarkVaryingLengthSqx512Text   5000                              301949   ns/op  2304    B/op  28  allocs/op
BenchmarkVaryingLengthDat4096Text  3000                              365685   ns/op  19064   B/op  19  allocs/op
BenchmarkVaryingLengthSql4096Text  3000                              370244   ns/op  9474    B/op  28  allocs/op
BenchmarkVaryingLengthSqx4096Text  3000                              370621   ns/op  9474    B/op  28  allocs/op
BenchmarkVaryingLengthDat8192Text  2000                              560822   ns/op  34430   B/op  19  allocs/op
BenchmarkVaryingLengthSql8192Text  3000                              451457   ns/op  17412   B/op  28  allocs/op
BenchmarkVaryingLengthSqx8192Text  3000                              451568   ns/op  17412   B/op  28  allocs/op
BenchmarkBuildExecSQLDat2          5000                              220634   ns/op  832     B/op  21  allocs/op
BenchmarkBuildExecSQLSql2          5000                              297181   ns/op  881     B/op  30  allocs/op
BenchmarkBuildExecSQLSqx2          5000                              297442   ns/op  881     B/op  30  allocs/op
BenchmarkBuildExecSQLDat4          5000                              224502   ns/op  1232    B/op  26  allocs/op
BenchmarkBuildExecSQLSql4          5000                              305092   ns/op  978     B/op  35  allocs/op
BenchmarkBuildExecSQLSqx4          5000                              305586   ns/op  978     B/op  35  allocs/op
BenchmarkBuildExecSQLDat8          5000                              257312   ns/op  1480    B/op  33  allocs/op
BenchmarkBuildExecSQLSql8          5000                              344629   ns/op  1194    B/op  44  allocs/op
BenchmarkBuildExecSQLSqx8          5000                              345671   ns/op  1194    B/op  44  allocs/op
BenchmarkExecSQLDat2               5000                              208888   ns/op  280     B/op  10  allocs/op
BenchmarkExecSQLSql2               5000                              298915   ns/op  881     B/op  30  allocs/op
BenchmarkExecSQLSqx2               5000                              296452   ns/op  881     B/op  30  allocs/op
BenchmarkExecSQLDat4               5000                              223017   ns/op  296     B/op  10  allocs/op
BenchmarkExecSQLSql4               5000                              303522   ns/op  978     B/op  35  allocs/op
BenchmarkExecSQLSqx4               5000                              304502   ns/op  978     B/op  35  allocs/op
ok                                 github.com/mgutz/dat/sqlx-runner  72.733s
bench 86233ms


PASS
BenchmarkDeleteSql   1000000          1314 ns/op
BenchmarkInsertValuesSql     1000000          1926 ns/op
BenchmarkInsertRecordsSql     300000          5720 ns/op
BenchmarkSelectBasicSql   500000          2882 ns/op
BenchmarkSelectFullSql    200000          6343 ns/op
BenchmarkUpdateValuesSql     1000000          1388 ns/op
BenchmarkUpdateValueMapSql    500000          2758 ns/op
ok      github.com/mgutz/dbr    10.892s
bench 11625ms

~/go/src/github.com/mgutz/dbr postgres-only ∆
› godo bench
PASS
BenchmarkDeleteSql   1000000          1305 ns/op
BenchmarkInsertValuesSql     1000000          1881 ns/op
BenchmarkInsertRecordsSql     300000          5706 ns/op
BenchmarkSelectBasicSql   500000          2828 ns/op
BenchmarkSelectFullSql    200000          6278 ns/op
BenchmarkUpdateValuesSql     1000000          1385 ns/op
BenchmarkUpdateValueMapSql    500000          2754 ns/op
ok      github.com/mgutz/dbr    10.774s
bench 11517ms

~/go/src/github.com/mgutz/dbr postgres-only ∆
› godo bench
PASS
BenchmarkDeleteSql   1000000          1323 ns/op
BenchmarkInsertValuesSql     1000000          1923 ns/op
BenchmarkInsertRecordsSql     300000          5709 ns/op
BenchmarkSelectBasicSql   500000          2879 ns/op
BenchmarkSelectFullSql    200000          6401 ns/op
BenchmarkUpdateValuesSql     1000000          1365 ns/op
BenchmarkUpdateValueMapSql    500000          2727 ns/op
ok      github.com/mgutz/dbr    10.848s
bench 11613ms

~/go/src/github.com/mgutz/dbr postgres-only ∆
› godo bench
PASS
BenchmarkDeleteSql   1000000          1324 ns/op
BenchmarkInsertValuesSql     1000000          1920 ns/op
BenchmarkInsertRecordsSql     300000          5725 ns/op
BenchmarkSelectBasicSql   500000          2883 ns/op
BenchmarkSelectFullSql    200000          6396 ns/op
BenchmarkUpdateValuesSql     1000000          1363 ns/op
BenchmarkUpdateValueMapSql    500000          2777 ns/op
ok      github.com/mgutz/dbr    10.884s
bench 11617ms


