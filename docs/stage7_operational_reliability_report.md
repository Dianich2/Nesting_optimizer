# Этап 7 — эксплуатационная устойчивость

## Основа, план и границы

HEAD `ce28255259f389241a028ff54938326076978897`, ветка `wip/geometry-editor-v2`. Перед этапом дерево содержало только сохранённые изменения этапов 5–6: общий integration helper, семь новых тестовых файлов, два отчёта, сводку и scripts. Полный исходный status сохранён в `stage7-start-status.log`. Commit/push/merge, новые миграции, изменение `.env`, рабочего PostgreSQL и алгоритмов не выполнялись.

План: аудит пула/лимитера/TTL → ограниченный pilot → повторные замеры роста → конкурирующая очистка и replay → read-only диагностика → свежий полный gate. Измеряется существующий контракт, а не вводится новый retention или глобальная очередь.

Изменения этапа:

- `backend/internal/integration/operational_growth_test.go`: 100/1000/5000 preview и save, реальные usecase/SQL/Baseline, квота, eviction, истёкшие leases, исторический replay, большой backlog и блокировка одной строки очистки.
- `backend/internal/integration/project_resources_test.go`: защита от пустого набора миграций и разрешения fixture-таблиц через public; regression отрицательного пути.
- `scripts/nesting_test_diagnostics.sql`: ограниченная read-only агрегированная диагностика только nesting_test/private schema.
- `scripts/summarize_operational_growth.py`: проверка полноты трёх запусков, точных состояний/счётчиков, CSV без UUID и геометрии.
- Этот отчёт, сводка, README и verification с командами воспроизведения.

## Безопасная среда и обнаруженный дефект harness

Go 1.26.2 linux/amd64, Node v24.22.2, golangci-lint 2.13.2. Фактическое SQL-подключение проверено: database=nesting_test, user=postgres, PostGIS 3.4 USE_GEOS=1 USE_PROJ=1 USE_STATS=1. На каждый тест отдельная `test_<timestamp>` схема; схемы сохранены для диагностики. Рабочий контейнер nesting_postgres не останавливался и не перенастраивался.

Отдельный parent-процесс в собственном контейнере: readonly root, CPU quota2/GOMAXPROCS2, RAM512 MiB, swap отключён, GOMEMLIMIT256 MiB, pids128, tmpfs work128MiB/tmp64MiB. PostgreSQL в нём не запускается; network namespace только выделенного тестового PostgreSQL. Контейнеры ограничены sleep1800/3600s, каждый тест timeout300s/context240s. Это ограничение памяти тестового клиента, не новый лимит памяти существующего тестового сервера PostgreSQL. Нагрузка выполнялась последовательно, без параллельного тяжёлого geometry benchmark.

Первый setup завершился FAIL: binary был запущен из `/`, относительный glob миграций вернул пустой список, старый helper считал его успехом. Пустая private schema позволила первому INSERT найти public.users в **тестовой** БД. INSERT завершился duplicate key; последующие fixture INSERT не выполнялись. Рабочая БД не затронута. Это реальный дефект изоляции тестовой инфраструктуры, а не SQL-гонка production API.

Исправление fail-closed: перед fixture write требуется непустой список миграций, current_schema должен совпасть с созданной, все 12 основных таблиц должны принадлежать private schema. PostGIS проверяется read-only вместо автоматической установки расширения. Новый TestMissingMigrationDirectoryFailsClosed и реальный запуск с неверным cwd подтверждают отказ до INSERT. При переносе миграций в финальный контейнер один дополнительный ошибочный путь также был безопасно отвергнут новой защитой; лог сохранён, затем каталог исправлен. Ошибочные setup не включены в успешную нагрузочную статистику.

## Что именно нагружалось

Каждый успешный запуск: 5000 PreviewWithKey + 5000 ExecuteWithKey, новые UUID, qty1 небольшого прямоугольного лекала, настоящий PostgreSQL/PostGIS и Baseline. 500 save используют точный preview; реальных Optimize ровно9500. Это **usecase/SQL нагрузка, не 10000 HTTP запросов**. Настоящие HTTP/JWT/два PID отдельно испытаны этапом5, с другими размерами очереди и характером нагрузки.

