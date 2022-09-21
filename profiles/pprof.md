Подготовка профилей для сравнения

```shell
cd internal/app/repository/db/file/
go test -bench=. -benchmem -memprofile=../../../../../profiles/base.pprof

cd internal/app/repository/db/inmemory/
go test -bench=. -benchmem -memprofile=../../../../../profiles/result.pprof
```

Сравниваем
```shell
go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof
```

Результат
```shell
Type: alloc_space
Time: Sep 15, 2022 at 8:14pm (MSK)
Showing nodes accounting for -14667.61MB, 90.29% of 16245.66MB total
Dropped 46 nodes (cum <= 81.23MB)
      flat  flat%   sum%        cum   cum%
-10141.36MB 62.43% 62.43% -10141.36MB 62.43%  encoding/json.(*decodeState).literalStore
-5441.22MB 33.49% 95.92% -15621.13MB 96.16%  github.com/yury-nazarov/shorturl/internal/app/repository/db/file.(*consumer).read
  380.02MB  2.34% 93.58%   375.51MB  2.31%  fmt.Sprintf
  226.95MB  1.40% 92.18%   226.95MB  1.40%  github.com/yury-nazarov/shorturl/internal/app/repository/db/inmemory.(*inMemoryDB).Add
  186.51MB  1.15% 91.03%   234.01MB  1.44%  fmt.Errorf
  -87.50MB  0.54% 91.57%  -597.53MB  3.68%  github.com/yury-nazarov/shorturl/internal/app/repository/db/file.BenchmarkFileDB.func5
      84MB  0.52% 91.06%   596.02MB  3.67%  github.com/yury-nazarov/shorturl/internal/app/repository/db/inmemory.BenchmarkInMemoryDB.func5
   48.50MB   0.3% 90.76%   282.51MB  1.74%  github.com/yury-nazarov/shorturl/internal/app/repository/db/inmemory.(*inMemoryDB).Get
      44MB  0.27% 90.49%   406.45MB  2.50%  github.com/yury-nazarov/shorturl/internal/app/repository/db/inmemory.BenchmarkInMemoryDB.func1
   36.50MB  0.22% 90.26%   564.02MB  3.47%  github.com/yury-nazarov/shorturl/internal/app/repository/db/inmemory.BenchmarkInMemoryDB.func2
   -2.50MB 0.015% 90.28% -10143.86MB 62.44%  encoding/json.(*decodeState).object
   -1.50MB 0.0092% 90.29% -7687.81MB 47.32%  github.com/yury-nazarov/shorturl/internal/app/repository/db/file.BenchmarkFileDB.func3
         0     0% 90.29% -10179.91MB 62.66%  encoding/json.(*Decoder).Decode
         0     0% 90.29% -10144.36MB 62.44%  encoding/json.(*decodeState).unmarshal
         0     0% 90.29% -10143.86MB 62.44%  encoding/json.(*decodeState).value
         0     0% 90.29% -7822.32MB 48.15%  github.com/yury-nazarov/shorturl/internal/app/repository/db/file.(*fileDB).Get
         0     0% 90.29% -7685.31MB 47.31%  github.com/yury-nazarov/shorturl/internal/app/repository/db/file.(*fileDB).GetToken
         0     0% 90.29%  -121.50MB  0.75%  github.com/yury-nazarov/shorturl/internal/app/repository/db/file.(*fileDB).GetUserURL
         0     0% 90.29% -7825.82MB 48.17%  github.com/yury-nazarov/shorturl/internal/app/repository/db/file.BenchmarkFileDB.func2
         0     0% 90.29%  -121.50MB  0.75%  github.com/yury-nazarov/shorturl/internal/app/repository/db/file.BenchmarkFileDB.func4
         0     0% 90.29% -14671.66MB 90.31%  testing.(*B).launch
         0     0% 90.29% -14674.66MB 90.33%  testing.(*B).runN

```