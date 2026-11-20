# Этап 12 — BLOCKED: расхождение строгого геометрического oracle

## Основа и фактический объём

HEAD `21eecdbc2a5fa35ddb59028622db3274f8580cb6`, ветка `wip/geometry-editor-v2`. На входе 14 собственных изменённых/новых файлов принятого этапа 11, сохранённых в stage11-accepted-source.tar.gz. Их gate прошёл до начала этого этапа. Чужих изменений не было.

План: детерминированные классы геометрии → метаморфные варианты и seeded-property corpus в bounded process → validateResult → независимый PostGIS oracle точного WKB → регрессии и gate. На oracle найден критический FAIL; по разделам 3.4 и 6 ТЗ продвижение остановлено. **Этап 12 не принят, этап 13 не начат.**

## Файлы

- `backend/internal/nesting/geometry_invariants_test.go`: обычный стабильный core suite и явно включаемый bounded corpus.
- `backend/internal/usecase/nesting/invariant_corpus_test.go`: настоящий server validateResult для корпуса и отрицательные outside/metrics проверки.
- `backend/internal/integration/geometry_oracle_test.go`: независимые PostGIS-предикаты на WKB сформированных Go-полигонов, автоматическое сохранение counterexample.
- `backend/internal/nesting/testdata/stage12_property12_counterexample.json`: полный исходный synthetic problem/result/formed geometry найденного случая.
- `backend/internal/geometry/simplefeatures/testdata/stage12_contact_counterexample.json`: минимизированные два треугольника.
- `backend/internal/geometry/simplefeatures/contact_oracle_regression_test.go`: маленький детерминированный **красный по умолчанию** regression с точными rational orientations. Его FAIL не скрыт opt-in флагом.
- Этот отчёт, stage13 BLOCKED report, сводка и документация.

Production geometry engine, numerical precision, NFP/Union/Minkowski, RNG, snapshot/idempotency/keep_existing не менялись. Новые зависимости/миграции не вводились.

## Матрица и успешно выполненные проверки

19 классов: rectangle, rotated_surface, L, C, U, radial, asymmetric triangle, 1/2/3 holes, narrow neck, boundary contact, near-contact, occupied, multiple pattern types, fully occupied, oversize, 4 и 36 углов. Внутри также есть один/два угла и нестандартные 33°/45°.

Для каждого: base, перенос (+100000,−100000), реверс обхода, масштаб 0.1 и 10. Каждый вариант проверен пятью алгоритмами: **475 результатов**. Ещё 24 seeded простых радиальных контура проверены Baseline/NFP: **48 результатов**. Всего **523**: Baseline119, NFP119, GA95, SA95, Hybrid95.

Fixture RNG seed=20261121; GA/SA test-only seed=20261120, comparison profile GA8/10 и SA100. Production randomness не менялась; это не production throughput. Для property геометрии число вершин ограничено 5–12, положительные радиусы и упорядоченные углы. Live неограниченный fuzz не запускался.

Для каждого результата Go проверяет ValidatePolygon, CoveredBy, InteriorsIntersect, IDs/количества/разрешённые углы, requested=placed+unplaced, независимую shoelace-площадь и utilization (relative tolerance1e−7 для сравнения числовых метрик, **не** topology). Результат переноса/реверса проверяется на допустимость, без ложного требования одинаковых poses. Surface validity проверяет сам optimizer. 50 нулевых результатов относятся к full/oversize; прочие 473 непустые, всего 946 сформированных полигонов. Глобальная оптимальность не заявляется.

- Stable core: 8 классов × Baseline/NFP — PASS (0.084s в финальном адресном запуске).
- Bounded corpus: 523 Go-проверки — PASS,40.202s.
- Настоящий server validateResult: 523 positive и 946 negative outside/metrics checks — PASS,0.08s. У серверного validator более строгая собственная проверка metrics1e−8, она не ослаблялась.
- Независимый PostGIS oracle — **FAIL на 501-м случае**, property/seed20261121/12/nfp_greedy, второе размещение. Первые 500 случаев прошли; оставшиеся 22 не объявляются проверенными.

## Минимальный counterexample и доказательство

Исходный случай — валидный six-vertex contour, quantity2, rotations[0,90], поверхность 100×100. NFP Greedy выдаёт два placement; Go и server validation принимают их. WKB берётся из **фактически сформированных Go geometry**, без SQL-пересчёта вращения.

Результат уменьшен до двух валидных треугольников с 3 вершинами каждый. Соответствующие вершины имеют X=3.7786684611105628 и 3.7786684611105636, одинаковый Y=6.129087362376157. Разность X≈8.881784197001252e−16.