На 100/1000/5000: ready cache ровно10, save completed=n, preview completed=n и failed=n/50 (добавленные и восстановленные просроченные leases). Старый save после4999 новых записей возвращает исходный JSON без нового Optimize и без перезаписи последней раскладки; ранний вытесненный preview возвращает preview_expired без пересчёта.

Затем добавлены **синтетические** 5000 истёкших payload, клонированных из fixture, и2500 running claims с истёкшим lease. Они намеренно обходят per-user quota исключительно для проверки глобального backlog; это не5000 одновременно готовых API preview. Все10 обычных payload также истекают. В финальных повторах одна строка удерживается FOR UPDATE: PurgeExpired(batch1000) пропускает её, удаляет5009, завершается без ожидания снятия lock; после rollback удаляет последнюю. Всего5010 payload, ни одной terminal operation не удалено. Последний committed save с исходным preview_id после TTL всё ещё возвращает точный ответ; повтор preview становится preview_expired. Placements=1/history=5000 сохраняются. Pilot без удерживаемой строки также завершился успешно.

## Размеры и рост: pilot, байты

| Пар preview/save | Save total / indexes | Preview operations total / indexes | Ready / payload / saved_result |
|---|---:|---:|---:|
|100|155648 /16384|122880 /32768|10 /15976 /760|
|1000|1138688 /73728|704512 /122880|10 /15975 /760|
|5000|5472256 /303104|3121152 /417792|10 /15969 /760|
|5000 после backlog cleanup|5472256 /303104|4382720 /671744|0 /0 /0|

До backlog:5000 save completed,5000 preview completed,100 preview failed. После:5000 save completed,5000 preview completed,2600 preview failed, running=0. В pilot:141.61s на10000 основных операций; UC p50=10.431ms,p95=39.704ms,max=71.553ms. Очистка:7 вызовов,6.067–22.679ms каждый. Pool InUse0/Open1/WaitCount0 в checkpoints, goroutines3–4; это точки наблюдения, не доказательство нулевого ожидания при любой конкуренции и не peak goroutines. Повторы и финальные метрики приведены ниже и в operational-growth.csv.

При5000 до backlog JSON save:2842869 logical bytes/4084892 pg_column_size; preview metadata1567322/1848822 bytes на5100строк. Для этого малого fixture ≈817 stored JSON bytes/save и≈363/preview row; физический расход с индексами/heap/visibility/free-space ≈1094/save и612/preview row. Значения pg_total_relation_size включают индексы — не складывать их повторно. PostgreSQL не обязан возвращать место ОС после DELETE; relation preview cache после очистки всё ещё1564672байта, из них1540096индексов. MVCC, autovacuum, TOAST, размер ответов и fillfactor меняют прогноз.

Иллюстративная модель на этих измерениях:20preview+10save в день на365дней дают≈8.1MiB/year только этим двум operation tables;1000preview+1000save/day дают≈594MiB/year. Это линейная экстраполяция малого fixture, без истории nesting_runs, placements, WAL, backups, большого payload и служебных расходов. Это не capacity plan для500 деталей с4096вершинами. Число UUID и terminal rows не ограничено TTL15минут. Ограничение10ready/user с8MiB payload и16MiB saved_result даёт до≈240MiB/user до overhead; число пользователей глобально не ограничено. У save nesting_operations.response нет отдельного SQL16MiB check, в отличие от nesting_preview_operations.response. Tiny fixture не доказывает верхнюю границу размера произвольного save.

## Локальные workers и глобальная ёмкость

В production pool25open/10idle, lifetime30min,idle timeout10min. Эксперимент использует явно отдельный pool8/4, чтобы не выдавать лабораторный лимит за production. На этапе5 два настоящих API с workers2/queue4 дали **4 одновременно выполняющихся Optimize** и12 допущенных independent jobs суммарно, остальные28из40получили перегрузку. Connection waits в storm были ненулевые (см. отчёт5), хотя вычисление не удерживало write transaction. Перезапуск/status/replay также проверены настоящими PID на этапе5; здесь повторно измеряется рост SQL.

