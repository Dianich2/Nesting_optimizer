# Этап 9 — целевая производительность

## Основа / изоляция

HEAD: 30aadbccedd56b5a0bb769e59403944b22fda75d, wip/geometry-editor-v2. На входе только 8 изменённых/новых файлов принятого этапа 8 (status в отчёте 8). Они сохранены. Сравнение идёт **с этапом 8**, а не с одним git archive HEAD: stage8-accepted-source.tar.gz содержит tracked sources + незакоммиченные bootstrap/исправленные тесты, SHA-256 рядом; он извлечён в /tmp/nesting-stage9-before. Никакого reset/clean/перезаписи чужих изменений.

План: новый профиль → один локальный эксперимент → differential correctness → pilot →3 пары одинаковых process runs → принять/убрать только свой патч → свежие regressions/SQL/race/frontend. Production RNG, objective, precision, geometry engine/library, cache limits, idempotency/snapshot/SQL lifecycle не менялись.

Среда: Go1.26.2 linux/amd64, network-none container nesting-stages8-10-measure создан текущей задачей; cpu2/GOMAXPROCS2, hardRAM1GiB/memory-swap1GiB (без swap), GOMEMLIMIT512MiB softtarget, pids128, readonlyroot, executable tmpfs work128MiB/tmp64MiB. Жизнь контейнера ограничена 7200s, каждый process300s, fixture40s; profile90s. В контейнере нет PostgreSQL. Paired/profile runs последовательны, без собственных других тяжёлых проверок; чужие процессы хоста не контролировались. Это описательные замеры, не bare-metal/SLA/95%CI.

## Профиль до патча

Новый настоящий Baseline occupied radial50: wall19.514s,CPU21.274s,TotalAlloc12608842464bytes,179629412allocations,process cumulativeRSS24600KiB. Все 50 размещены/валидны. CPU cumulative Bounds4.28%,ValidatePolygon28.16%,LineString.IsSimple21.02%; alloc_space cumulative Bounds7.81%,toSimpleFeaturesPolygon16.34%,LineString.IsSimple30.37%. Cumulative проценты перекрываются; складывать их нельзя. Это профиль post-stage8 кода, старые 18.6s из этапа 6 не выдаются за базу.

## Исследованный performance-патч — СНЯТ

Единственный исследованный production performance-патч: Engine.Bounds сохраняет **всю ValidatePolygon**, но после неё не создаёт второй simplefeatures.Polygon со всеми ring copies ради Envelope. В текущей dependency v0.59.0 Polygon.Envelope использует exterior ring; прямой проход по тем же точкам даёт те же extrema. Порядок tie comparisons и closing point сохранён, включая signed zero; epsilon/округление не вводились. Невалидный/пустой/NaN polygon всё так же отклоняется до доступа к points[0]. Объекты не мутируются, cache/state/aliasing не добавлены. CoveredBy,InteriorsIntersect,Transform и server validateResult не обходятся и не изменены. Это не замена predicate bbox-проверкой.

bounds_equivalence_test.go:100 детерминированных convex/concave contours с 4–32vertices,holes,reverse/closed rings,offsets/rotations + signedzero; сравнение bitwise с прежним validated toSimpleFeaturesPolygon.Envelope().MinMaxXYs(). Negative empty/collinear/self-crossing/NaN отклоняются. Полный engine suite PASS. Цена валидации остаётся основной; большого ускорения от этой небольшой конверсии не ожидаем.

Оставленные файлы этапа: internal/nesting/performance_experiment_test.go (explicit gated10cases); scripts/summarize_targeted_performance.py (строгая полнота/quality/work CSV); этот отчёт/сводка. Экспериментальные engine.go diff и bounds_equivalence_test.go перенесены в stage9-rejected artifacts; исходный engine.go восстановлен только из принятого stage8 snapshot, свой экспериментальный test удалён из рабочего дерева. Иначе он после снятия патча зеркалил бы прежнюю реализацию. Никаких вторых speculative optimizations не добавлено.

## Набор / метод

Baseline и NFP Greedy:rectangle6,concave L10 с 3 углами,holes10 с 2 отверстиями,occupied radial50,occupied rectangle100. Набор включает 6/10/50/100 и занятые области; это контролируемые fixtures, не arbitrary500stress. Production-config все 5 семейств дополнительно проверяются server validation suite. Одинаковые fixture/seed/harness для before/after; Baseline/NFP детерминированны, production stochastic generators не трогаем.

