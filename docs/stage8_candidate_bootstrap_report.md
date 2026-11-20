# Этап 8 — bootstrap кандидатов

## Основа и план

HEAD: 30aadbccedd56b5a0bb769e59403944b22fda75d, ветка wip/geometry-editor-v2, на входе чистое дерево. Отчёты 5–7 и резервная копия проверены; baseline HEAD отдельно заархивирован вне /tmp. План: fail-before с конструктивным witness → bounded fallback → unit/server/HTTP/PostGIS → before/after малых fixtures → полный свежий gate. К этапу 9 нельзя переходить до его завершения.

## Причина / реализация

collectAnchors берёт вершины границ, candidatesFromAnchors совмещает каждую вершину повёрнутого лекала с каждым anchor. У радиального 12-вершинного вогнутого лекала нет вершины в bbox-corner: при контакте с углом листа другая часть детали выходит наружу. В пустом листе NFP от occupied отсутствует. Конструктивный witness Transform(pattern,50,50,0) проходит ValidatePolygon/CoveredBy/InteriorsIntersect; причина — неполнота кандидатов, не невалидность входа.

Добавлен общий chooseBootstrap: только после неуспеха прежнего chooseBestCandidate, не вместо него. На каждый разрешённый угол до 5 seeds:4 совмещения границ bbox и центр bbox листа. Для Transform(p,0,0,r) вычисляется rotated bbox: x=minSurfaceX−minRotatedX и т.п.; позиция — координата исходного bbox rotation center, а не обязательно центр повернутого bbox. Это учитывает асимметричные полигоны/непрямые углы. Детерминированный порядок, exact Candidate dedup, finite checks, ctx до/между вычислениями и после selection. Бюджет fallback <=200000−legacyCandidateCount; при превышении ErrComplexity. При лимите 36angles максимум 180seeds до dedup. Стандартные Transform/Bounds/CoveredBy/InteriorsIntersect/score и серверный validateResult сохранены, epsilon/случайность/objective не менялись. Legacy feasible candidate имеет полный приоритет.

Файлы: новый internal/nesting/bootstrap.go; минимальные вызовы из place_one_pattern.go/place_one_pattern_nfp.go; bootstrap_test.go; расширение usecase geometry_fixture_validation_test.go; integration/bootstrap_preview_test.go; обоснованное обновление nfp_exact_determinism_test.go; этот отчёт/сводка. Все 5 алгоритмов используют общие вызовы, без копирования логики в 5 оптимизаторов. Схема/API контракты не менялись.

## Fail-before / pass-after и найденные отказы

TestRadialBootstrapRegression на неизменённом production HEAD:4FAIL (Baseline/NFP×[0]/[0,90]), placed0 при валидном witness. После:4PASS,placed1, фактическая transformed geometry проходит engine validation. Все 5 алгоритмов с единственным разрешённым 90 размещают 1 без добавления 0. Source fixture тот же experimentPolygon(12), поверхность 100×100, no obstacles.

Первый full Go/race выявил устаревший **строгий Hybrid golden**: NFP calls8258→8205,repeated7257→6923,mismatches0. Проверены два диагностических повтора и 3 финальных: одинаковые результаты, новое качество **9 вместо 8** placements,unplaced2 вместо 3,usedArea4800/height60 без ухудшения; все placements математически валидны. Эталонные counts и качество обновлены на эти значения, assertions не удалены. Изменение вызвано доступностью новых допустимых позиций в декодере, production RNG не менялся. Исходные failed логи сохранены.

Первый SQL oracle по собственной формуле ST_Rotate/ST_Centroid/ST_Translate не совпал с порядком float операций движка на точном boundary contact. Проверка исправлена: PostGIS независимо выполняет ST_IsValid/ST_CoveredBy на **точном WKB возвращённой геометрии**, без другого преобразования. Engine также реконструирует геометрию из XY/rotation при validateResult. Не вводился epsilon и не отключалась проверка границ.

Первый race также достиг локального 30s deadline в двух production-config fixtures (holes/Hybrid и radial/Hybrid) при одновременном полном native regression. Это test budget, а не изменение production120s. Новый radial server fixture ограничен qty1 (остальные существующие qty2), в соответствии с задачей первого размещения. Финальные результаты повторного gate приведены ниже; отказы не считаются PASS.

## Регрессии / окружение

bootstrap_test: radial exact regression; все 5 семейств; Baseline/NFP rectangle6,L,1/2holes,3obstacles,полностью занятый лист / слишком большая деталь→0,diamond,±100000,асимметричныйтреугольник 33°; разрешённые углы/metrics/unplaced/utilization, касание границ и legacypriority; отмена до/generation/selection без partial, остаточный candidate budget.

Server validation:4 старых fixtures×5productionconfig qty2 плюс radial×5productionconfig qty1; намеренно испорченный X=1e8 отклоняется. Integration: настоящий PostgreSQL/PostGIS private schema, HTTP handler POST preview/save с UUID, exact preview_id и geometry, Optimize count1, повтор save не создаёт историю, stale после concurrentUPDATE не перезаписывает layout, отмена не создаёт partial. Middleware owner в этом адресном handler-test задан fixture; настоящие JWT/TCP/PID проверяет полный TestRealProcessLifecycle и этап 10.