| Вариант | Достоинство | Цена / решение |
|---|---|---|
|Один API, текущие local2/queue4|Простой ограниченный deployment для курса|Зафиксировать ровно одну реплику; не глобальный лимит при scale-out|
|PG lease-based admission|Общая ёмкость без нового сервиса|Короткие CAS-транзакции, lease/heartbeat, fairness, crash recovery; не держать SQL/session advisory lock во время120s Optimize; нужна отдельная схема и согласованный контракт|
|Отдельная очередь/worker processes|Изоляция CPU/RAM, backpressure|Новая инфраструктура/deployment, durable ownership/cancel/recovery; текущие snapshot/idempotency всё равно нужны|
|Разделение workers по репликам + CPU/RAM quota|Полезно для фиксированного числа реплик на одном хосте|Не распределённая очередь и не строгая глобальная справедливость; сумма лимитов должна соответствовать capacity|

Автоматически глобальный semaphore/очередь не добавлялись. Для текущего coursework предпочтителен один API с local limiter; для нескольких реплик владелец должен выбрать global admission либо явно принять сумму локальных бюджетов и ресурсные квоты.

## Tombstones и безопасная retention-модель

Пока существует project_surface, старый UUID сохраняет исход без ограничения возраста. Logout не удаляет operation rows. Удаление scope (project_surface/project/account по существующим FK) каскадно удаляет связанные строки; права пользователя продолжают проверяться. Preview operation хранит компактный результат/терминальный исход, payload имеет отдельный TTL/квоту. После expiry completed metadata не означает доступность payload: публичный replay/status сообщает preview_expired и не запускает новый Optimize. Save хранит полный исторический ответ: после нового explicit replace старые placement IDs могут уже отсутствовать; replay подтверждает прежний commit, но не переписывает текущую раскладку. Поэтому авторитетный GET placements из этапа4 остаётся необходимым.

Удаление terminal v4 UUID через месяц превращает старый повтор в новый запрос: после DELETE сервер не отличает его от впервые увиденного случайного UUID. Это нарушает текущий контракт. Простая квота tombstones/Bloom filter проблему точно не решает. Возможный будущий bounded contract: версия протокола с конечным сроком, server-issued подписанным ключом с проверяемым возрастом (или отдельной принятой моделью времени для новых UUIDv7), явный expired-key отказ без вычисления, документированное окно replay. Legacy UUIDv4 всё равно требуют бессрочных tombstones либо breaking change. Отдельное удаление saved result тоже меняет ответ API/UI. Ни terminal DELETE, ни retention migration здесь не введены.

Существующая безопасная очистка: batch<=1000, SKIP LOCKED, индекс expiry, лимит фонового цикла5s, интервал1min. Expired running claims переводятся failed, истёкшие payload удаляются; recent ready и terminal rows не стираются. Фоновый цикл останавливается, когда **payload count<1000**, даже если обновил1000expired claims. Поэтому backlog только из claims>1000 может дренироваться по минутным тикам, а не весь за один тик; status также восстанавливает lease лениво. Это ограничение throughput, не повторное вычисление/потеря идемпотентности. В тесте измерен repository PurgeExpired непосредственно, не утверждается идентичное время всего фонового цикла.

## Наблюдаемость / воспроизведение

Добавлена агрегированная read-only диагностика: состояния/числа rows, bytes JSON/cache/indexes, pool/process samples в test harness. SQL script требует nesting_test и явно выбранную private schema; statement_timeout2s,lock_timeout200ms, BEGIN READ ONLY/ROLLBACK. Никаких UUID/JWT/паролей/приватных SVG в output. Это операторский SQL-инструмент, не новый публичный endpoint/Prometheus. Production per-operation telemetry и percentiles ожидания очереди не внедрялись; существующие startup/shutdown/cleanup-error logs сохраняются. Для production такой инструмент потребует отдельно согласованных прав/фильтров, не следует убирать test guard.

```sh
# Только после read-only gate фактического DSN и cwd backend/internal/integration:
NESTING_TEST_DSN="$VALIDATED_TEST_DSN" NESTING_OPERATIONAL_EXPERIMENT=1 \
  ./growth.test -test.run '^TestOperationalGrowth$' -test.count=1 -test.v -test.timeout=300s
# Диагностика: оператор явно выбирает созданную тестом схему через PGOPTIONS.
psql "$VALIDATED_TEST_DSN" -f scripts/nesting_test_diagnostics.sql
python3 scripts/summarize_operational_growth.py run1.log run2.log run3.log out.csv
```