| Проверка минимального случая | Результат |
|---|---|
| Go ValidatePolygon обеих фигур | PASS |
| Go Relate | FF2F01212 |
| Go InteriorsIntersect | false |
| PostGIS ST_IsValid обеих WKB | true |
| PostGIS ST_Relate | 212101212 |
| PostGIS ST_Relate(...,'2********') | true |
| Точные rational orientations float64-координат | два строгих пересечения рёбер |

Точная проверка использует big.Rat.SetFloat64, то есть точные рациональные значения сохранённых double, а не округлённые десятичные приближения. Два strict crossings подтверждены независимо Python fractions и Go math/big. Это не просто другое построение координат и не случайный HTTP/fuzz failure.

**ST_Area(ST_Intersection) возвращает 0** из-за конечной точности построения столь малого пересечения. Поэтому площадь не используется как единственный oracle и не заявляется ненулевая измеренная площадь. DE-9IM и независимые exact orientation signs подтверждают пересечение внутренностей.

В simplefeatures v0.59.0 `geom/dcel_re_noding.go:24–44` явно выполняется snapping близких узлов: вычисляется ULP и геометрии преобразуются через nodes.insertOrGet. В результате минимальное пересечение трактуется как касание. Это конфликт строгого oracle ТЗ с численной snapping-моделью библиотеки, а не обязательно нарушение заявленного самой библиотекой контракта.

Красный regression повторён трижды в bounded процессе. Отдельный git archive исходного **21eecdb** с добавленными только test/fixture также даёт тот же FAIL. Следовательно, дефект строгой топологии существовал до нового candidate budget; production Engine не отличается от исходного HEAD. Pass-after не заявляется — исправление не внесено.

## Почему остановка и варианты решения

Нельзя автоматически объявить такое пересечение допустимым, применить tolerance к PostGIS oracle, запретить все касания или произвольно сдвинуть NFP poses на epsilon. Это меняет требуемый контракт и может изменить качество раскроя.

Возможные направления для отдельного решения:

1. Строгий robust predicate / adaptive exact проверка near-contact с сохранением разрешённых касаний. Потребуются bounded spatial filtering, предел работы, cancellation и before/after benchmarks: наивный pairwise exact orientation в горячем loop может резко увеличить CPU/allocations.
2. Явная согласованная precision/snapping-модель API и одинаковый oracle. Это **продуктовое изменение контракта**, а не повод ослабить текущий тест без решения владельца.
3. Независимая финальная PostGIS-валидация preview/save/manual. Потребуются архитектурная оценка batch/SQL latency, snapshot atomicity и поведение при отказе; она не исправит выбор кандидатов сама по себе.

Ни одно направление не внедрялось наугад. Блокер — способ гарантировать строгую топологию при сохранении касаний и управляемой стоимости, а не отсутствие времени/тестовой БД. Данных рабочей БД это исследование не меняло; признаков её потери не обнаружено.

## Среда, команды и ресурсы

Go1.26.2, PostGIS3.4 на PostgreSQL16.4, фактический gate строго nesting_test/userpostgres. Собственный network-none corpus runner: CPU2,hardRAM512MiB,swapoff,GOMEMLIMIT192MiB,pids128,readonlyroot,tmpfs; процесс 300s,каждый case10s. Corpus TotalAlloc=26328373632bytes — **накопленные аллокации**,не RSS; измеренный после corpus cgroup memory.peak=39428096bytes,OOM/OOMkill0. Goroutines2 в конце,не непрерывный peak.

Собственный SQL runner: CPU2,hardRAM768MiB,swapoff,GOMEMLIMIT256MiB,pids256,network namespace только выделенного test PostgreSQL. Измеренный после oracle peak=56483840bytes,OOM0. Сам test PostgreSQL вне этого runner cap; его параметры не менялись. Private schema/migrations проверены до fixture; oracle-запросы read-only. Минимальный контроль SQL BEGIN READ ONLY/ROLLBACK.

```sh
# В отдельных capped processes, compiled test binaries:
NESTING_INVARIANT_EXPERIMENT=1 NESTING_INVARIANT_OUTPUT=/work/corpus.json \
 /work/invariants.test -test.run '^TestGeometryInvariantExperiment$' -test.count=1 -test.v -test.timeout=300s
NESTING_INVARIANT_CORPUS=/work/corpus.json \
 /work/server.test -test.run '^TestInvariantCorpusServerValidation$' -test.count=1 -test.v -test.timeout=60s
# DSN предварительно проверен, секреты опущены; cwd integration, миграции доступны:
NESTING_TEST_DSN="$VALIDATED_TEST_DSN" NESTING_INVARIANT_CORPUS=/work/corpus.json \
 NESTING_ORACLE_COUNTEREXAMPLE=/work/counterexample.json \
 ./oracle-final.test -test.run '^TestGeometryCorpusPostGISOracle$' -test.count=1 -test.v -test.timeout=90s
# Маленький default-red regression:
go test -count=1 ./internal/geometry/simplefeatures -run '^TestStrictContactOracleRegression$' -v
```

