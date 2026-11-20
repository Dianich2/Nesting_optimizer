# Этапы 14–16 — итог для ревью

Основа и текущий HEAD: `3d49cbbf1120fb9527ef6e295f9772ab58798579`, ветка `wip/geometry-editor-v2`. На входе дерево чистое, ранее реализованные этапы сохранены. Результат — только локальные изменения; commit/push/merge/reset/clean не выполнялись.

| Этап | Статус | Доказательства |
|---|---|---|
| [14 — строгая топология](stage14_strict_topology_report.md) | **PASS** | Исходный red regression зелёный без изменения fixture/expected; exact/PostGIS/touch/holes/containment, bounded work, before/after benchmarks, Go/race/SQL |
| [15 — инварианты и нагрузка](stage15_geometry_invariants_report.md) | **PASS** | Исходные523 +40examples;47-case нагрузка;606successful results/1452actualpolygons/8970pairs по независимому WKB oracle |
| [16 — браузерная приёмка](stage16_automated_acceptance_report.md) | **PASS** | Два полных race-API E2E одной версии:14/14+14/14,78.569015/76.030953s, retries/skip/flaky0; SQL audit и отсутствие API race reports |

## Что изменено

Production изменение ограничено общим `Engine.InteriorsIntersect`. Предикат различает строгое пересечение внутренностей и boundary-only contact по точным значениям float64: AABB, сертифицированные orientation signs, rational fallback, holes/containment/совпадающие рёбра и vertex-on-edge intervals. Нет topology epsilon, произвольных сдвигов или запрета всех касаний. Пределы работы и exact calls fail closed через ErrComplexity. Интерфейс Engine, NFP/RNG, candidate budget, SQL lifecycle, preview exact save, idempotency, CAS и авторизация сохранены.

Важная поправка исследования: фактический geom.Relate в simplefeatures0.59.0 вызывает JTS port. Старое наличие DCEL snapping не доказывало именно эту причину ошибки Relate. Исходный дефект подтверждён independently, новый предикат не строит округлённые intersection coordinates.

Fixed-seed Hybrid counters изменились8205→8200 /6923→6915. До обновления двух golden assertions все44 old=false/strict=true различия подтверждены exact slab oracle и PostGIS; конечные9 деталей/36пар допустимы, качество прежнее. В исходных523 corpus Metrics не изменились ни у одного результата, poses изменились у пяти. Проверки качества/детерминизма не удалены.

## Полный список изменённых/новых файлов

Production:

- `backend/internal/geometry/simplefeatures/engine.go`
- `backend/internal/geometry/simplefeatures/strict_interiors.go` — новый

Backend tests / evidence harness:

- `backend/internal/geometry/simplefeatures/strict_interiors_test.go` — новый
- `backend/internal/integration/strict_topology_test.go` — новый
- `backend/internal/integration/browser_api_test.go`
- `backend/internal/nesting/strict_transition_test.go` — новый
- `backend/internal/nesting/nfp_exact_determinism_test.go`
- `backend/internal/nesting/geometry_invariants_test.go`
- `backend/internal/nesting/geometry_experiment_test.go`
- `backend/internal/usecase/nesting/invariant_corpus_test.go`

Browser / orchestration:

- `frontend/e2e/bootstrap.spec.mjs`
- `frontend/e2e/stage16.spec.mjs` — новый
- `frontend/playwright.config.mjs`
- `scripts/run_stage16_browser.py` — новый

Documentation:

- `README.md`
- `docs/verification.md`
- `docs/stage14_strict_topology_report.md` — новый
- `docs/stage15_geometry_invariants_report.md` — новый
- `docs/stage16_automated_acceptance_report.md` — новый
- `docs/stages14_16_summary.md` — новый

Итого20файлов:2production Go,8backendtests,3browser/config,1script,6documentation. Productionfrontend/CSS, dependencies, migrations и `.env` не менялись. Обычный git diff --stat не включает untracked; полный status и source archive сохранены отдельно.

## Проверки и измерения