Статическая CGO_ENABLED=0 сборка тестов копируется в bounded container; миграции находятся именно в /work/migrations, cwd=/work/internal/integration. DSN задавался явно, без чтения .env. Race parent/children используют CGO и совместимый ELF loader, точная команда в отчёте5 и финальном gate ниже.

## Не выполнено / ручная приёмка

Исходного зависавшего SVG нет. Нет доказательства worst-case arbitrary geometry, многомесячного production SLA или500largegeometryjobs под full concurrency. Большая growth-матрица не HTTP benchmark. Frontend не менялся, поэтому новый Playwright прогон на этапе7 SKIP по условию ТЗ; unit/build/lint обязательны и запускаются свежими. Архитектура глобальной очереди, новый retention и production telemetry требуют отдельного решения владельца.

Ручной checklist: проверить отчёты/сырые CSV; подтвердить одну либо несколько API-реплик и общий CPU/RAM budget; согласовать бессрочность idempotency/объём storage; отдельно испытать исходный SVG; проверить интерфейс preview/save/reload после потери ответа с актуальным frontend; сверить исторический ответ save с текущим GET placements после следующей раскладки. Не запускать fixture на рабочем DSN.

## Финальный gate

Результаты свежих повторов и проверок приведены ниже; незавершённые команды не считаются PASS.


### Повторы нагрузки — PASS

Три свежих процесса (pilot и два усиленных повтора), суммарно30000 основных usecase операций/28500 реальных Baseline Optimize,1500 точных preview save без повторного расчёта. Все логи полные PASS; strict summarizer проверил24 агрегированных table snapshots. Final runs проверяют SKIP LOCKED и последний save после TTL; pilot эти две дополнительные проверки не включал.

|Run|Общее время до final snapshot, s|UC p50 ms|UC p95 ms|UC max ms|Pool WaitCount checkpoint|Cleanup batch ms min–max|
|---|---:|---:|---:|---:|---:|---:|
|1|142.016|10.431|39.704|71.553|0|6.067–22.679|
|2|143.278|10.376|41.052|70.707|0|2.437–19.140|
|3|144.142|10.320|41.837|71.763|0|2.276–22.163|

В batch_ms финальных повторов входят7 вызовов до снятия блокировки, последний отдельный PurgeExpired после rollback проверен по n=1, но его длительность отдельно не замерялась. Не приписываем её таблице. Процентили рассчитаны по10000 последовательным UC latency каждого run, включаютpreview+save, не HTTP и не tail при concurrent users. Checkpoints goroutines3–7, итоговые3; отдельного непрерывного peak sampler в growth нет. Последовательные JSON/pool samples не являются production load SLA.

Финальный собственный cgroup: memory.peak63397888bytes (включая tmpfs/binary), memory.max536870912, cpu.max200000/100000, oom=oom_kill=0. Это parent container, не RSS всего PostgreSQL и не per-operation peak. В shadow тестовых схемах по10000 операций каждая; автоматическое удаление этих retained schemas не выполнялось.

Read-only диагностика без выбранной private schema также проверена: ожидаемый отказ до SELECT пользовательских таблиц. Read-only диагностика на сохранившейся pilot schema:5000 completed save,5000 completed/2600 failed preview,0ready. Возврат ROLLBACK подтверждён. Отрицательный путь summarizer с failed log отвергнут, синтаксис обоих Python scripts проверен. operational-growth.csv содержит24 aggregate table snapshots без payload/UUID.


### Свежие regression / integration

