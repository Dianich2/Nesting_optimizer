# Проверка проектных экземпляров и раскроя

## Автоматические проверки

Из `frontend` выполняются `npm run build` и `npm run lint`.
Из корня выполняется `node --test --test-isolation=none frontend/tests/*.test.mjs` (Node 24).
Из `backend` выполняются `go test ./...` и `go vet ./...`.

Интеграционный тест запускается из `backend` командой

```sh
NESTING_TEST_DSN='postgres://postgres:PASSWORD@127.0.0.1:PORT/nesting_test?sslmode=disable&connect_timeout=5' go test -v ./internal/integration
```

Используйте отдельный PostGIS. Тест требует имя БД `nesting_test`, создаёт уникальную схему, применяет существующие миграции в ней и сохраняет схему для диагностики. Без переменной окружения тест пропускается. В сценарии проектных ресурсов контекст пользователя фиксирован, а проверки принадлежности проходят через usecase и SQL. Отдельный сценарий аккаунта проходит через настоящий HTTP-роутер, JWT и хранилище сессий.

Проверяются повторное создание ресурсов, независимость геометрии, сохранение экземпляров после удаления библиотечного источника, ошибки геометрии, запреты изменений и удаления при размещениях, переименование, доступ к чужому проекту, гонки удаления/изменения с размещением. Дополнительно проверяются ручные перемещения, поворот, пересечения, отверстия, сохранение результатов пяти алгоритмов и независимость поверхностей.

## Playwright

Запускайте `npm run test:e2e` из `frontend` с явным `NESTING_API_TARGET`, указывающим на API с отдельной базой. Для тестового API установите `HTTP_BROWSER_ORIGIN=http://localhost:5178`. Переменная `E2E_CHROME` позволяет использовать установленный Chrome.

Сценарий использует реальные формы, API и базу. Он проверяет вход, импорт SVG, масштабирование, проектные копии, ручную раскладку, Undo/Redo, потерю ответа на POST, восстановление по ключу, историю, SVG и удаление.

## Ручная проверка в браузере

- Войти и создать поверхность с произвольным контуром, затем лекало.
- Изменить масштаб, шаг и видимость сетки. Переместить поле и проверить ввод координат после изменения вида.
- Создать отверстие, изменить его вершины и убедиться, что пересечение контуров отклоняется.
- Создать проект и трижды добавить одну поверхность. Проверить разные номера экземпляров.
- Изменить второй экземпляр. Открыть первый, третий и библиотечный оригинал и сравнить геометрию.
- Изменить оригинал в библиотеке и проверить, что экземпляры проекта сохраняют свою геометрию.
- Пройти несколько страниц библиотеки, добавить лекало с последней страницы.
- Открыть раскрой первой поверхности, выбрать лекало и разместить его кликом.
- Перетащить размещение при разных масштабах, изменить X/Y, повернуть на 90°.
- Попытаться переместить лекало за границу, в отверстие и поверх другого лекала. Проверить сообщение сервера и возврат к сохранённой геометрии.
- Проверить выбор размещений с клавиатуры и через список объектов.
- Попытаться изменить геометрию или удалить занятую поверхность и используемое лекало. Проверить отказ и возможность переименования.
- Запустить каждый алгоритм с несколькими лекалами и количествами. Проверить сохранение существующих размещений.
- Отменить подтверждение замены раскладки, затем выполнить замену и проверить результат.
- Сопоставить общую площадь и количество размещений с метриками последнего запуска.
- Перезагрузить страницу. Проверить восстановление размещений и текущих метрик.
- Переключиться на вторую поверхность, выполнить раскрой и вернуться на первую.
- Проверить отмену удаления, Escape, фокус диалогов, повторное нажатие во время запроса и сообщение серверной ошибки.
- Удалить размещения и освободившиеся проектные экземпляры. Проверить сохранность библиотеки и остальных экземпляров.
- Проверить узкое окно, сенсорный ввод и длинные названия ресурсов.

## Границы текущей реализации

Мутации сериализуются по строке проекта только при чтении снимка и финальной записи. Вычисление выполняется между этими короткими фазами. Изменение геометрии или размещений приводит к HTTP 409 nesting_stale. Изменение независимой поверхности не инвалидирует расчёт.

В API доступны реальная отмена, серверный deadline, durable-состояние операции и сохранение точного предпросмотра по preview_id. Контекст передаётся в управляемые циклы NFP. Отмена сторонних earcut/simplefeatures остаётся кооперативной между вызовами. Подробные лимиты и конфигурация приведены в README.

