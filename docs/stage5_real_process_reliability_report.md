# Этап 5 — два OS-процесса API

База: ce28255259f389241a028ff54938326076978897, ветка wip/geometry-editor-v2. Перед этапом дерево чистое. Production-код не изменён. Изменения этапа: backend/internal/integration/real_process_test.go (harness) и project_resources_test.go (read-only gate current_database/current_user/postgis_version до мутаций), этот отчёт.

## План и фактическая топология

План: строгий DSN gate → отдельные PID/пулы и реальная auth → управляемые SQL-барьеры → конкурентные POST/status/cancel → kill/restart/SQL fault → ограниченная нагрузка → регрессии.

Родительский integration binary создаёт уникальную test_<timestamp> схему строго в nesting_test. Два дочерних integration binary независимо вызывают container.New и NewRouter; у каждого собственные JWT/session middleware, memory и sqlx pool max=8, worker=2/queue=4, GOMAXPROCS=2. Каждый слушает отдельный 127.0.0.1:случайный_порт. application_name включает test schema и PID. Пароль fixture-хеша задаётся только в test schema; настоящий /auth/login выдаёт Bearer для обоих API. Config.Load и пользовательский .env не читаются.

Тестовые барьеры Reserve/Optimize/UOW и diagnostics listener компилируются только в *_test.go; production cmd/api не содержит admin endpoints. Межпроцессная синхронизация — отдельные таблицы в test schema, не общая Go map/channel. Счётчик INSERT process_calls выполняется при каждом фактическом Optimize. После release работает настоящий Baseline, результат проходит обычную server validateResult/atomic persistence. Отдельный child smoke использует production optimizer registry.

## Безопасность

Read-only gate подтвердил current_database=nesting_test, current_user=postgres, PostGIS 3.4. PostgreSQL 16 / image postgis/postgis:16-3.4; Go 1.26.2 linux/amd64. Использован выделенный nesting-reliability-final. Рабочий nesting_postgres и его volume не затрагивались. Дочерние PID принадлежат текущему exec.Cmd; только они получают SIGTERM/Kill. PostgreSQL не останавливался. SQL network fault закрывает только сокеты, созданные test Dialer конкретного child, и запрещает только ему reconnect.

## PASS-матрица

- Два одновременных POST одного preview/save K: один Optimize, byte-equivalent JSON результата, один operation row; preview ready row один. Оба keep_existing проверены.
- 409 при изменении quantity не создаёт дополнительного Optimize и не меняет оригинальный K.
- GET/DELETE на B управляют работой A. Cancel даёт failed, повтор K не считает заново.
- Kill A во время preview, новый PID, replay и новый ключ. Lease ускоряется контролируемым SQL UPDATE только конкретного тестового K; реальное ожидание двух минут не заявляется.
- Graceful restart готового preview сохраняет preview_id; expiry даёт failed/preview_expired без Optimize. Применение точного preview через B не считает заново.
- Save cancel↔COMMIT на барьере: либо completed с целой историей, либо failure без частичной истории.
- Deadline перед COMMIT: 408 и отсутствие новой истории.
- Два разных K одной поверхности дают completed/stale; разные поверхности оба completed.
- Клиент закрывает TCP во время Optimize; сервер завершает сохранение, B видит completed, replay не дублирует историю.
- Ответ после реального COMMIT клиентом не читается; B status/replay дают один результат/историю.
- Kill сразу после Reserve до Optimize: вызовов Optimize ноль, lease приводит к failure.
- Недоступный SQL pool/status: 500, не ложный completed.
- Реальный разрыв собственных TCP-соединений к PostgreSQL перед commit: 500 и rollback истории; B продолжает работать; истечение lease разрешает оставшийся running. Это изоляция сети одного API, не остановка PostgreSQL.
- Во время заблокированного Optimize в pg_stat_activity нет idle-in-transaction у PID-specific application_name.

## Нагрузка и пределы измерений

10 свежих обычных матриц и 3 race-матрицы. В каждой: 160 одновременных дубликатов одного preview и 40 разных save ключей; native оптимизация в этих storm-сценариях намеренно маленькая. Это нагрузка на HTTP/auth/idempotency/backpressure, не benchmark тяжёлой геометрии.

Обычные 10 повторов: 1600 duplicate clients и 400 отдельных jobs. На каждый storm: Optimize=1, completed=65, nesting_pending503=95. Это один owner плюс не более 32 ожидающих на каждый из двух процессов.