|Проверка|Результат|Подтверждение|
|---|---|---|
|go test -count=1 ./...|PASS|stage7-go.log; SQL без DSN здесь ожидаемо SKIP, отдельный реальный набор ниже|
|go vet ./...|PASS|команда завершилась0 после go test|
|golangci-lint run ./...|PASS|0 issues; exit0|
|npm run build / npm run lint|PASS|stage7-frontend.log, exit0|
|node --test --test-isolation=none frontend/tests/*.test.mjs|PASS|72 pass,0 fail,0 skipped|
|Актуальный полный PostGIS suite|PASS|stage7-integration.log,22 верхнеуровневых PASS; реальные private schemas, guard, все5 алгоритмов/keep modes, права/geometry/idempotency/stale/cancel/commit|
|PostGIS parent+children race|PASS|stage7-sql-race.log: TestRealProcessLifecycle, NativeDeadlinePreservesSavedLayout, MissingMigrationDirectoryFailsClosed; exit0|
|Growth3 свежих процесса|PASS|30000 UC операций, strict CSV24 snapshots, ожидаемые счётчики|
|Read-only SQL diagnostics / negative parser|PASS|ROLLBACK и отказ failed log|
|Первоначальный неверный cwd/setup|FAIL → исправлен|не скрывается; negative guards и правильный cwd затем проверены|
|Playwright|SKIP|frontend не менялся; условие задачи не требует нового E2E|
|Полная growth10000 под race|SKIP|проведена bounded ordinary матрица3раза; отдельная SQL-race матрица реальных PID/timeout выполнена, это разные проверки|
|Worst-case SVG / production SLA|SKIP|нет исходного SVG и согласованного worst-case resource proof|

Актуальный full integration запускается из /work/internal/integration, /work/migrations присутствует:

```sh
# DSN передан явно после отдельного read-only database gate; секреты опущены.
./stage7.test -test.count=1 -test.v -test.timeout=300s
# Race CGO binary, parent и spawned children используют один совместимый loader:
NESTING_TEST_PROCESS_BINARY=/work/internal/integration/stage7-race.test \
NESTING_TEST_CHILD_LOADER=/work/host-loader NESTING_TEST_CHILD_LIBRARY=/work \
 /work/host-loader --library-path /work ./stage7-race.test \
 -test.run '^Test(MissingMigrationDirectoryFailsClosed|NativeDeadlinePreservesSavedLayout|RealProcessLifecycle)$' \
 -test.count=1 -test.v -test.timeout=300s
```

Service-only TestRealProcessWorker SKIP в parent ожидаем: он исполняется в отдельных children. TestOperationalGrowth в полном ordinary integration SKIP по gate, а три явных growth запуска выше — PASS, не пропуск SQL проверки. Это разделение сохранено в логах.

После нагрузки остановлен только собственный nesting-stage7-final; собственные stage6-run/stage7-run завершились по sleep. Read-only Docker inspect подтвердил nesting_postgres и nesting-reliability-final running. Существующие пользовательские API/PID не останавливали. Изменения оставлены локально без commit/push/merge.


### Завершающий race и Git — PASS

`go test -race -count=1 -timeout=20m ./internal/usecase/nesting ./internal/transport/http/handler ./internal/nesting ./internal/geometry/nfp` завершился0: usecase98.679s, handler1.016s, native319.611s, nfp1.345s. Это новый запуск этапа7, не кэш/лог этапа6. Никаких race reports.

`git diff --check` чистый; дополнительно whitespace проверен у каждого untracked файла через diff --no-index --check, одна trailing space в отчёте6 удалена. HEAD/ветка остаются исходными. Обычный stat ниже не учитывает14untracked файлов: общий список17файлов указан полностью.

```text
 README.md                                          |  7 ++++
 .../internal/integration/project_resources_test.go | 39 +++++++++++++++++++---
 docs/verification.md                               | 28 ++++++++++++++++
 3 files changed, 70 insertions(+), 4 deletions(-)
```

```text
 M README.md
 M backend/internal/integration/project_resources_test.go
 M docs/verification.md
?? backend/internal/geometry/nfp/stage_cancellation_experiment_test.go
?? backend/internal/integration/native_deadline_test.go
?? backend/internal/integration/operational_growth_test.go
?? backend/internal/integration/real_process_test.go
?? backend/internal/nesting/cancellation_measurement_test.go
?? backend/internal/nesting/geometry_experiment_test.go
?? backend/internal/usecase/nesting/geometry_fixture_validation_test.go
?? docs/stage5_real_process_reliability_report.md
?? docs/stage6_geometry_reliability_report.md
?? docs/stage7_operational_reliability_report.md
?? docs/stages5_7_summary.md
?? scripts/nesting_test_diagnostics.sql
?? scripts/summarize_geometry_experiment.py
?? scripts/summarize_operational_growth.py
```

Все raw логи/CSV/profiles и финальный архив исходников с SHA-256 сохранены вне /tmp в `/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages5_7`. Итоговые17файлов оставлены в рабочем дереве для review.