Дополнительные интеграционные тесты находятся в backend/internal/integration/nesting_lifecycle_test.go. Они проверяют единственное доступное соединение PostgreSQL во время заблокированного оптимизатора, независимые ручные записи, устаревшие поверхности/лекала/размещения, удаление проекта, конкурирующие финальные транзакции, одинаковые ключи, рестарт, отмену и timeout, сбой и отмену во время сохранения, точный preview, TTL и квоту, panic, бюджеты и совместимость старой миграции.

Для targeted race-проверки выполните go test -race ./internal/usecase/nesting ./internal/transport/http/handler ./internal/geometry/nfp ./internal/nesting -run 'Test(LiveCancellation|CancellationDuringRealNFP|ComputeLimiter|DuplicateResponse|ExecutionSurvives|LegacyNestingFingerprint)'. Отдельная race-сборка интеграционных тестов требует CGO и совместимой libc тестовой среды.

Для воспроизводимых измерений используйте BenchmarkReliabilityWorkloads и BenchmarkReliabilityNFPStages с -benchtime=1x -benchmem. Первая группа включает 10, 50 и 100 экземпляров прямоугольного и вогнутого лекала, Baseline и NFP Greedy, количество кандидатов и геометрическую проверку результата. Вторая отдельно измеряет триангуляцию, суммы Минковского, Union и hit NFP-кэша на выпуклом/вогнутом контуре и одном/двух отверстиях. Для всех пяти алгоритмов остаются BenchmarkOptimizers и BenchmarkOptimizersComparison.

Ручная приёмка дополнительно включает изменение геометрии/размещений из второй вкладки во время расчёта, сохранение точного предпросмотра, повтор после потери ответа, истёкший предпросмотр и смену аккаунта при наличии незавершённой операции. Исходная пользовательская геометрия, на которой ранее наблюдалось зависание, по-прежнему нужна для отдельного воспроизведения.

## Межпроцессный preview

`TestPreviewHTTPKeySharesComputationAndSurvivesRestart` использует два независимых handler/usecase с общей PostgreSQL: один Optimize, одинаковый preview_id, конфликт параметров и replay после рестарта. `TestPreviewCrossInstanceCancellationAndLease` проверяет DELETE второго handler, чужого владельца, истечение lease и независимость UUID save/preview. `TestPreviewExpiryDoesNotRecompute` проверяет TTL, `TestPreviewCancelCommitRace` — однозначный итог отмены/COMMIT. Выполняйте также интеграцию с `-race`.

Playwright теперь теряет реальный ответ preview POST, обновляет страницу, восстанавливает готовый preview через статус и сохраняет исходную геометрию без второго preview POST. Ручная проверка: тот же PostgreSQL и два API; POST preview через первый, GET/DELETE через второй с `?kind=preview`; затем повтор ключа и проверка неизменного preview_id. Не используйте рабочую БД.


## Этапы 5–7: отдельные процессы и ограниченные эксперименты

Результаты и границы: `stage5_real_process_reliability_report.md`, `stage6_geometry_reliability_report.md`, `stage7_operational_reliability_report.md`, сводка `stages5_7_summary.md`. Не считать старый two-handler тест доказательством двух PID: новый `TestRealProcessLifecycle` запускает отдельные executable, независимые SQL pools и настоящие HTTP/JWT. Test-only worker/control listener отсутствует в production binary.

До любого integration сначала read-only `current_database()`, `current_user`, `postgis_version()`, строго `nesting_test`. Нужны заранее установленный PostGIS в public и миграции репозитория. Helper требует непустой набор миграций и все fixture tables в своей private schema. Неверный cwd — отказ до INSERT. Не передавать DSN рабочей базы и не читать backend/.env автоматически.

Из backend доступны свежие обычные проверки:

```sh
go test -count=1 ./...
go vet ./...
golangci-lint run ./...
go test -race -count=1 -timeout=20m ./internal/usecase/nesting ./internal/transport/http/handler ./internal/nesting ./internal/geometry/nfp
```

