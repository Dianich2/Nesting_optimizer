# Nesting Optimizer — итог этапов 17–19

Ветка `wip/geometry-editor-v2`, HEAD `585614491453ebd6067d8fce89b0c9430115bbc0`.
Первоначальное дерево чистое; изменения14–16 сохранены. Commit/push/merge/reset/clean
не выполнялись. Рабочий.env и рабочая БД не использовались. Все SQL-записи относятся
к собственной лаборатории `nesting-1719-0afaf4c3`, базе `nesting_test`, private schemas.

## 1. Подтверждённые и исправленные дефекты

- **CoveredBy false positive:** деталь проникала в отверстие; минимальный случай9вершин,
  независимый rational oracle/PostGIS. Добавлен bounded strict coverage с сохранением
  касаний/rect fast path и прежних work/exact budgets.
- **Complexity →500:** реальные1024×1024vertex fixtures теперь дают400/validation_error
  во всех create/update/preview/save/replay ветках, без потери layout/history.
- **Потеря причины:** exact budget больше не превращается в InvalidPolygon; durable
  first-return/replay сохраняет errors.Is(ErrComplexity) без новой схемы/public error code.
- **Production defaults:** startup отказывается принимать dev/одинаковые/короткие
  JWT secrets, слабые DB defaults/SSL, неверные TTL/ports/worker/queue/bcrypt/origin.
  Дополнительно path '/' у origin отклоняется до запуска; CSRF comparison не ослаблен.
- **DSN/ошибки:** reserved credentials/IPv6 корректно экранируются; startup connection
  Error() не раскрывает пароль/DSN, Unwrap остаётся доступен внутреннему коду.
- **Body limit:** настоящий Content-Length>4MiB корректно413 вместо500.

Миграции, математика NFP, RNG, Defaults,2mindeadline, TTL, квоты и транзакционная
архитектура не менялись ради зелёных проверок. Production исправления только
по подтверждённым failure-before/pass-after; алгоритмического ускорения не добавлялось.

## 2. Проверки и результаты

| Проверка | Наблюдаемый результат |
|---|---|
| Go `-count=1 ./...`, vet, golangci-lint | PASS; последние окончательные логи в final-checks-v3 |
| Race обязательных5пакетов | geometry/simplefeatures,nfp,nesting,usecase/nesting,HTTPhandler PASS |
| Дополнительный race | config,postgres,HTTPerrors,geometry_error,SQL lifecycle/security/identity/transport PASS |
| PostGIS suite | PASS с явными563-case и154coverage corpora, всеWKB проверены |
| Геометрия |154cases×4variants;563native results/1016actual polygons/698pairs;server563positive/1589negative PASS |
| Differential fuzz |20s,185825executions,0oracle mismatches; execution count включает discarded inputs |
| Frontend | build/lint,72unit tests PASS |
| Финальные E2E |14/14+14/14, retries0,flaky0, API-race warnings0; один production source hash |
| Ранние E2E17 |дополнительно14/14+14/14; не подменяют2финальных |
| HTTP ramp |smoke/2/4/8/16,1vs2API,4→8→4burst PASS; штатные503 отдельно |
| Native S/M/L/XL |10success,4штатных120s deadline408; нет partial commits |
| Разнотипные+occupied |50фигур3типов/4угла,1старое размещение,NFP200за1.713s;все51WKB PASS |
| Положительная owner isolation |2реальных пользователя/2PID/одинаковыйUUID/save+preview+keep+replay;race PASS |
| Controlled chaos |Reserve/crash/precommit/lost COMMIT response/graceful/DB restart/network/logout/password/delete PASS |
| Full-queue read/cancel |реальные GETstatus+DELETE при заполненной очереди:PASS,5.87ms в контрольном сценарии |
| Soak +idle |600.85s+300.33s,12552business HTTP,2282Optimize,0unexpected5xx/network errors/OOM/deadlocks |
| Post-soak WKB |3041saved geometries/1521history PASS; terminal UUID/tombstones сохранены |
| Parser/graphs |null idle Samples/nearest-rank/expected503 regression PASS;SVG XML validated |