- Исходный `TestStrictContactOracleRegression` PASS, exact proof2propercrossings сохранён.22math pairs ×swap/reverse/open-closed;300seeded pairs против независимого rational slab oracle;3000orientation triples (1362fallback); NaN/Inf/invalid; непредставимая середина one-ULP интервала — PASS.
- Full `go test -count=1 ./...`, `go vet ./...`, `golangci-lint run ./...` — свежие PASS,0issues. Полный race пяти пакетов выполнен дважды на production версии (14/15), плюс новые geometry checks и actual browser API подrace. Финальный full PostGIS и адресный SQL/race с настоящими instrumented дочерними PID — PASS.
- Frontend build/lint/72unit — свежие PASS. Два финальных полного E2E:28/28compoundtests, retries0; source manifest14Go/JS/Python файлов не менялся между прогонами; API logs без DATA RACE.
- Corpus563: Go40.272s,server563positive/1589negative; PostGIS1016polygons/698pairs. Нагрузка47:43completed/4expectedtimeouts, повторена с экспортом actual geometry. Combined606: server1718negative, PostGIS1452polygons/8970pairs PASS1.09s. Все successful stress placements проверены независимо.
- Workloads10/50/100 rect/concave, Baseline/NFP: все12before/after, count/utilization/candidates прежние. Concave100Baseline13.975→6.098s, NFP3.169→1.676s; это одиночные bounded workload measurements, не throughput SLA. Pair benchmarks3повтора, exact/op/allocs/ns сохранены. Diagonal contact имеет известный локальный tradeoff~23µs против~13µs; CPU/память ограничены.
- Candidate watchdog3PASS: theoretical603979776 candidates отвергнуты до Transform,65536allocatedB; огромный список не создавался. Overflow/cancellation/40HTTPkeep/replace/replay checks прошли.
- Read-only SQL audit обеих E2E схем: по15users/13projects/19placements/15history, все5algorithms; save15completed/1failed (настоящий stale), дублей нет.

Финальные native/SQL/race/DB cgroup peaks:53 510 144 /40 833 024 /227 672 064 /259 129 344B при hardcaps512/768/1024/1024MiB; OOM/OOMkill/max0. Это cumulative runner peaks, не per-op RSS. Browser/APIfunctional тесты не выдаются за hard-capped нагрузку.

## Сохранность и завершение стендов

Любой SQL тест сначала проверял actual database=nesting_test/userpostgres/PostGIS. Все fixture writes — в private schemas; рабочая `nesting_optimizer` не использовалась. Нет изменений Goose migrations, существующих volumes/настроек рабочих контейнеров. Исходная база HEAD сохранена вне/tmp до исправления.

Собственные `nesting-stage14-native`, `nesting-stage14-sql`, `nesting-stage14-race`, `nesting-stage16-db` остановлены после завершения jobs. Test volumes/schema оставлены для диагностики. Final API/stats ports и Vite5188 закрыты. Рабочий `nesting_postgres` и прежний выделенный `nesting-reliability-final` оставлены running. Подтверждения inspect/cleanup находятся в artifacts.

Резервная копия, rawlogs (включая неудачные pilots), original/extended/load/combined corpora, exact/SQL evidence, оба JSON E2E reports/screenshots/exports и SHA-256 manifest:

`/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages14_16`

`baseline-3d49cbb.tar` сохраняет исходный HEAD; `final-source.tar.gz` — все20изменённых/новых файлов. Archive entries сверены с actual source hashes; SHA256SUMS проверяется после записи. Финальные status/diff-check сохранены. Все изменения оставлены пользователю для независимого ревью.

## Ограничения и ручная проверка

- Strict predicate предполагает валидные полигоны; валидация входа остаётся прежней библиотечной. Очень сложные/вырожденные контуры могут получить ErrComplexity; context внутри прежнего неделимого Engine interface отсутствует, отмена между bounded calls. Native сторонние geometry вызовы остаются кооперативными.
- Нет гарантии глобальной оптимальности или arbitrary500сложных деталей/120s. Deadline500 проверяет безопасный отказ без partial;128vertices также достигли предусмотренного5sdeadline.
- Firefox/WebKit не установлены — SKIP. Chrome touch emulation не физический телефон; pinch не проверен. Zoom touch-кнопками, native tap/drag/pan проверены.
- Исходный пользовательский SVG отсутствует. Владелец вручную проверяет свой SVG, удобство панелей и редактора, drag/rotate/pan/zoom, читаемость и контраст обеих тем, реальные устройства и ощущение скорости. Скриншоты не заменяют эту приёмку.

Работа по этапам14–16 остановлена для ревью; дальнейших этапов/публикации/commit нет.