Для integration задайте `NESTING_TEST_DSN` явно после gate, запускайте `go test -count=1 -timeout=20m -v ./internal/integration`. При запуске скомпилированного binary cwd должен быть backend/internal/integration (в контейнере /work/internal/integration с /work/migrations). Для child под альтернативным ELF loader дополнительно нужны `NESTING_TEST_PROCESS_BINARY`, `NESTING_TEST_CHILD_LOADER`, `NESTING_TEST_CHILD_LIBRARY`; подробности фактической команды в отчёте5. Race требует совместимой CGO/libc, а не статической CGO_ENABLED=0 сборки.

Тяжёлые измерители намеренно SKIP по умолчанию. Только в отдельном capped container с CPU/RAM/timeout, по одному процессу:

- `NESTING_GEOMETRY_EXPERIMENT=1`: `TestGeometryReliabilityExperiment`, timeout300s; 47case, включая ожидаемые короткие timeout на500. До/после — одинаковые seed/config, три пары. `scripts/summarize_geometry_experiment.py --input LOG_DIR --output OUT_DIR` проверяет шесть полных логов.
- `NESTING_PROFILE_GEOMETRY=1`: `TestProfileOccupiedConcaveGeometry` с `-test.cpuprofile`/`-test.memprofile`; это реальные occupied Baseline50, не production throughput SLA.
- `NESTING_OPERATIONAL_EXPERIMENT=1`: `TestOperationalGrowth`, timeout300s/context240s;10000 usecase операций, затем synthetic expired backlog. Запускать только при достаточном месте в выделенной test DB. Схемы сохраняются. `scripts/summarize_operational_growth.py LOG1 LOG2 LOG3 OUT.csv` отвергает неполные/failed логи.

`scripts/nesting_test_diagnostics.sql` выполняет только агрегированные read-only запросы с timeout2s; требует nesting_test и явно выбранную схему test_<digits> через search_path. Не переносить его на production отключением guard. Скрипт не выводит UUID/payload/JWT/DSN. Размер relation после DELETE не обязан уменьшаться.

Ограничения: radial_without_seed выявил нулевой результат до этапа 8; bounded bootstrap исправляет первый feasible placement (см. отчёт 8). Эвристика остаётся неполной для произвольной геометрии. Исходный пользовательский SVG отсутствует. Native earcut/UnionMany не прерываются внутри вызова; timeout500/3s не доказывает успешный расчёт500/120s. Workers/queue локальны на API-процесс. Terminal idempotency rows бессрочны в существующем scope — не добавлять TTL DELETE без нового согласованного контракта.


## Этапы 8–10

Сводка — [stages8_10_summary.md](stages8_10_summary.md), отдельные отчёты описывают fail-before/pass-after, отрицательный performance-эксперимент и настоящую браузерную приёмку. Bootstrap вызывается только после неуспеха прежних кандидатов; до 5 seeds на разрешённый угол, те же topology predicates и остаточный candidate budget.

Точечный эксперимент Bounds **снят**: меньше allocations не дало убедительного выигрыша wall. Для повторения 10-case сравнения нужен точный post-stage8 working-tree snapshot, а не один git archive HEAD. `NESTING_TARGETED_PERFORMANCE=1` включает TestTargetedPerformanceExperiment; запускать только по одному process в capped container. `scripts/summarize_targeted_performance.py --input LOG_DIR --output OUT_DIR` требует 3 полные before/after пары и неизменность качества/NFP work.

Browser helper TestBrowserAcceptanceAPI по умолчанию SKIP. Explicit `NESTING_BROWSER_API=1`, проверенный `NESTING_TEST_DSN`, `NESTING_E2E_READY_FILE`, `NESTING_E2E_ORIGIN=http://localhost:5186` включают настоящий test API. Запуск binary из backend/internal/integration, миграции доступны по ../../migrations. Helper создаёт private schema и не читает .env. Counter/stats listener отсутствует в production cmd/api.

Playwright требует NESTING_API_TARGET из проверенного readiness конкретного собственного PID; NESTING_E2E_PORT=5186 задаёт localhost frontend/baseURL, NESTING_E2E_STATS — test-only stats URL. E2E_CHROME задаёт установленный Chrome. Команда: `npm run test:e2e -- --workers=1 --trace=off`. Полный воспроизводимый запуск, настройки DB и cleanup описаны в stage10_automated_e2e_acceptance_report.md. Не подключать браузерные мутации к рабочему API.