Default Go без DSN намеренно пропускает DB/opt-in helpers: это **SKIP**, не SQL PASS.
Они проверялись отдельно с own-DSN. В full SQL старые optional OperationalGrowth,
StrictTransition/StrictMatrix остаются SKIP; новый реальный soak и явные563/154корпуса
реально выполнены. BrowserAcceptanceAPI/RealProcessWorker как отдельные helpers
SKIP вне своих wrapper, но исполнялись в E2E/real-process tests. Полные inventories
есть в raw logs/test-counts.json. Firefox/WebKit не устанавливались/не выполнялись.

Неудачные pilots не скрыты: исправлены ошибки самого стенда (SQL field name,
короткий idle tx sample, invalid multi-login model,1API replay routing,
expected503 assertion, exit-vs-metrics race, idle null Samples, read-only corpus
export env). Их результаты не используются как успешный production benchmark.
Production/integrity дефектов в нагрузочных/chaos прогонах не обнаружено.

## 3. Фактически измеренная ёмкость

Hard caps: DB1CPU/1GiB; load parent+2API вместе2CPU/1536MiB/pids256/swapoff;
workers2+queue4 на каждый API. Машина32CPU/60GiB, фонload≈33.
Окна ограничены количеством запросов, временем,DB2GiB/WAL512MiB.

- Лёгкий расширенный трафик8клиентов:после исправлений91.72RPS/1API,
  92.42RPS/2API; p95≈32.5/31.2ms. Before92.21/93.18RPS. Это шум коротких
  30sокон, не доказанный speedup и не максимальная физическая ёмкость.
- На16клиентах174.90RPS,32штатных503,p95=53.59ms,p99=86.85ms.
- Mixed-all-five soak:20.89RPS,p95=815.58ms,p99=1936.43ms. Эти quantiles
  включают медленные Genetic/SA/Hybrid; сравнивать их с light traffic нельзя.
- M20/4угла:Baseline0.863s,NFP1.028s;GA/SA/Hybrid достигли120s.
  L50/8:NFP10.318s. XL100/16:NFP60.433s,Baseline408. Это конкретные
  синтетические входы, не обещание500деталей/120s. Отдельный500/36regression
  использует короткий caller deadline50ms и проверяет rollback/replay, не SLA.

После idle heap15.93/21.09→1.18/2.04MiB,goroutines17/16→12/12,FD16→13,
InUse0,poolOpen3. Наблюдавшийся heap peak59.59MiB/API. DB delta≈5.41MiB,
WAL≈17.87MiB заsoak,0long idle tx/deadlocks. За15min не видно удержания workers;
это не24h leak certification. RSS может оставаться зарезервированным runtime.

CPU profiles указывают на repeated Polygon.Validate/allocations:≈54%cumulative
CPU Validate и≈34%cumulative mallocgc уAPI1 (вложенные проценты не складывать).
Около486GiB TotalAlloc — cumulative churn, не RAM. Следующий обоснованный шаг:
снизить повторную validation/allocations на immutable geometry внутри расчёта,
с отдельным exact/PostGIS regression benchmark. Точность/лимиты менять не надо.
Global Redis/queue/распределённый limiter автоматически не добавлялись.

## 4. Что остаётся ручным/ограниченным

- Отсутствующий исходный пользовательский SVG зависания:только synthetic fixtures.
- Реальная UX/CAD-приёмка,физический телефон,touch/сложные авторские лекала.
- Firefox/WebKit при наличии executable;сейчас Chrome и touch emulation.
- Настоящий HTTPS/TLS/CA/рольБД/origin/secret entropy и эксплуатационный deployment.
  Readiness gate — конкретная защита от defaults, не сертификат production security.
- Длительные/распределённые нагрузки на выделенном железе и иной traffic model.
  Никакой выдуманный SLA/max concurrent users из локального теста не выводится.

## 5. Отчёты и evidence