Pilot before28.90s/after28.94s:10completed каждый, все запрошенные детали реально размещены и валидны. Pilot отдельно от основных 3 пар, не смешивается в статистике. Основные 10cases×6freshprocesses дают 60 строк. wall,CPU,TotalAlloc/Mallocs delta,2ms sampled goroutines,process cumulativeMaxRSS,NFPcalls/cachehits/misses,placed/unplaced/utilization/valid записаны тем же runGeometryExperiment helper. Validation выполняется после timing замера. Непрерывный RSS пооперациям не измеряется; MaxRSS кумулятивен на процесс, GOMEMLIMIT не hardlimit. Нет speedup claim по 3 шумнымповторам и перекрывающимсядиапазонам.

Summarizer требует 10 различных cases/полный PASS на run,все completed/valid/placed=quantity,неизменные placed/unplaced/utilization/NFPcalls/hits/misses во всех 6 процессах. Ошибка/timeout не считается размещением; failed/incomplete log отвергается. Для внешней пересборки baseline snapshot включает незакоммиченныйэтап 8; measurement_test добавлен одинаково кобеимверсиям.

## Команды

```sh
# В backend соответствующей изолированной копии:
CGO_ENABLED=0 go test -c -o /tmp/stage9-before.test ./internal/nesting
# В network-none capped container, последовательно before1/after1/.../after3:
NESTING_TARGETED_PERFORMANCE=1 /work/stage9-before.test \
 -test.run '^TestTargetedPerformanceExperiment$' -test.count=1 -test.v -test.timeout=300s
# Before/after profile отдельными процессами:
NESTING_PROFILE_GEOMETRY=1 /work/stage9-before.test \
 -test.run '^TestProfileOccupiedConcaveGeometry$' -test.count=1 -test.v -test.timeout=90s \
 -test.cpuprofile=/work/profile.cpu -test.memprofile=/work/profile.mem
python3 scripts/summarize_targeted_performance.py --input /tmp --output ARTIFACT_DIR
```

Без секретов/рабочего DSN. Profile загружен docker exec cat (tmpfs не виден docker cp), анализ go tool pprof -top/-alloc_space.

## Границы / ручная проверка

Исходного пользовательского SVG нет; гарантия arbitrary500/36 за 120s не заявляется. Неизменённые Earcut/UnionMany остаются кооперативно отменяемыми между вызовами. rtree/проверки IsSimple требуют отдельной доказанной оптимизации, не удалялись ради скорости. Владелец позже проверяет реальные SVG/качество/ощущение скорости; человеческая UX-приёмка отложена. Новые распределённые workers/retention/миграции не вводились.

## Решение по результатам / свежий gate

Фактические результаты ниже; кэши/старые PASS не заменяют новый запуск.


### Результат эксперимента / решение

Все 60rows PASS:10cases×3 пары, без timeout, placed6/10/50/100, valid, прежние NFPcalls/hits/misses и utilization.

|Fixture/algorithm|Before wall ms median [min,max]|After wall ms median [min,max]|Before/after allocated MiB median|Before/after allocations median|Placed|
|---|---:|---:|---:|---:|---:|
|concave10/baseline|125.699 [122.862,143.141]|133.055 [123.222,138.079]|87.22/85.68|1370939.00/1348404.00|10|
|concave10/nfp_greedy|77.100 [76.771,77.995]|77.826 [76.767,79.551]|48.40/47.68|791563.00/781132.00|10|
|holes10/baseline|78.436 [73.621,81.617]|75.762 [73.609,77.217]|49.78/49.25|958039.00/948178.00|10|
|holes10/nfp_greedy|36.210 [36.138,36.783]|37.733 [34.603,40.750]|23.25/23.00|436810.00/432070.00|10|
|occupied100/baseline|3897.408 [3787.780,3967.907]|3891.848 [3731.484,4074.011]|2284.10/2247.79|44517217.00/43837416.00|100|
|occupied100/nfp_greedy|425.506 [420.676,434.889]|469.501 [431.012,492.469]|258.71/252.88|4712500.00/4603333.00|100|
|occupied50/baseline|18731.598 [17673.634,18776.577]|18967.751 [17960.810,19441.362]|12024.72/11858.24|179629360.00/178174453.00|50|
|occupied50/nfp_greedy|3706.589 [3418.942,3739.525]|3785.344 [3732.338,3835.936]|2277.76/2251.65|36002714.00/35774334.00|50|
|rectangle6/baseline|8.292 [7.819,9.461]|8.197 [7.849,8.522]|5.62/5.48|92878.00/90102.00|6|
|rectangle6/nfp_greedy|4.592 [4.552,7.017]|4.399 [4.261,4.888]|3.13/3.06|51598.00/50268.00|6|