Новая матрица проверяет 1440/1366/390 и обе темы, SVG import/editor, manual Undo/Redo, radial preview/exact save с Optimize counter, reload, историю и содержимое SVG/JSON. Дополнительно: реальный second-tab stale, confirmed через reload при недоступных placements. Исправлены HTML pattern логина и сохранение контекста при отказе initial placements read. Human UX, другие browser engines, настоящий touch и исходный пользовательский SVG остаются отдельной приёмкой.

## История этапов 11–13

[Сводка](stages11_13_summary.md) и отчёт12 описывают состояние до исправления: candidate budget принят, а TestStrictContactOracleRegression был красным. Два валидных треугольника имеют два exact rational proper crossings; прежний Go Relate возвращает false, PostGIS true. Fixture сохранён без изменения expected. Тогда обязательный stop соблюдён, старые E2E не выдавались за новую приёмку.

## Текущие этапы 14–16 — PASS

[Сводка](stages14_16_summary.md) содержит результаты, изменённые файлы, артефакты и ограничения. Regression теперь зелёный по умолчанию. `strict_interiors.go` не использует topology epsilon: filtered exact signs, AABB, holes/containment/совпадающие рёбра и vertex-on-edge intervals. Пределы — 1 000 000 work units /4096 exact operations на вызов; исчерпание возвращает ErrComplexity. Context по прежнему публичному интерфейсу проверяется между bounded вызовами.

Малые математические tests и300 seeded pairs входят в обычный Go suite;3000 orientation triples сравниваются с прямой rational арифметикой. Explicit тяжёлые измерения запускаются отдельно в capped runner, не считаются PASS из обычного SKIP:

```sh
# Из backend; только в отдельном CPU/RAM/deadline runner:
NESTING_INVARIANT_EXPERIMENT=1 NESTING_INVARIANT_OUTPUT=/tmp/corpus.json \
 go test -count=1 -timeout=300s -run '^TestGeometryInvariantExperiment$' -v ./internal/nesting
NESTING_INVARIANT_CORPUS=/tmp/corpus.json \
 go test -count=1 -run '^TestInvariantCorpusServerValidation$' -v ./internal/usecase/nesting
# После actual current_database/current_user/PostGIS gate, только nesting_test:
NESTING_TEST_DSN="$VALIDATED_TEST_DSN" NESTING_INVARIANT_CORPUS=/tmp/corpus.json \
 go test -count=1 -run '^TestGeometryCorpusPostGISOracle$' -v ./internal/integration
```

Corpus сохраняет исходные523 rows/order и добавляет40 near-contact/seeded examples. Новый server test отвергает invalid surface у каждого результата, включая нулевые. PostGIS получает фактически сформированную Go WKB без SQL-реконструкции вращения. Нельзя ослаблять oracle через площадь: микропересечение может иметь округлённую ST_Area=0.

`NESTING_GEOMETRY_EXPERIMENT=1 NESTING_GEOMETRY_CORPUS=/tmp/load.json` включает прежний47-case load и дополнительно сохраняет actual geometry успешных результатов. Timeout rows не входят в positive corpus: они обязаны иметь zero partial. Для этапа15 объединены563 invariant +43 completed load rows; все606 прошли server/WKB oracle. Команды,47-case результаты и метрики — в отчёте15.

`NESTING_STRICT_TRANSITION=/tmp/transition.json` включает TestStrictHybridTransitionEvidence (native). Его44 различия old/strict проверяются TestStrictTransitionExact (geometry) и TestStrictTransitionPostGIS (integration). `NESTING_STRICT_MATRIX=/tmp/matrix.json` сохраняет22 math pairs из TestStrictTopologyMatrix; тот же файл проверяет TestStrictMatrixPostGIS. Эти evidence-пути явно opt-in; минимальный исходный regression остаётся обязательным.

Для E2E используются scripts/run_stage16_browser.py, test-only browser helper, отдельный собственный `nesting-stage16-db` и private schema. Test DB после приёмки остановлена; перед повторением отдельно подготовьте этот изолированный стенд, не подставляйте рабочий DSN. Wrapper проверяет DB/PID/origin, использует loopback5188, retries0/traceoff и сохраняет JSON report. Финальные прогоны выполнены с `/tmp/nesting-stage16-api-race.test`, API logs отдельно проверены на DATA RACE. Полный рецепт — в отчёте16.

Физический телефон, pinch gesture, Firefox/WebKit (не установлены), исходный SVG и человеческое удобство не объявляются проверенными. Chrome native touch emulation покрыла tap/drag/pan и zoom-кнопки. Нет обещания оптимальности или arbitrary500/120s.
