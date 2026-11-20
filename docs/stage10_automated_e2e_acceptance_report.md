# Этап 10 — автоматическая браузерная приёмка

## Основа и план

HEAD: `30aadbccedd56b5a0bb769e59403944b22fda75d`, ветка `wip/geometry-editor-v2`. На входе только 11 изменённых/новых файлов принятых этапов 8–9 (status в отчёте 9). Пользовательские изменения отсутствовали; предыдущие этапы сохранены. План: изолированная topology → positive pilot → desktop/mobile × темы → recovery/second-tab/confirmed-reload → минимальные доказанные fixes → два полных свежих прогона → обязательные gates. Редизайн не входит.

## Фактическая среда

Создан отдельный `nesting-stage10-db`, образ `postgis/postgis:16-3.4`, новый anonymous volume, CPU=2, hard RAM=1 GiB, swap отключён, pids=128. PostgreSQL **16.4 (Debian 16.4-1.pgdg110+2)**, PostGIS **3.4 USE_GEOS=1 USE_PROJ=1 USE_STATS=1**. Публикация только `127.0.0.1:32768 → 5432`. Рабочий `nesting_postgres` и его тома не перенастраивались. Старый `nesting-reliability-final` отдельно используется для regression SQL.

Перед каждым API-запуском docker psql проверяет current_database/current_user/PostGIS. Затем **фактическое Go-подключение** повторяет gate строго `nesting_test`, создаёт private `test_<timestamp>` схему и проверяет миграции/12 таблиц до записи. Fallback в public запрещён. Каждый прогон получает новую схему и новых пользователей.

`TestBrowserAcceptanceAPI` — test-only executable: настоящий container.New/NewRouter, bcrypt/JWT/cookies/CSRF/CORS и native factories с production Defaults, новым NFP cache на запрос и счётчиком настоящих Optimize. Workers=2/queue=4, pool=8/4, GOMAXPROCS=2, GOMEMLIMIT=512 MiB **soft**, bcrypt=4 только для тестовых регистраций. Config.Load и `.env` не читаются. API и отдельный read-only stats listener слушают только `127.0.0.1` на случайных портах. Диагностики нет в production cmd/api.

React/Vite: `http://localhost:5186`, strictPort, reuseExistingServer=false. NESTING_API_TARGET берётся из readiness file конкретного spawned API; проверены PID, database, schema и loopback URL. BrowserOrigin/baseURL совпадают с frontend. Точные runtime порты/PID/схемы сохранены в `full5-topology.json`/`full6-topology.json` без секретов.

Go 1.26.2, Node 24.22.2, Playwright 1.65.1, Google Chrome 155.0.8059.84. Browser workers=1, action timeout=10s, test timeout=90s, orchestrator timeout=600s, API-helper timeout=20min. Это функциональная проверка, не throughput benchmark. Hard cap относится к DB, API/browser не выдаются за cgroup-capped нагрузочный тест.

## Файлы и production-исправления

- `backend/internal/integration/browser_api_test.go`: opt-in API, реальная auth/private schema, native factories, Optimize counter, readiness и остановка своего процесса.
- `frontend/e2e/bootstrap.spec.mjs`, `fixtures/radial-pattern.svg`: шесть positive вариантов 1440/1366/390 × light/dark, а также реальная мутация во второй вкладке.
- `frontend/e2e/nesting.spec.mjs`: сохранены прежние recovery; добавлены confirmed через reload при продолжающемся отказе placements, видимый контекст и блокировка ручных записей до синхронизации.
- `frontend/playwright.config.mjs`: проверяемый NESTING_E2E_PORT, прежний default 5178 сохранён; отдельный localhost frontend без reuse чужого Vite.
- `RegisterPage.tsx`: исправлен **invalid HTML pattern**. Современный Chromium использует флаг v и требует экранирования дефиса. Правила логина сохранены. Настоящий input.checkValidity отвергает `bad!`, принимает допустимый логин, SyntaxError исчез.
- `NestingPage.tsx`: исправлен **recovery-дефект**. Отказ GET placements раньше отвергал весь Promise.all вместе с успешно загруженными metadata поверхности/лекал/проекта. После confirmed/reload статус восстанавливал placements, но surface оставалась null и холст показывал загрузку. Теперь placement promise сразу получает settled result: контекст сохраняется, неизвестный layout блокирует записи через refreshRequired, после синхронизации появляется авторитетная раскладка. UUID/confirmed/scopes, SQL lifecycle и predicates не переписывались.
- README, verification, этот отчёт и сводка обновлены. CSS/дизайн не менялись.

## Positive matrix без подмены ответов

Все шесть вариантов регистрируют пользователя через UI и входят в настоящую сессию. Создают лист 100×100 и валидный радиальный SVG с 12 вершинами, добавляют project copies. В конструкторе читают X вершины, изменяют его на 0.25 и сохраняют проектное лекало с длинным именем.

