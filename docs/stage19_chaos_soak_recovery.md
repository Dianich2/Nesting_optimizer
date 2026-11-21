# Этап 19 — controlled chaos и soak/recovery

Основа `585614491453ebd6067d8fce89b0c9430115bbc0`, ветка
`wip/geometry-editor-v2`. Выполнение разрешено только после успешного этапа18.
Gate18 PASS наблюдён до controlled chaos. Controlled chaos, полный непрерывный
soak/idle, два финальных E2E, PostGIS и race gates завершены PASS. Свои стенды
остановлены по recorded IDs; том/диагностика сохранены. Полный ledger приложен.

## Реализованные test-only сценарии

- `TestRealProcessLifecycle`: независимые PID/pool, concurrent same-key preview/save
  и keep/replace, fingerprint conflict, owner/surface scoping, cross-cancel,
  crash/expiry/restart, cancel-at-commit, stale snapshot/manual mutation,
  duplicate storm,2×(2running+4queued), ready restart/expiry, closed pool,
  client disconnect, lost response after actual COMMIT, reserve-before-Optimize
  crash, timeout-at-commit, SQL socket loss. Новые barrier cases:
  kill after Optimize/before actual commit и graceful shutdown с active+queued.
- `TestStage19PendingIdentityChanges`: logout, password change, project deletion,
  account deletion во время barrier-held accepted operation. Отдельные схемы,
  сохранность чужого пользователя и существующие account/session regressions.
- `TestStage19StatusTransportFaults`: loopback-only proxy delay/cancel, RST и
  dropped response; последующий status через другой API, без повторного commit.
  Cancel-before-Reserve и cancel-after-COMMIT явно проверяют404 и судьбу UUID.
- `TestStage19PGRecovery`: только guardian перезапускает записанный ID собственной
  PostgreSQL после marker, когда ready completed и barrier-held pending известны.
  Worker не имеет Docker socket. Проверка реального outage/recovery, replay
  completed, expiry pending, отсутствие partial history, новый UUID после восстановления.

Все hook/barrier/diagnostic существуют только в integration test binary;
production API, transactions, deadlines, local limiter и tombstones не изменяются.
Lease expiry в crash fixtures моделируется изменением времени конкретного private
fixture claim, без ожидания двух минут и без изменения production TTL.

## Soak protocol

Подготовлены10min смешанного реального HTTP с4клиентами/2API и production Defaults
всех5алгоритмов на малых S-входах, затем5min idle. Worker2+queue4 на каждый API;
общий cgroup CPU2/hard1536MiB, БД CPU1/hard1GiB. Bound request iterations,
25min wall deadline, DB2GiB/WAL512MiB guards. Периодические CPU/RSS/heap/allocs/GC,
goroutines/FD/threads/pool, queue histograms, SQL locks/idle tx/disk/WAL snapshots.
Терминальные UUID и tombstones не удаляются ради снижения памяти/размера БД.

После успешного soak выполнены2полных Playwright прогона retries0 на одной production
версии и финальная Go/static/SQL/race регрессия. Ручная UX-приёмка, настоящий
телефон и отсутствующий исходный пользовательский SVG автоматикой не заменяются.

## Результаты

Controlled chaos PASS: lifecycle (в том числе новые kill-before-commit и graceful),
pending logout/password/project-delete/account-delete, delayed/RST/drop status,
реальный own-PG restart. Дополнительный повтор lifecycle проверил настоящий GET
status и DELETE cancel при заполненной очереди. Рабочие службы/сеть не затрагивались.

Активная фаза финального soak:12552business HTTP,600.85s≈20.89RPS,
p95=815.578ms,p99=1936.433ms,0network errors/неожиданных5xx.
9506×200,1332×201,952×204;762×404 — намеренные owner-isolation проверки.
Смешивать эти quantiles с лёгким S-only benchmark нельзя: работают все5алгоритмов.
Счётчики физических Optimize отделены от HTTP (который включает replay).
CPU profile содержит algorithm labels; это sampled CPU, не точный CPU-clock на
каждую конкурентную задачу. Последние active snapshots:2282physical Optimize,
0janitor errors, pool/goroutine/heap time series сохранены. Окончательное сравнение с idle приведено ниже.

Два ранних soak pilots намеренно остановлены оператором для дооснащения
наблюдаемости (timestamps, затем per-algorithm CPU/elapsed). Они не считаются
полным soak PASS. Остановка только по recorded own container ID; evidence и
operator-stop.json сохранены. Полный новый непрерывный soak выполнен до конца.
Production алгоритмы/лимиты/TTL не менялись. Test-only maintenance повторяет
app.cleanupPreviews: тот же repository,1000row batches/5s/immediate+minute.


## Gate19 — PASS и измерения retention

Полный test901.96s: активная фаза600.85s и idle300.33s плюс audits/shutdown.
12552business HTTP,2282physical Optimize (replay не считается новым Optimize),
0transport errors/неожиданных5xx,0OOM/deadlocks/long idle tx. Post-soak audit:
3042completed records (включая preview identity),3041saved geometries/1521history,
все фактические WKB Go/PostGIS проверки PASS. Истёкшие/terminal UUID не удалялись.

| Метрика | API1 active→idle | API2 active→idle |
|---|---:|---:|
| Live heap MiB |15.93→1.18|21.09→2.04|
| RSS KiB |68332→33640|60916→32236|
| Goroutines |17→12|16→12|
| FD |16→13|16→13|
| PoolOpen / InUse |3/0→3/0|3/0→3/0|
| NumGC |30872→30876|30098→30103|
| Cleanup passes / errors |11/0→16/0|11/0→16/0|