## Gate и ограничения

| Набор | Статус |
|---|---|
| Go invariant core / bounded corpus / server validation | PASS в рамках Go predicate model |
| Независимый WKB PostGIS oracle | **FAIL**, повторён, counterexample сохранён |
| Exact rational минимальный proof | PASS, подтверждает пересечение |
| Минимальный strict regression | **FAIL**,3 повтора и исходный HEAD |
| Свежий go test -count=1 ./... | **FAIL только TestStrictContactOracleRegression**; не зелёный gate |
| Свежие go vet / golangci-lint | PASS,0issues |
| Свежие frontend build/lint/72unit | PASS |
| Новый полный race/PostGIS regression gate12 | SKIP после критического oracle FAIL; результаты этапа 11 не выдаются за новые |
| Browser/этап 13 | BLOCKED предшествующим gate |

Тяжёлый корпус по умолчанию SKIP в обычном suite, но явно выполнен выше. **Минимальный красный regression не gated и не скрыт.** Нет полной Cartesian fuzz/оптимальности/SLA500. Исходный пользовательский SVG отсутствует, человеческий UX не проверялся.

Артефакты, corpus, original/minimal counterexamples, SQL proof и logs сохранены вне /tmp в stages11_13. Собственные runner containers завершились по своему TTL3600s;финальный inspect подтвердил exited, рабочие контейнеры не затронуты. Изменения оставлены для ревью без commit/push/merge. Финальный Git snapshot ниже.

Последний запрос cgroup после истечения TTL недоступен:финальные lifetime peak/counters не заявляются. Приведены сохранённые фактические samples после corpus/oracle,не ceiling произвольной нагрузки. Все диагностические test processes успели завершиться,собственных listeners нет.


## Финальный Git snapshot

`git diff --check` PASS. HEAD/ветка неизменны;никакихcommit/push/merge. Нижеобычныйstatдляtrackedфайлов(новыеfixtures/tests/reportsвнёмнеотображаются),затемполныйstatus.

```text
 README.md                                          |   4 +
 backend/internal/nesting/anchors.go                |  78 +++++--
 backend/internal/nesting/anchors_test.go           |  11 +-
 backend/internal/nesting/bootstrap.go              |  11 +-
 backend/internal/nesting/bootstrap_test.go         |  11 +-
 backend/internal/nesting/candidate_from_anchors.go |  56 +++--
 .../nesting/candidate_from_anchors_test.go         |  12 +-
 backend/internal/nesting/generate_candidates.go    |  63 +++---
 backend/internal/nesting/nfp_candidates.go         | 247 ++++++++-------------
 backend/internal/nesting/place_one_pattern.go      |  25 ++-
 backend/internal/nesting/place_one_pattern_nfp.go  |   5 +-
 docs/verification.md                               |  10 +
 12 files changed, 272 insertions(+), 261 deletions(-)
```

```text
 M README.md
 M backend/internal/nesting/anchors.go
 M backend/internal/nesting/anchors_test.go
 M backend/internal/nesting/bootstrap.go
 M backend/internal/nesting/bootstrap_test.go
 M backend/internal/nesting/candidate_from_anchors.go
 M backend/internal/nesting/candidate_from_anchors_test.go
 M backend/internal/nesting/generate_candidates.go
 M backend/internal/nesting/nfp_candidates.go
 M backend/internal/nesting/place_one_pattern.go
 M backend/internal/nesting/place_one_pattern_nfp.go
 M docs/verification.md
?? backend/internal/geometry/simplefeatures/contact_oracle_regression_test.go
?? backend/internal/geometry/simplefeatures/testdata/stage12_contact_counterexample.json
?? backend/internal/integration/candidate_budget_test.go
?? backend/internal/integration/geometry_oracle_test.go
?? backend/internal/nesting/candidate_budget.go
?? backend/internal/nesting/candidate_budget_regression_test.go
?? backend/internal/nesting/geometry_invariants_test.go
?? backend/internal/nesting/testdata/stage12_property12_counterexample.json
?? backend/internal/usecase/nesting/invariant_corpus_test.go
?? docs/stage11_candidate_budget_safety_report.md
?? docs/stage12_geometry_invariants_report.md
?? docs/stage13_automated_acceptance_report.md
?? docs/stages11_13_summary.md
```