- [Геометрия/deployment17](/home/dianich/projects/nesting_optimizer/docs/stage17_geometry_and_deployment_audit.md)
- [Mixed load18](/home/dianich/projects/nesting_optimizer/docs/stage18_mixed_production_like_load.md)
- [Chaos/soak/recovery19](/home/dianich/projects/nesting_optimizer/docs/stage19_chaos_soak_recovery.md)
- [Machine evidence root](/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages17_19)

Backup исходного HEAD вне/tmp:baseline-5856144.tar, SHA256
`738bd937bd83e4fbcaa948261644e668ec7142df005cb92790de46d44b5b3301`.
Финальные source checkpoint/manifest/patch лежат там же; credentials/JWT в public
artifacts не включаются. Cleanup-ledger фиксирует остановленные только own IDs,
сохранённый own volume; final readonly audit показывает0client connections,
0idle tx/ungranted locks/prepared transactions/deadlocks. Browser test port5188 закрыт.
Рабочая БД не была целью наших команд; активность чужих процессов не измерялась.

## 6. Изменённые файлы

Полный список относительно HEAD приведён ниже и дублируется в source manifest.
- [README.md](/home/dianich/projects/nesting_optimizer/README.md)
- [backend/internal/config/config.go](/home/dianich/projects/nesting_optimizer/backend/internal/config/config.go)
- [backend/internal/config/config_audit_test.go](/home/dianich/projects/nesting_optimizer/backend/internal/config/config_audit_test.go)
- [backend/internal/config/dsn_test.go](/home/dianich/projects/nesting_optimizer/backend/internal/config/dsn_test.go)
- [backend/internal/config/validate.go](/home/dianich/projects/nesting_optimizer/backend/internal/config/validate.go)
- [backend/internal/geometry/simplefeatures/engine.go](/home/dianich/projects/nesting_optimizer/backend/internal/geometry/simplefeatures/engine.go)
- [backend/internal/geometry/simplefeatures/stage17_audit_test.go](/home/dianich/projects/nesting_optimizer/backend/internal/geometry/simplefeatures/stage17_audit_test.go)
- [backend/internal/geometry/simplefeatures/stage17_coverage_test.go](/home/dianich/projects/nesting_optimizer/backend/internal/geometry/simplefeatures/stage17_coverage_test.go)
- [backend/internal/geometry/simplefeatures/stage17_fuzz_test.go](/home/dianich/projects/nesting_optimizer/backend/internal/geometry/simplefeatures/stage17_fuzz_test.go)
- [backend/internal/geometry/simplefeatures/strict_coverage.go](/home/dianich/projects/nesting_optimizer/backend/internal/geometry/simplefeatures/strict_coverage.go)
- [backend/internal/geometry/simplefeatures/strict_interiors.go](/home/dianich/projects/nesting_optimizer/backend/internal/geometry/simplefeatures/strict_interiors.go)
- [backend/internal/geometry/simplefeatures/testdata/stage17_coverage_counterexample.json](/home/dianich/projects/nesting_optimizer/backend/internal/geometry/simplefeatures/testdata/stage17_coverage_counterexample.json)
- [backend/internal/integration/browser_api_test.go](/home/dianich/projects/nesting_optimizer/backend/internal/integration/browser_api_test.go)
- [backend/internal/integration/project_resources_test.go](/home/dianich/projects/nesting_optimizer/backend/internal/integration/project_resources_test.go)
- [backend/internal/integration/real_process_test.go](/home/dianich/projects/nesting_optimizer/backend/internal/integration/real_process_test.go)
- [backend/internal/integration/stage17_audit_test.go](/home/dianich/projects/nesting_optimizer/backend/internal/integration/stage17_audit_test.go)
- [backend/internal/integration/stage17_security_test.go](/home/dianich/projects/nesting_optimizer/backend/internal/integration/stage17_security_test.go)
- [backend/internal/integration/stage18_load_test.go](/home/dianich/projects/nesting_optimizer/backend/internal/integration/stage18_load_test.go)
- [backend/internal/integration/stage19_recovery_test.go](/home/dianich/projects/nesting_optimizer/backend/internal/integration/stage19_recovery_test.go)
- [backend/internal/repository/postgres/connection_audit_test.go](/home/dianich/projects/nesting_optimizer/backend/internal/repository/postgres/connection_audit_test.go)
- [backend/internal/repository/postgres/postgres.go](/home/dianich/projects/nesting_optimizer/backend/internal/repository/postgres/postgres.go)
- [backend/internal/transport/http/errors/body_limit_test.go](/home/dianich/projects/nesting_optimizer/backend/internal/transport/http/errors/body_limit_test.go)
- [backend/internal/transport/http/errors/handler.go](/home/dianich/projects/nesting_optimizer/backend/internal/transport/http/errors/handler.go)
- [backend/internal/transport/http/handler/browser_auth_test.go](/home/dianich/projects/nesting_optimizer/backend/internal/transport/http/handler/browser_auth_test.go)
- [backend/internal/usecase/geometry_error/complexity.go](/home/dianich/projects/nesting_optimizer/backend/internal/usecase/geometry_error/complexity.go)
- [backend/internal/usecase/geometry_error/complexity_test.go](/home/dianich/projects/nesting_optimizer/backend/internal/usecase/geometry_error/complexity_test.go)
- [backend/internal/usecase/nesting/operations.go](/home/dianich/projects/nesting_optimizer/backend/internal/usecase/nesting/operations.go)
- [backend/internal/usecase/nesting/run.go](/home/dianich/projects/nesting_optimizer/backend/internal/usecase/nesting/run.go)
- [backend/internal/usecase/placement/create.go](/home/dianich/projects/nesting_optimizer/backend/internal/usecase/placement/create.go)
- [backend/internal/usecase/placement/update.go](/home/dianich/projects/nesting_optimizer/backend/internal/usecase/placement/update.go)
- [docs/stage17_geometry_and_deployment_audit.md](/home/dianich/projects/nesting_optimizer/docs/stage17_geometry_and_deployment_audit.md)
- [docs/stage18_mixed_production_like_load.md](/home/dianich/projects/nesting_optimizer/docs/stage18_mixed_production_like_load.md)
- [docs/stage19_chaos_soak_recovery.md](/home/dianich/projects/nesting_optimizer/docs/stage19_chaos_soak_recovery.md)
- [docs/stages17_19_summary.md](/home/dianich/projects/nesting_optimizer/docs/stages17_19_summary.md)
- [docs/verification.md](/home/dianich/projects/nesting_optimizer/docs/verification.md)
- [scripts/build_stages18_lab.py](/home/dianich/projects/nesting_optimizer/scripts/build_stages18_lab.py)
- [scripts/create_stages17_19_lab.py](/home/dianich/projects/nesting_optimizer/scripts/create_stages17_19_lab.py)
- [scripts/plot_lab_metrics.py](/home/dianich/projects/nesting_optimizer/scripts/plot_lab_metrics.py)
- [scripts/run_stage16_browser.py](/home/dianich/projects/nesting_optimizer/scripts/run_stage16_browser.py)
- [scripts/run_stages18_19_lab.py](/home/dianich/projects/nesting_optimizer/scripts/run_stages18_19_lab.py)
- [scripts/summarize_stages18_19.py](/home/dianich/projects/nesting_optimizer/scripts/summarize_stages18_19.py)
- [scripts/test_summarize_stages18_19.py](/home/dianich/projects/nesting_optimizer/scripts/test_summarize_stages18_19.py)

Финальная фиксация:790 PASS test/subtest events,83SKIP в default Go run
(DSN/opt-in отключены намеренно; соответствующие SQL/HTTP cases исполнены отдельно).
Генератор corpus/env pilot не перезаписывал read-only fixture. Свежие full SQL,
5mandatory race,additional4race,SQL race и independent-owner race PASS.
Дополнительный heterogeneous50 occupied profile PASS. Final cleanup ledger:
33own containers stopped,0client connections/idle tx/locks/prepared tx/deadlocks.
`git diff --check` PASS;42изменённых/новых файлов. Итог остановлен для проверки,
следующий этап/изменение алгоритмов не начинались.