Аллокации уменьшились на всех 10fixtures, но выигрыш wall не подтверждён. NFP occupied100 median425.506→469.501ms (около+10%), диапазоны частично перекрываются: это риск регрессии, а не статистический proof универсального замедления. Baseline occupied50 median18731.598→18967.751ms при меньших bytes12024.72→11858.24MiB. Одного сокращения≈1–2% cumulative allocations недостаточно, чтобы принять рефакторинг с неясным влиянием на отзывчивость. **Патч снят**; stage9 production-файлы идентичны принятому stage8, время release не выдаётся за experimental after. Это безопасное завершение этапа по пункту 3.5 ТЗ, не незавершённый ускоряющий код.

Не переносились глобальные caches/prepared mutable geometry, не пропускались topology predicates/валидация. Следующее потенциальное направление — immutable derived geometry с явным lifetime/aliasing contract и доказательством эквивалентности, отдельным профилем и tests; здесь это не внедрялось.

before отдельный профиль: wall19514.460ms,CPU21273.921ms,allocated12608842464bytes,allocations179629412,cumulativeRSS24600KiB,placed50,valid=True. After — отклонённый эксперимент, не итоговая production версия.

after отдельный профиль: wall19761.298ms,CPU20931.861ms,allocated12434279504bytes,allocations178174526,cumulativeRSS28028KiB,placed50,valid=True. After — отклонённый эксперимент, не итоговая production версия.

Cgroup после всех измерений (bytes; включает tmpfs binaries/profiles, не per-op RSS):
```text
69173248
low 0
high 0
max 0
oom 0
oom_kill 0
oom_group_kill 0
sock_throttled 0
```


### Завершённые свежие проверки

- go test -count=1 ./...,go vet ./...,golangci-lint run ./... — PASS,0issues (stage9-go.log).
- Frontend npm run build/lint и 72unit — PASS,0skipped (stage9-frontend.log).
- Полный актуальный PostGIS suite — PASS (stage9-integration.log), явный read-only DSN gate перед запуском,private schemas,radial HTTP/exactsave/idempotency/cancel/stale и 5 алгоритмов/keep modes.
- PostGIS race parent/children +native timeout+radial preview — PASS (stage9-sql-race.log), актуальные CGO binaries через совместимый loader.
- Отрицательный тест summarizer — PASS: изменённый valid=false вкопии лога отвергнут до outputCSV.
- Byte comparison итогового Engine со snapshot этапа 8 — PASS, свой эксперимент снят полностью.
- Network-none measurement container остановлен только по собственному известному имени; working nesting_postgres и test nesting-reliability-final running.

Explicit TestTargetedPerformanceExperiment SKIP в обычном общем Go запуске, но фактически 60rows+20pilot+2profiles отдельно PASS выше. Полную 30k SQL growth и 500heavygeometry не повторяли: SKIP по целевому объёму этапа 9. Новый Playwright до завершения gate9 не выполнялся; это этап 10. Снятый патч не объявляется ускорением итоговой версии.


### Финальный gate этапа 9 — PASS (negative performance outcome)

Свежий go test -race -count=1 -p=1 -timeout=20m четырёхобязательныхпакетов завершился 0: native323.970s,nfp1.383s,usecase116.980s,handler1.017s. Без race reports. Производственный performance diff отсутствует; геометрический bootstrap этапа 8 сохранён. Все обязательные проверки завершены, отрицательный результат оптимизации оформлен; переход к 10 разрешён безопасным условием 3.5 ТЗ.

```text
 .../internal/nesting/nfp_exact_determinism_test.go   | 12 ++++++++----
 backend/internal/nesting/place_one_pattern.go        |  2 +-
 backend/internal/nesting/place_one_pattern_nfp.go    |  2 +-
 .../nesting/geometry_fixture_validation_test.go      | 20 ++++++++++++++++++--
 4 files changed, 28 insertions(+), 8 deletions(-)
```

```text
 M backend/internal/nesting/nfp_exact_determinism_test.go
 M backend/internal/nesting/place_one_pattern.go
 M backend/internal/nesting/place_one_pattern_nfp.go
 M backend/internal/usecase/nesting/geometry_fixture_validation_test.go
?? backend/internal/integration/bootstrap_preview_test.go
?? backend/internal/nesting/bootstrap.go
?? backend/internal/nesting/bootstrap_test.go
?? backend/internal/nesting/performance_experiment_test.go
?? docs/stage8_candidate_bootstrap_report.md
?? docs/stage9_geometry_performance_report.md
?? scripts/summarize_targeted_performance.py
```
Stat не включает untracked измерители/отчёты. git diff --check PASS; все новые файлы дополнительно проверены на whitespace. Commit/push/merge/рабочаяБД не тронуты.
