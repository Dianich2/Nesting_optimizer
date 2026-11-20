# Этап 14 — строгая топология: PASS

Основа: `3d49cbbf1120fb9527ef6e295f9772ab58798579`, ветка `wip/geometry-editor-v2`, исходное дерево чистое. Проверены status/HEAD/diff-check. Архив исходного HEAD сохранён вне /tmp в каталоге артефактов stages14_16, SHA256 `3a50ce16bc8db2bcc0af70b0d3a38018e291fd2b0b24e02ebee35edab4fe20d4`.

## Анализ до исправления

Минимальная исходная regression действительно FAIL: exact proper crossings=2, geom.Relate=FF2F01212, Engine=false. Fixture и ожидание сохраняются.

Важное уточнение предыдущего отчёта: наличие snapping в `dcel_re_noding.go` само по себе НЕ доказывает причину ошибки Relate. В фактическом v0.59.0 `geom.Relate` вызывает `jtsRelateNG`, сериализует WKB и использует внутренний JTS port, включая extended-precision orientations и вычисленные intersection coordinates. DCEL snapping используется другими операциями. Дефект подтверждён, но прежняя атрибуция именно snapping не доказана.

Сравнение вариантов:

1. Локальный строгий предикат: AABB, filtered exact orientation, строгая классификация точек/рёбер, отдельная обработка коллинеарного общего ребра и vertex-on-edge. Не полагаться на округлённое построение координат intersection. Выбран для прототипа: общий Engine автоматически покрывает ручные и автоматические проверки, без изменения SQL/API/RNG.
2. Новая библиотека: добавляет ещё один контракт точности и зависимости; доказательств, что она решает все случаи, пока нет. Не выбрана.
3. Финальный PostGIS: полезен как независимый тестовый oracle, но сетевой запрос для каждого кандидата неприемлем; финальная SQL-проверка не исправляет поиск и требует отдельного анализа транзакций. Production SQL не меняется.
4. Precision/snapping contract: меняет пользовательскую семантику. Не выбран и не внедряется.

Требования прототипа: никакого topology epsilon, exact fallback только для несертифицированных float signs/вырожденных контактов; ограничить суммарную работу и exact calls, при превышении вернуть ErrComplexity. Существующий интерфейс без ctx не получает горутин; отмена остаётся между bounded Engine calls. Проверить performance до принятия. Gate пока НЕ пройден.

## Прототип и математические основания

`strict_interiors.go` заменяет только реализацию InteriorsIntersect после прежней ValidatePolygon обоих входов. Публичный интерфейс не меняется. AABB сравнивает исходные координаты без padding/epsilon; непересекающиеся **открытые** bounds позволяют немедленно вернуть false, включая прямоугольное edge-only touch. Для оставшихся пар проверяются AABB сегментов, proper crossings, coincident intervals с направлением внутренности, strict point-in-polygon (even/odd по всем кольцам), а также открытые интервалы рёбер, разделённые точными входными вершинами на контактах. Последний случай необходим для вписанного ромба: только vertex containment и proper crossing недостаточны.

Обоснование полноты для валидных полигонов: при positive interior overlap либо границы пересекаются proper, либо открытая часть границы одной фигуры находится внутри другой, либо границы совпадают на интервале с внутренностями с одной стороны. При отсутствии proper crossing вход/выход на ребре возможен только через вершину другой границы. Поэтому деление по этим вершинам и точная классификация середин покрывают оставшиеся vertex-on-edge случаи. Направление внутренности определяется ориентацией валидного кольца на lexicographic extreme, для отверстия инвертируется. Порядок обхода не влияет на результат.

Знак orientation сертифицируется first-stage orient2d error bound `(3+16u)u`, `u=2^-53`, с дополнительным отказом от float certificate при underflow/overflow. Неоднозначные знаки вычисляются big.Rat по исходным float64. Осевая факторизация использует только точные сравнения. Midpoint сначала проверяет exact representability через TwoSum; иначе рациональная середина не округляется в endpoint. Это численная оценка погрешности, не topology tolerance.

Per-call пределы: 1 000 000 единиц работы и 4096 exact operations. Превышение — ErrComplexity, не silent false. Состояние локально, нет global caches/очередей/горутин. Публичный Engine не принимает context; отмена остаётся между ограниченными вызовами, как в существующей архитектуре. Валидация входов остаётся прежней библиотечной; доказательство строгого предиката предполагает валидность исходных полигонов.

## Failures во время разработки

Исходный общий Go suite сначала упал на прежнем fixed-seed Hybrid golden: total8205→8200, repeated6923→6915; bitwise mismatch0, placed9, usedArea4800/height60 сохранились. В отдельном opt-in instrumented Engine собраны 44 различия old=false/strict=true. Независимый exact horizontal-slab oracle подтвердил все 44; PostGIS на тех же WKB подтвердил все44, а также validity/coverage конечных9 деталей и отсутствие overlap у всех36 пар. Только после этого обновлены два точных счётчика golden; assertions качества и детерминизма не удалены. Evidence reproducible через NESTING_STRICT_TRANSITION.

