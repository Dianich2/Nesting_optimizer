# Этап 15 — геометрические инварианты: PASS

Основа: HEAD `3d49cbbf1120fb9527ef6e295f9772ab58798579`, ветка `wip/geometry-editor-v2`. Этап начат только после gate14 PASS. Принятые локальные изменения строгого Engine сохранены. Working DB не используется.

## Корпус и независимые проверки

В `geometry_invariants_test.go` сохранены исходные523 rows/order, их число явно проверяется. Добавлены40 случаев: четыре near-contact fixtures ×5solvers (original property12 как obstacle, diagonal contact, one-ULP width, hole boundary) и четыре новых seeded radial cases ×5solvers с33°/45°/90° (seed20261126). Исходный seed20261121 и production RNG не менялись.

| Проверка | Факт | Статус |
|---|---|---|
| Go invariant experiment | 563 cases,40.271945641s | PASS |
| Server validateResult | 563 positive /1589 negative,0.07s | PASS |
| PostGIS WKB oracle | 563 cases /1016 formed polygons /698 interior pairs,0.39s | PASS |
| Oracle controls | outside rejected, identical interiors detected, touch allowed | PASS |
| Бывший501-й случай и следующие22 | все присутствуют и проверены | PASS |

Распределение: Baseline127, NFP127, GA103, SA103, Hybrid103. Каждый результат проверяет validity, IDs/quantities/rotations, requested=placed+unplaced, независимую площадь/utilization, Go coverage/interiors. Server negative checks: invalid surface у каждого из563 (включая нулевые результаты), плюс1026 outside/metrics checks. Повреждается копия входа, исходный corpus не меняется. Допуск метрик не используется для topology.

PostGIS получает **фактически сформированную Go WKB**: ST_IsValid/ST_CoveredBy/ST_Relate(...,'2********'), без SQL-реконструкции rotation и повторного округления координат. Новых разногласий oracle нет; старый property12 исправлен, а не исключён из набора.

## Сравнение с исходными523 случаями

Имена, алгоритмы и порядок исходных523 совпали. Ни одна Metrics не изменилась. Позиции изменились у пяти результатов: multi_type/small SA+Hybrid, angles36/large SA+Hybrid, property12 NFP. Новый поиск может отвергать прежние промежуточные кандидаты и выбирать другую допустимую pose; count/utilization не ухудшились. Все новые конечные WKB проверены независимо. Старый property12 доказан недопустимым exact+PostGIS, его poses не являются корректным golden.

Исходные523:40.202s, TotalAlloc26 328 373 632B; новые563:40.272s,27 324 107 768B. Это cumulative allocations, **не RSS**. Число случаев различается: сравнение не объявляется статистически доказанным speedup. Среда соответствует предыдущей: CPU2/hard512MiB/GOMEMLIMIT192MiB, per-case10s/общий300s. В конце goroutines2, не измеренный непрерывный peak.

## Дополнительная нагрузка и её WKB

Прежний47-case TestGeometryReliabilityExperiment явно включён под CPU2/hard512MiB/swapoff/GOMEMLIMIT192MiB/300s. Проверены пять solvers на семи классах, quantities10/50/100/500,32/64/128vertices,36углов. Два реальных запуска:95.31s и95.99s.

Второй запуск добавляет opt-in `NESTING_GEOMETRY_CORPUS` в test-only helper. Экспортируются Transform outputs успешных результатов; budget/RNG/optimizer/error assertions не меняются. Без env прежние profile/performance helpers не экспортируют геометрию. В production этого нет.

Итог47:43completed и четыре ожидаемых deadline (Baseline100/500, NFP500,128vertices). Все deadline возвращают zero partial. Для500 бюджет≈3s, для128vertices≈5s; это **не** доказательство успешного расчёта500 деталей. Успешные43 результата содержат436 сформированных полигонов.

Объединены563 invariant +43 completed load rows:

| Проверка | Факт | Статус |
|---|---|---|
| Server validator | 606 positive /1718 negative,0.21s | PASS |
| Независимый PostGIS | 606 cases /1452 actual polygons /8970 interior pairs,1.09s | PASS |
| Timeout results | не включены в positive corpus; проверены как errors без partial | PASS |

Таким образом независимо проверены и все successful stress placements. Не только Go, и не только первые500 исходного набора.

## Остальные gates

Full Go count1/vet/golangci-lint — PASS,0issues. Повторный полный race пяти пакетов — PASS. Full PostGIS suite и SQL/race parent/children (RealProcessLifecycle, strict manual/API, candidate budget, exact preview save, deadline) — PASS. Явный server563 corpus подrace:563positive/1589negative PASS0.37s. Последние test-only изменения экспортера проверены компиляцией/реальным bounded запуском/vet/lint; production после stage14 не менялся.

Full SQL включает keep_existing true/false, настоящие snapshots/CAS, replay UUID, cancel/timeout и40HTTP candidate-budget requests. Candidate overflow/cancellation tests входят в full Go/race. Watchdog отдельно трижды выполнен в14: early rejection до Transform при theoretical603979776 candidates,65536allocatedB; candidate budget не менялся на15.

Cumulative cgroup memory.peak после15: native53 510 144B, SQL40 833 024B, race227 672 064B; во всех memory.events OOM/OOMkill/max0. Это peak всего runner за испытания14–15, не per-op RSS и не память внешнего test PostgreSQL. Нет безлимитного fuzz.

## Команды и безопасность

```sh
# compiled test binaries, каждый в своём package cwd, capped runners:
NESTING_INVARIANT_EXPERIMENT=1 NESTING_INVARIANT_OUTPUT=/tmp/stage15-corpus.json \
 /tests/nesting.test -test.run '^TestGeometryInvariantExperiment$' -test.count=1 -test.v -test.timeout=300s
NESTING_INVARIANT_CORPUS=/tests/stage15-corpus.json \
 /tests/server.test -test.run '^TestInvariantCorpusServerValidation$' -test.count=1 -test.v -test.timeout=60s
NESTING_GEOMETRY_EXPERIMENT=1 NESTING_GEOMETRY_CORPUS=/tmp/stage15-load-corpus.json \
 /tests/nesting.test -test.run '^TestGeometryReliabilityExperiment$' -test.count=1 -test.v -test.timeout=300s
# После actual DB gate, только nesting_test/private schema:
NESTING_TEST_DSN="$VALIDATED_TEST_DSN" NESTING_INVARIANT_CORPUS=/tests/stage15-combined.json \
 NESTING_ORACLE_COUNTEREXAMPLE=/tmp/counterexample.json \
 /tests/integration.test -test.run '^TestGeometryCorpusPostGISOracle$' -test.count=1 -test.v -test.timeout=90s
```

Эксперименты opt-in: их обычный SKIP не выдаётся за PASS, выше перечислены реально выполненные запуски. Actual current_database=nesting_test/current_user=postgres/PostGIS3.4 проверены до schemas/fixtures. Изолированные schemas сохранены. Working DB, Goose migrations, `.env` и существующие volumes не менялись; commit/push/merge не выполнялись. `git diff --check` PASS, HEAD/ветка неизменны.

## Ограничения и переход

Gate15 PASS, разрешён переход16. Bounded seeds/profiles не доказывают оптимальность для произвольного контура. Original пользовательский SVG отсутствует и не объявляется проверенным. Deadline500 подтверждает корректный отказ, а не throughput SLA. Человеческая UX-приёмка отдельно. Corpora/logs/source backup находятся вне `/tmp` в stages14_16; финальная сводка перечисляет все файлы и cleanup.