SQL gate:current_database=nesting_test,current_user=postgres,PostGIS3.4. PostGIS заранее public, private schema/table/migration failclosed guard. Working PostgreSQL/volumes/.env не затрагивались. Go1.26.2,Node24.22.2,golangci2.13.2. Тяжёлые 30k operations не повторялись. Для SQL static binary CGO_ENABLED=0, race CGO binary через host-loader/libc из совместимой тестовой среды. Служебный child helper и explicit heavy experiments SKIP в обычном parent ожидаемы; отдельные фактические прогоны обозначаются отдельно.

## Границы / ручная проверка

Это ограниченная эвристика, не полный continuous packing search. Внутренняя узкая свободная область сложного листа может не совпасть с 5seeds; произвольный 500/36 за 120s не обещается. Исходного пользовательского SVG нет; earcut/UnionMany остаются неделимыми. Владелец позже импортирует radial SVG/свой оригинал, проверит положительный preview, exactsave/reload и занятый лист. UX и внешний вид не менялись.

## Свежий gate / замеры

Фактически выполненные проверки и замеры приведены ниже.


### Before/after малых fixtures (3 независимые пары)

CPU2,hardRAM1GiB,swapoff,GOMAXPROCS2,GOMEMLIMIT512MiB,networknone,readonlyroot,tmpfs exec,process timeout90s. Измерения последовательно после завершения тяжёлых regressions; сторонние процессы хоста не контролировались. До — отдельный git archive HEAD: 30aadbc; после — копия текущих tracked backend sources + bootstrap.go. Один и тот же measurement test,seed/config/fixtures.36rows в bootstrap-measurements.csv. Время различающихся radial результатов не считается speedup.

|Алгоритм/fixture|Before wall ms median[min,max]|After wall ms median[min,max]|Placed before→after|
|---|---:|---:|---|
|baseline/concave|32.155[27.673,34.059]|27.207[26.443,40.056]|6→6|
|baseline/radial_without_seed|1.088[0.998,1.108]|1.136[1.117,1.294]|0→1|
|baseline/rectangle|8.780[8.054,10.329]|8.847[8.216,9.666]|6→6|
|nfp_greedy/concave|22.978[21.474,26.566]|23.218[22.071,26.201]|6→6|
|nfp_greedy/radial_without_seed|1.198[1.160,1.208]|1.229[1.224,1.288]|0→1|
|nfp_greedy/rectangle|5.008[4.711,7.213]|4.909[4.881,5.568]|6→6|

Существующие rectangle/L fixtures размещают 6 какдо, радиальный 1 вместо 0. Все 36completed, post-transform validation и метрики проверены. Простой путь не вызывает fallback при legacy success. Короткие timings подвержены scheduler noise: не объявляем ускорение/жёсткий SLA по 3 повторам.

### Gate этапа 8 — PASS

- go test -count=1 ./...,go vet ./...,golangci-lint run ./...: свежие PASS,0issues (stage8-go-final.log). Обычный полный прогон включал более тяжёлый radial server qty2 до уменьшения только этого fixture; финальный qty1 проверен новым race gate. Production-код после этого не менялся.
- go test -race -count=1 -timeout=20m ./internal/nesting ./internal/geometry/nfp ./internal/usecase/nesting ./internal/transport/http/handler: PASS; native325.886s,nfp1.361s,usecase119.179s,handler1.019s. Старые 30s test-budget failures выше не скрыты, final holes/Hybrid с исходным qty2 также PASS.
- Полный актуальный PostGIS: PASS (stage8-integration-v3.log), новый HTTP+SQL radial3 повтора PASS. Первое ожидание общего conflict code в новом stale assertion исправлено на точный nesting_stale; это test assertion, не productionfix.
- PostGIS race: RadialPreviewSavedExactly,RealProcessLifecycle,NativeDeadlinePreservesSavedLayout PASS с актуальным binary,parent/children race-enabled. NESTING_TEST_DSN явно проверен; секреты не печатаются.
- npm run build/lint и node --test --test-isolation=none frontend/tests/*.test.mjs: PASS,72/72. Frontend не менялся; новый браузерный прогон запланирован этапом 10, здесь SKIP.
- Три fixed Hybrid golden повтора PASS;4radial fail-before/4pass-after; математические и negative cases PASS.
- git diff --check PASS; дополнительные untracked whitespace checks выполнены. Production DB/container не изменены, commit/push/merge не выполнялись.

Команды воспроизведения: из backend go test -count=1 -v ./internal/nesting -run 'Test(RadialBootstrap|Bootstrap|NFPExactInputDeterminismHybrid)'; полный Go/race выше. Из backend/internal/integration с валидированным DSN: ./stage8-v3.test -test.count=1 -test.v -test.timeout=300s. Для SQL race совместимый loader и NESTING_TEST_PROCESS_BINARY/NESTING_TEST_CHILD_LOADER/NESTING_TEST_CHILD_LIBRARY как в отчёте 5. Малые measurement binaries в network-none capped container, -test.run '^TestStage8Measurement$' -test.count=1 -test.v -test.timeout=90s.

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
?? docs/stage8_candidate_bootstrap_report.md
```
Обычный stat не включает untracked новые файлы выше.