Ручное размещение и Undo/Redo/Undo выполняют реальные запросы. GET placements затем возвращает 0: radial preview начинается **без occupied**. POST preview возвращает preview_id и одно размещение; Optimize counter увеличивается на 1. Save содержит тот же preview_id, geometry/poses совпадают, counter больше не увеличивается. После reload HTTP placements совпадают с сохранёнными, история содержит один run.

SVG проверяется на surface + один placement path и координаты точной geometry. JSON содержит один placement с теми же poses и конечными координатами в пределах 0..100. Экспорты сохранены: проверялось содержимое, не только появление download.

Тема проверяется через html[data-theme]. Холст имеет ненулевой размер, кнопки доступны и кликаются, root horizontal overflow не превышает 2px. Панели открываются обычными «Свойства», «Лекала», «Панель», без force click. Скриншоты не заменяют оценку удобства.

На positive path нет pageerror/unhandled promise, 5xx и неожиданных console errors. Ожидаемый HTTP 400 на browser refresh **до входа без cookie** записан отдельно. Остальные ошибки не игнорируются. В stale-фазе отдельно ожидаются реальный промежуточный 401 после refresh второй вкладки и конечный 409; UUID повторов один.

## Negative / recovery: реальные и синтетические отказы

| Сценарий | Доказательство / источник |
|---|---|
| Preview loss → reload → status → exact save | route.fetch выполняет реальный POST, route.abort теряет ответ; outcome/preview_id настоящие |
| Pre-reservation 400 | Route меняет quantity на 501; **настоящий API** отвергает запрос до Reserve, pending не зависает |
| Unknown/retry | Контролируемая потеря запроса/статуса, повтор того же UUID |
| DELETE 404 до Reserve | Не трактуется как остановка будущего POST; закрытие ожидания preview предупреждает пользователя |
| Cancel Hybrid60 | Реальная работа, DELETE 204, без partial; zoom работает во время расчёта |
| Save loss после COMMIT | Реальный authenticated POST до 200, затем потеря ответа; status/replay без дублей |
| 401 → refresh → 409 retry | В старом сценарии 401/409 **синтетические**, refresh настоящий, исходный COMMIT реальный |
| Status 503 / placements abort | Детерминированная браузерная сеть, не выключение PostgreSQL; pending/confirmed сохранены |
| Confirmed после reload | Placements всё ещё недоступны, canvas/context видны, manual writes disabled. После восстановления видны 3 placements без дубля истории |
| Save без close waiting | При pending нет кнопок «Закрыть ожидание» |
| Настоящий second-tab stale | Вторая вкладка добавляет placement после preview первой. **Реальный 409 nesting_stale**, один UUID, чужая запись сохранена, Optimize не повторён |
| 500/36 | 501 экземпляр и 37 углов отклоняются до POST, validationPosts=0 |
| HTML pattern | Невалидный логин отклонён настоящим browser input, без SyntaxError |

В новой positive/stale матрице нет route.fulfill. SQL success, authorization, geometry и completed не подменяются.

## Отказы и итоговые прогоны

Диагностические отказы сохранены, не включены в успешную статистику:

- Pilot1: неверный новый selector «Сохранить» вместо «Создать».
- Pilot2/3: strict console выявила invalid HTML pattern и ожидаемый anonymous refresh 400.
- Full1: desktop и старый recovery прошли; mobile остановился на закрытых properties. Тесту добавлено обычное открытие панелей.
- Full2: test isVisible race сам закрывал панель во время загрузки. Исправлен на ожидание controls и aria-expanded toggle.
- Full3: проверка читала промежуточный 401 вместо stale-409; confirmed-reload нашёл настоящий контекстный дефект. Исправлены oracle и production load boundary.
- Full4: 7/7 PASS, 48.8s.
- **Full5 и Full6 одной финальной версии: 7/7 PASS за 49.0s и 7/7 PASS за 47.7s.** Дополнительно проверены canvas visible/manual disabled при продолжающемся отказе placements. Итого 14 финальных compound tests, retries=0, новые схемы/users/API в каждом запуске.

## Воспроизведение и артефакты

```sh
# backend, без .env:
CGO_ENABLED=0 go test -c -o /tmp/stage10-api.test ./internal/integration
# После фактического DB gate строго nesting_test:
NESTING_BROWSER_API=1 NESTING_TEST_DSN="$VALIDATED_TEST_DSN" \
 NESTING_E2E_READY_FILE="$READY_FILE" NESTING_E2E_ORIGIN=http://localhost:5186 \
 /tmp/stage10-api.test -test.run '^TestBrowserAcceptanceAPI$' -test.count=1 -test.v -test.timeout=20m
# frontend, значения из проверенного readiness конкретного PID:
NESTING_API_TARGET="$VERIFIED_API_URL" NESTING_E2E_PORT=5186 \
 NESTING_E2E_STATS="$VERIFIED_STATS_URL" E2E_CHROME=/usr/bin/google-chrome \
 npm run test:e2e -- --workers=1 --trace=off
```