GET status во время storm (10 замеров): min 1.101 мс, median 3.091 мс, max 12.811 мс. Cross-process cancel (20 замеров): min 84.224 мс, median 96.630 мс, max 101.850 мс. Это выборочные значения конкретной среды, не worst-case/SLA/p99. Во время замеров параллельно шли regression/race проверки; замеры характеризуют contention, не изолированную скорость.

40 разных jobs: одновременно 4 Optimize (2 на каждый API), 28 immediate overload, оставшиеся completed/stale согласно актуальности снимка. SQL-пулы после коротких операций возвращались к InUse=0; примеры Open=4, WaitCount≈689–709 на весь accumulated run после storm. Высокий WaitCount не скрывается: запросы Reserve/auth конкурируют за pool до admission ответа. Пример sampled goroutines=97, heap≈1.97–2.39 МБ на API под нагрузкой; это samples, не пиковый RSS. Лимит памяти GOMEMLIMIT 384 MiB на child/512 MiB parent — soft Go memory target, не hard cgroup cap. Общая нагрузка ограничена 160/40, deadline каждого HTTP 12 с, child 180 с, parent suite 180 с.

## Команды и результаты

- go test -count=1 ./...: PASS (без DSN SQL тесты SKIP; отдельный полный настоящий прогон ниже).
- go vet ./... и golangci-lint run ./...: PASS, 0 issues.
- go test -race -count=1 ./internal/usecase/nesting ./internal/transport/http/handler ./internal/nesting ./internal/geometry/nfp: PASS, реальный запуск, native nesting 323.743 с.
- CGO_ENABLED=0 go test -c -o /tmp/stage5-integration.test ./internal/integration; ./stage5.test -test.run '^TestRealProcessLifecycle$' -test.count=10 -test.v -test.timeout=180s: PASS.
- go test -race -c ... ./internal/integration; race binary -test.run '^TestRealProcessLifecycle$' -test.count=3 -test.v -test.timeout=180s: PASS, и parent, и child реально race-instrumented. Для libc тестового контейнера явно использованы /work/host-loader --library-path /work и NESTING_TEST_PROCESS_BINARY/NESTING_TEST_CHILD_LOADER/NESTING_TEST_CHILD_LIBRARY. Первоначальная проблема пути os.Executable под ELF loader исправлена в harness.
- Полный ./stage5.test -test.v -test.timeout=180s в выделенном контейнере с явным NESTING_TEST_DSN (пароль не печатается): результат ниже. Миграции применяются только в уникальной test schema.
- frontend npm run build, npm run lint, 72 unit tests: PASS свежего запуска. Frontend не менялся; Playwright на этапе 5 не обязателен и отдельно не запускался (SKIP).
- git diff --check: PASS.

## Ограничения и ручная приёмка

Тестовая задержка Opt/UOW/Reserve и ускоренный lease не заменяют latency реальной сложной геометрии — этап 6. Контейнер PostgreSQL целиком не выключался; реально закрыты только собственные SQL-сокеты одного child. Формальный proof всех interleavings невозможен: 10/3 повторов — измеренное покрытие, не отсутствие всех будущих гонок. Исходная пользовательская геометрия отсутствует.

Для воспроизведения: подготовить отдельную nesting_test с PostGIS в public; передать явный DSN; запустить из backend NESTING_TEST_DSN=... go test -count=1 -v ./internal/integration -run '^TestRealProcessLifecycle$' -timeout 180s. Проверить разные PID/loopback-порты и database gate в логе. Не направлять на nesting_optimizer. Production-изменений по итогам измерений не потребовалось.

Финальный raw log GET: n=10, min=1.100801 ms, median=3.090687 ms, max=12.811044 ms.

Финальный raw log cancel: n=20, min=84.224473 ms, median=96.630262 ms, max=101.850067 ms.

Финальный полный PostGIS-набор: PASS. SKIP только у служебного TestRealProcessWorker в родителе без child env; реальные SQL-сценарии и subprocess suite исполнены.

```text
 backend/internal/integration/project_resources_test.go | 10 +++++++---
 1 file changed, 7 insertions(+), 3 deletions(-)
```

```text
 M backend/internal/integration/project_resources_test.go
?? backend/internal/integration/real_process_test.go
?? docs/stage5_real_process_reliability_report.md
```