Максимальный наблюдавшийся live heap59.59MiB/API, goroutines30,FD19. Cgroup working
set в конце idle40.1MiB при hard1536MiB. DB134492643→140161507bytes,
WAL delta18742613bytes, temp/deadlocks0. Эти размеры относятся только к этому
soak; финальный полный test lab больше из-за сохранённых private schemas.
Освобождение heap/RSS и снижение goroutines/FD не показывают удержания workers
в этом15min эксперименте. Это не доказательство отсутствия всех утечек за сутки.

Суммарный TotalAlloc двух API ≈486GiB за активную фазу — **оборот аллокаций**,
не потребление486GiB RAM. CPU profiles дали практическое направление улучшений:
в API1 Polygon.Validate ≈54.3% cumulative CPU, mallocgc≈34.2% cumulative;
проценты вложены и не складываются. Аналогичная картина у API2. Для уменьшения
allocation/validation overhead нужен отдельный точный benchmark и регрессия
strict-oracle, без ослабления математики, смены Defaults/RNG/deadline.

| Алгоритм | Optimize calls | Mean / max wall ms | Sampled CPU s (2API) |
|---|---:|---:|---:|
| Baseline |458|0.98 /4.35|0.45|
| NFP Greedy |456|1.38 /9.72|0.70|
| Genetic |456|785.90 /1840.32|206.62|
| Simulated Annealing |456|521.43 /1137.84|138.92|
| Hybrid |456|1429.31 /2889.91|348.70|

Это S-входы1–3детали/2угла; wall intervals могут пересекаться. CPU sampled
с100Hz labels, не точный CPU-clock каждого запроса. HTTP поалгоритмам включает
replay:1068Baseline и1064для каждого из остальных. Десять полных60sокон имеют
1171–1307samples; последнее неполное окно4samples хранится отдельно и не даёт
устойчивого tail percentile. См.latency-windows.csv и per-endpoint/outcome CSV.

Финальные2Playwright:14/14+14/14, retries0, flaky0, skipped0; API собран с-race,
логи без DATA RACE. Источник проверен одним production_sources_sha256 до/после
обоих прогонов.10viewport/theme cases (1440/1366/768/700/390 light/dark),
project copies/manual edits/recovery/exports, diagonal collision/SQL preservation,
owner/version/keyboard, все5native exact preview/save и Chrome touch emulation.
Lost preview/save,401refresh,unknown status,second-tab stale/reload проверяются
внутри этих сценариев. Firefox/WebKit и физический телефон не объявляются PASS.

Полный финальный SQL suite с actual563corpus и coverage154 PASS. Пять обязательных
race пакетов +SQL lifecycle/security/identity/transport race +config/postgres/
HTTPerrors/geometry_error race PASS. Отдельный owner/UUID race PASS.
Обычные Go/unit/static/frontend gates повторены после последней readiness-поправки.

### Ограничения диагностики

Перед завершением metrics collector может встретить уже завершившийся контейнер;
это отличено от ошибки живого процесса, exit/OOM проверены отдельно. В unit-race
pilot по ошибке передали output flag генератору coverage с read-only /tests;
write был отклонён, corpus не изменён. Затем reader/exporter env разделены,
те же математические tests PASS. Parser null Samples в idle исправлен с отдельной
unittest регрессией; nearest-rank и expected503 classification проверены.
Неудачные pilots/логи сохранены, не выданы за product PASS. Periodic cleanup
проверен test-only эквивалентом app loop с тем же repository/bounds; cfg.Load
проверяется отдельно clean-child tests, TLS deployment целиком не эмулировался.

## Артефакты и графики

- [summary.json](/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages17_19/stage19-soak-complete/summary.json)
- [endpoint-outcomes.csv](/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages17_19/stage19-soak-complete/endpoint-outcomes.csv)
- [latency-windows.csv](/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages17_19/stage19-soak-complete/latency-windows.csv)
- [http-by-algorithm.json](/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages17_19/stage19-soak-complete/http-by-algorithm.json)
- [optimizer-breakdown.json](/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages17_19/stage19-soak-complete/optimizer-breakdown.json)
- [cpu-api-11-tags.txt](/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages17_19/stage19-soak-complete/cpu-api-11-tags.txt)
- [cpu-api-18-tags.txt](/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages17_19/stage19-soak-complete/cpu-api-18-tags.txt)
- [cpu-api-11-top.txt](/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages17_19/stage19-soak-complete/cpu-api-11-top.txt)
- [cpu-api-18-top.txt](/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages17_19/stage19-soak-complete/cpu-api-18-top.txt)
- [heap.svg](/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages17_19/stage19-soak-complete/heap.svg)
- [goroutines.svg](/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages17_19/stage19-soak-complete/goroutines.svg)
- [fds.svg](/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages17_19/stage19-soak-complete/fds.svg)
- [memory.svg](/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages17_19/stage19-soak-complete/memory.svg)
- [cpu.svg](/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages17_19/stage19-soak-complete/cpu.svg)
- [wal.svg](/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages17_19/stage19-soak-complete/wal.svg)
- [latency-trend.svg](/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages17_19/stage19-soak-complete/latency-trend.svg)
- [final-source-version.json](/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages17_19/final-source-version.json)
- [final-e2e-gate.json](/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages17_19/final-e2e-gate.json)
- [final-sql-race-gate.json](/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages17_19/final-sql-race-gate.json)
- [final-db-readonly-audit.txt](/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages17_19/final-db-readonly-audit.txt)
- [cleanup-ledger.json](/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages17_19/cleanup-ledger.json)