Orchestrator сохранён в artifacts/run-browser.py. Он использует созданный этой задачей nesting-stage10-db, проверяет DB/PID/readiness и в finally останавливает только spawned API. Playwright убирает свой Vite/browser. Для повторения оператор отдельно подготавливает эту test-среду; неизвестный рабочий DSN использовать нельзя.

Trace выключен, чтобы не архивировать JWT/password network capture. Сохранены логи без секретов, topology JSON, screenshots, SVG/JSON синтетических fixtures и source backup с SHA-256:
`/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages8_10`.
Скриншоты/экспорты — browser-full4/5/6; runtime topology — full5-topology.json/full6-topology.json.

## Требует ручной UX-проверки владельцем

Расположение панелей и поверхности; удобство mobile drawer, ручного добавления, drag/zoom; размеры controls и фокус; контраст и читаемость тем; ощущение скорости; **точный пользовательский SVG зависания**. Скриншоты эту оценку не заменяют.

Firefox/WebKit и настоящий touch-device не проверялись (SKIP), только Chrome с разными viewport. Worst-case arbitrary 500/36 за 120s не обещается. Этапы 11+ не начинались.

## Финальный gate

Фактически выполненные проверки и Git snapshot ниже.


- `go test -count=1 ./...`, `go vet ./...`, `golangci-lint run ./...`: свежие PASS, 0 issues (stage10-go.log).
- `go test -race -count=1 -p=1 -timeout=20m ./internal/nesting ./internal/geometry/nfp ./internal/usecase/nesting ./internal/transport/http/handler`: PASS; native327.405s, nfp1.351s, usecase117.232s, handler1.016s. Race reports отсутствуют.
- Полный актуальный PostGIS suite и адресный SQL race (RadialPreviewSavedExactly/RealProcessLifecycle/NativeDeadlinePreservesSavedLayout): PASS, явный database gate, actual parent/children CGO race через совместимый loader.
- Frontend build/lint/72 unit: свежие PASS после production fixes; дополнительный финальный lint после последних E2E assertions — PASS.
- Два финальных полных E2E: 14/14 compound tests, 49.0s/47.7s, retries=0.
- Финальный read-only SQL audit: 7 users, 6 projects, 7 placements, 6 runs; save 6completed/1failed, preview7completed. Лишней истории нет; failed — настоящий stale.
- Helper TestBrowserAcceptanceAPI/TestRealProcessWorker SKIP в обычном parent — ожидаемо. Browser helper фактически запускался явно для E2E, process worker — в отдельных PID; не выдаём эти ветки за пропуск SQL проверки. Большой growth/perf эксперимент по умолчанию SKIP и отдельно не повторялся на этапе10.
- Исходный SVG/Firefox/WebKit/touch/worst-case 500: SKIP с границами выше.
- Собственный E2E DB остановлен, volume сохранён. Read-only Docker inspect подтвердил nesting_postgres и nesting-reliability-final running. Чужие API/PID не останавливали.

`git diff --check` PASS; все untracked файлы дополнительно проверены на whitespace. HEAD/ветка исходные. Обычный diff stat ниже не учитывает новые untracked файлы; полный список всех 22 файлов дан отдельно.

```text
 README.md                                            |  7 +++++++
 .../internal/nesting/nfp_exact_determinism_test.go   | 12 ++++++++----
 backend/internal/nesting/place_one_pattern.go        |  2 +-
 backend/internal/nesting/place_one_pattern_nfp.go    |  2 +-
 .../nesting/geometry_fixture_validation_test.go      | 20 ++++++++++++++++++--
 docs/verification.md                                 | 15 ++++++++++++++-
 frontend/e2e/nesting.spec.mjs                        |  9 ++++++++-
 frontend/playwright.config.mjs                       |  9 ++++++---
 frontend/src/pages/NestingPage.tsx                   | 19 ++++++++++++++++---
 frontend/src/pages/RegisterPage.tsx                  |  2 +-
 10 files changed, 80 insertions(+), 17 deletions(-)
```

```text
 M README.md
 M backend/internal/nesting/nfp_exact_determinism_test.go
 M backend/internal/nesting/place_one_pattern.go
 M backend/internal/nesting/place_one_pattern_nfp.go
 M backend/internal/usecase/nesting/geometry_fixture_validation_test.go
 M docs/verification.md
 M frontend/e2e/nesting.spec.mjs
 M frontend/playwright.config.mjs
 M frontend/src/pages/NestingPage.tsx
 M frontend/src/pages/RegisterPage.tsx
?? backend/internal/integration/bootstrap_preview_test.go
?? backend/internal/integration/browser_api_test.go
?? backend/internal/nesting/bootstrap.go
?? backend/internal/nesting/bootstrap_test.go
?? backend/internal/nesting/performance_experiment_test.go
?? docs/stage10_automated_e2e_acceptance_report.md
?? docs/stage8_candidate_bootstrap_report.md
?? docs/stage9_geometry_performance_report.md
?? docs/stages8_10_summary.md
?? frontend/e2e/bootstrap.spec.mjs
?? frontend/e2e/fixtures/radial-pattern.svg
?? scripts/summarize_targeted_performance.py
```