Первый прототип имел дорогие degenerate ветки: вписанный ромб ~200µs против ~23µs; concave100 Baseline ~22s против ~15s. Это измерение НЕ принято. Точные axis shortcuts и representable midpoint certificate снизили ромб до ~4µs и устранили ненужный rational hot path. Все неоднозначные случаи по-прежнему exact. Итоговые workload measurements приведены отдельно; ранние логи сохранены. Один ранний workload pilot пересекался по времени с дополнительным pair benchmark: для итогового сравнения он повторён последовательно.

Ошибки нового harness исправлены открыто: неверный маршрут project_surface_id (400), неверная колонка status вместо state, неверный cwd race (fixture отсутствует). После axis shortcut новый budget-test неправильно ожидал fallback на осевом прямоугольнике; изменён на прямую проверку действительно exact коллинеарной неосевой orientation. Исходный strict regression/fixture/ожидания не менялись. Неудачные запуски не включаются в PASS.

## Проверки корректности и ресурсов (промежуточные результаты)

- Исходный TestStrictContactOracleRegression PASS без изменения исходного test/fixture: exact proof2 пересечения сохранён; legacy Relate по-прежнему ошибочно FF2F01212.
- 22 детерминированные пары × swap/reverse/open-closed (176 подслучаев), containment/1–3 holes/concave/negative/large/small, настоящие diagonal edge/vertex touches, one-ULP gap/overlap с перекрывающимися AABB, плюс33°/45°/90°/NaN/Inf/invalid: PASS. Все22expected независимо подтверждены exact slab и WKB PostGIS; не по площади.
- 300 seeded radial pairs против независимого rational horizontal-slab oracle: PASS. Алгоритм oracle учитывает точные Y всех вершин и пересечений, затем сравнивает открытые интервалы внутри каждого slab; production orientation/классификатор не используется.
- 3000 bounded triples со степенями 2 от -1000 до999, near-collinear и nextafter: filtered sign совпал с direct rational determinant;1362 exact fallback. Есть отдельная проверка непредставимой float64 середины one-ULP интервала — не округляется на границу.
- Manual create/update исходного property counterexample через реальный handler/repository/transaction отвергнуты409; count/прежняя pose сохраняются. Buggy optimizer инъекция явно синтетическая, но HTTP/SQL/валидация настоящие: preview/save × keep/replace × first/replay =8запросов;500/internal_error, history0/ready0/completed0, Optimize ровно2 на keep case (по одному save/preview). Настоящее edge-only touch проходит201.
- Полный PostGIS suite выполнен с NESTING_TEST_DSN и private schemas, including independent API PIDs и SQL cancellation/COMMIT; PASS. Heavy corpus/Browser helper SKIP по умолчанию: это не stage15/16PASS.
- Frontend build/lint/72units: PASS, production frontend не менялся.

## Измерения

Одинаковый CPU2/RAM512MiB/swapoff/network-none native runner; GOMEMLIMIT192MiB soft внутри hard cap. Pair benchmarks: 3повтора по100ms, validation включена и у старого, и у нового Engine. Старый предикат воспроизведён без изменения module cache. Фактические ns/op, B/op, allocs/op и exact/op сохранены в raw logs. На обычных прямоугольниках/исходном regression/concave10–50 новый путь быстрее и выделяет меньше памяти. Вписанный diamond ~4µs/22allocs против ~20µs/272. Diagonal boundary-only contact остаётся дороже: ~23µs/379allocs/5exact против ~13µs/190allocs. Это ограниченный известный tradeoff строгого contact path, не скрытый общий speedup; bytes ~8169 против8384.

Existing BenchmarkReliabilityWorkloads, `-benchtime=1x -count=1`, все12before/after cases: без изменения generated candidates/placed/utilization. Не statistical throughput SLA. Крупные примеры: rectangle100 Baseline2686→1262ms, NFP478→375ms; concave100 Baseline13975→6098ms, NFP3169→1676ms. Concave100 Baseline cumulative allocations8.98GB→4.28GB (не RSS), allocs151796385→32961952. Один малый concave10/NFP показал47→57ms в одиночной паре; 3fresh targeted повтора дали before53.6–58.8ms / after38.6–44.4ms, allocs~507589→345089. Крупные нагрузки не повторялись без основания. Никаких13GiBсписков/OOMfuzz.

SQL runner CPU2/hard768MiB/pids256, только network namespace выделенного nesting-reliability-final; actual gate current_database=nesting_test/current_user=postgres/PostGIS3.4 проверен до schemas/fixtures. Рабочий nesting_postgres не трогается. Race runner CPU2/hard1GiB/свойtmpfs/hostlib readonly для совместимого CGO. Final counters/gates будут зафиксированы после завершения.

## Gate 14 — PASS

Full Go count1 PASS (native70.363s, usecase26.261s), vet PASS, golangci0issues. После последних дополнительных test-only assertions ещё раз проверены geometry/integration, vet/lint; новых production изменений после полного gate нет. Все5race packages PASS в отдельных binaries, последовательно; новая midpoint regression отдельно normal/race PASS. Полный PostGIS suite PASS, адресный SQL race (instrumented parent/children RealProcessLifecycle, strict API, CandidateBudgetHTTPPreservesLayout, RadialPreviewSavedExactly, NativeDeadlinePreservesSavedLayout) PASS. Candidate watchdog3 повтора PASS:17.9–48.2µs,65536B,Transform0 на theoretical603979776. Это early rejection, не выделение этого списка.

Фактические команды (из backend, без .env):

```sh
go test -count=1 -timeout=10m ./...
go vet ./...
golangci-lint run ./...
# CGO_ENABLED=0 go test -c — normal binaries;
# go test -race -c — race binaries, отдельный cwd каждого пакета.
# Равнозначно отдельным последовательным invocations:
go test -race -count=1 -p=1 -timeout=20m \
 ./internal/geometry/simplefeatures ./internal/geometry/nfp ./internal/nesting \
 ./internal/usecase/nesting ./internal/transport/http/handler
# capped runner, explicit opt-in:
NESTING_CANDIDATE_EXPERIMENT=1 /tests/nesting.test \
 -test.run '^TestCandidateBudgetWatchdog$' -test.count=3 -test.timeout=15s
NESTING_STRICT_TRANSITION=/tmp/transition.json /tests/nesting.test \
 -test.run '^TestStrictHybridTransitionEvidence$' -test.v -test.timeout=90s
NESTING_STRICT_TRANSITION=/tests/transition.json NESTING_STRICT_MATRIX=/tmp/matrix.json \
 /tests/geometry.test -test.run '^TestStrict' -test.v -test.timeout=30s
# Только после фактического DB gate; VALIDATED_TEST_DSN здесь обозначение,
# пароль в отчёт не записывается.
NESTING_TEST_DSN="$VALIDATED_TEST_DSN" NESTING_STRICT_MATRIX=/tests/matrix.json \
 NESTING_STRICT_TRANSITION=/tests/transition.json /tests/integration.test \
 -test.count=1 -test.v -test.timeout=300s
```

Frontend build/lint и `node --test --test-isolation=none frontend/tests/*.test.mjs`:72PASS. Полные rawlogs сохраняются в stages14_16. `git diff --check` PASS, HEAD/ветка неизменны. Миграций/production SQL/frontend не меняли. Рабочая DB/существующие Docker volumes не менялись. Новые runner containers принадлежат только этому испытанию.

Этап14 принят для перехода к15; это не означает прохождение полного corpus15 или browser16. Ограничения: валидность входов проверяет прежняя библиотека; сложные near-degenerate контуры могут получить ErrComplexity; context внутри неделимого bounded Engine call отсутствует; диагональные touches имеют измеренный локальный CPU tradeoff. Human UX/исходный пользовательский SVG не проверены.

Фактический cumulative cgroup memory.peak после gate14: native50 245 632B, SQL40 833 024B, race227 672 064B; во всех memory.events OOM/OOMkill0 и max0. Это peak всего runner за его испытания, не per-op RSS и не память внешнего test PostgreSQL. Rawlogs сохранены вне/tmp в `/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages14_16`.

## Итоговая таблица pair benchmarks

Медианы трёх100ms повторов; validation включена. Old branch измеряет Relate без финального mask-match, поэтому это немного занижает стоимость старого полного Engine. Workload comparison выше использует целиком неизменённый исходный Engine из архива HEAD. Exact/op — только собственный новый rational fallback.

| Пара | Old ns/op | Strict ns/op | Old allocs/op | Strict allocs/op | Strict exact/op |
|---|---:|---:|---:|---:|---:|
| far | 8431 | 2186 | 108 | 16 | 0 |
| touch | 16819 | 1896 | 233 | 16 | 0 |
| overlap | 15479 | 2178 | 228 | 18 | 0 |
| regression | 15584 | 2253 | 217 | 18 | 0 |
| vertex-crossing | 20379 | 4029 | 272 | 22 | 0 |
| diagonal-touch | 12989 | 23114 | 190 | 379 | 5 |
| concave10-hole | 18222 | 6909 | 236 | 31 | 0 |
| concave50-hole | 35142 | 22417 | 419 | 47 | 0 |
