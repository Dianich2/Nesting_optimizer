# Этап 18 — реальные HTTP-процессы и ограниченная смешанная нагрузка

Основа: `585614491453ebd6067d8fce89b0c9430115bbc0`, ветка
`wip/geometry-editor-v2`. Production-изменения перечислены в отчёте17;
на этапе18 меняются только тесты, runner и документация. Ни рабочая БД,
ни пользовательские контейнеры/тома не использовались.

## Протокол и пределы

Собственная лаборатория `nesting-1719-0afaf4c3`, actual database `nesting_test`,
user `nesting_lab`, PostGIS3.4. Перед записью — проверка идентичности контейнера,
сети, тома, loopback binding, БД/пользователя/PostGIS; private schema/12 таблиц.
Два настоящих дочерних Go API PID, независимые sqlx pools8/4, JWT и router.
Алгоритмы создаются `container.New` с production Defaults. Bcrypt4 используется
только в синтетических учётных записях вне горячего пути.

DB: CPU1, hard1GiB, swap off, pids128. Parent+два API: общий CPU2,
hard1536MiB, swap off, pids256, read-only root/source, tmpfs128MiB.
Child GOMAXPROCS2/GOMEMLIMIT384MiB; parent192MiB soft limit. DB2GiB/WAL delta512MiB
safety guards; отдельный wall deadline25min. Долгий idle-in-transaction (>5s),
OOM или неожиданный HTTP/network outcome прекращают повышение нагрузки.
Машина32CPU,60GiB RAM, фоновый load≈33: измерения не относятся к выделенному
production hardware. Think time150ms и конечное число iterations ограничивают
клиентов. Это инженерная модель трафика, не статистика реальной аудитории.

Чтение проекта/профиля/размещений/истории/status, ручные create/delete,
preview и точное сохранение, keep/replace, replay UUID, refresh,
create/edit/delete временного проекта и копий библиотечных поверхности/лекала.
Клиенты моделируют вкладки одной сессии; refresh согласован RW-lock на стороне
синтетических клиентов. Вторая учётная запись проверяет запрет доступа к чужому
проекту. Модель не имитирует тысячи независимых пользователей. При1API весь
бизнес-трафик/replay идёт только в него; второй PID обслуживает diagnostics.
При2API retry пересекает процессы. Локальные лимитеры остаются2workers+4queue.

Для прямого измерения queue wait используется только Go test overlay:
`scripts/build_stages18_lab.py` добавляет bounded atomic counters/32 histogram
buckets. Production limiter byte-identical HEAD, SHA256
`401ce533b8e9c1c9460a147d181158c7e33079416ab41b8eb2122ea080fcac99`.
Observer source и overlay JSON сохранены; есть небольшой observer overhead.
Unit regression с overlay PASS. Обычные gates проверяют также сборку без него.

## Native геометрия — завершено

| Профиль | Вход | Алгоритмы / исходы |
|---|---|---|
| S |3 прямоугольника,2угла | Baseline52.6ms, NFP33.9ms, Genetic830.2ms, SA520.9ms, Hybrid1390.6ms; все200 |
| M |20 вогнутых с отверстием,4угла | Baseline863.2ms, NFP1028.0ms; Genetic/SA/Hybrid408≈120s |
| L |50 таких деталей,8углов | Baseline17.081s, NFP10.318s;200 |
| XL |100 таких деталей,16углов | Baseline408≈120s, NFP60.433s/200 |

Все14 сценариев завершены;10 successful,4 штатных deadline408. Это **не**
успешный расчёт четырёх прерванных задач. После408 сохранённый layout/history
не изменились. Дедлайн2min, точность и Defaults не повышались/не снижались.
GA/SA/Hybrid на L/XL отдельно не разгонялись после подтверждённого120s на M.
Конкурентность тяжёлых заданий1; параллельно569 GET, p50=6.134ms,
p95=26.771ms,p99=43.126ms,max=66.123ms,0ошибок. Все10 сохранённых ответов,
включая superseded layouts:255 геометрий прошли independent actual-Go-WKB
PostGIS IsValid/CoveredBy/Relate и Go coverage/strict-interiors audit.

## Целостность и сбор данных

На каждой ступени — SQL snapshot audit и фактически сохранённые poses:
Go Transform→WKB→PostGIS. После серии дополнительно проверяются все completed
save-ответы, включая заменённые позже раскладки, и соответствие числа completed
save числу history. Preview operation может хранить только preview_id: это
тombstone/identity, его нельзя выдавать за полный сохранённый polygon payload.
Успешные save проверяются по фактической геометрии. Replay200 проверяется отдельно;
overload503 при save сохраняет terminal failure, повтор того же UUID тоже503.

Raw JSON: endpoint/outcome/count/latency; nearest-rank p50/p90/p95/p99/max и n
генерирует `scripts/summarize_stages18_19.py`. Метрики: heap/allocs/GC/goroutines,
FD/threads/RSS/CPU поPID, poolOpen/InUse/Idle/WaitCount, limiter counters/histogram;
cgroup cpu/memory/pids/peaks; database size/WAL/temp/deadlocks/locks/idle tx,
максимальный возраст active query/transaction. CPU поPID — process total,
**не** точная атрибуция CPU конкретному конкурентному Optimize. Optimize wall
сохранён в history.duration_ms. SQL tracing каждого запроса не включался.

## Контроль качества самого harness

Все pilots сохранены отдельно и не подменяют финальный benchmark:

- Неверное имя SQL поля status исправлено на state до реальной нагрузки.
- Монитор сначала останавливался на любом кратком idle-in-transaction; отделено
  состояние >5s, без ослабления запрета зависших транзакций.
- 16 логинов одного пользователя инвалидировали старые сессии по контракту
  login Upsert. Невалидный401-прогон отброшен; используется одна общая сессия
  с согласованным refresh. Проверка401 теперь останавливает основной поток.
- В прежнем pilot «1API» replay обслуживал API2. Он не используется как честный
 1vs2 benchmark. Финальная маршрутизация исправлена и проверена.
- На16клиентах exact-save получил корректный503. Harness ошибочно допускал только
 200/409; классификация исправлена, добавлена проверка terminal503 replay.
 Smoke/2/4/8 уже PASS, поэтому повторена провалившаяся ступень16 и последующие.

## Gate 18 — PASS

Smoke/2/4/8, повтор16,2API и burst PASS. Controlled before/after PASS.
Общий read-only audit11 нагрузочных схем PASS (в том числе остановленные pilots);
конкретные числа и финальные метрики приведены ниже. Перезапуск/chaos19 начинается
только после этого gate. Ни одного неожиданного5xx/network error, OOM, deadlock
или долгого idle tx. Штатные503/408 учитываются отдельно.

| Ступень | API / clients | n | Seconds | RPS | p50 / p95 / p99 ms | 503 |
|---|---:|---:|---:|---:|---:|---:|
|smoke|1/1|725|60.09|12.07|2.20/19.84/31.83|0|
|baseline|1/2|1980|90.15|21.96|2.36/20.23/33.88|0|
|step4|1/4|1403|30.15|46.54|3.28/28.27/45.89|0|
|step8|1/8|2755|30.13|91.43|3.87/37.13/56.08|0|
|step16|1/16|5273|30.15|174.90|3.92/53.59/86.85|32|
|twoAPI|2/8|2781|30.16|92.21|3.16/32.77/54.95|0|
|burst4|2/4|964|20.08|48.01|2.72/24.70/39.48|0|
|burst8|2/8|1894|20.14|94.03|3.19/32.86/48.22|0|
|burst4tail|2/4|954|20.09|47.49|2.94/30.07/42.91|0|

Таблица включает только PASS-окна: первые четыре ступени из final-ramp и
повторённую верхнюю серию high-verified с новой private schema. Предыдущая
step16 с ошибочной проверкой503 не включена в capacity-таблицу.

| Версия / активныхAPI | n | RPS | p95 / p99 ms | 503 |
|---|---:|---:|---:|---:|
|5856144 before/1|2773|92.21|34.33/56.02|1|
|5856144 before/2|2809|93.18|28.63/46.02|0|
|after17/1|2765|91.72|32.53/52.87|0|
|after17/2|2785|92.42|31.20/53.80|0|

Одинаковые input/caps и тестовый observer, окна30s,8клиентов, последовательный
before→after. Before production получен из immutable git archive в отдельном
/tmp tree; наложены только2 test files, рабочее дерево не откатывалось. Разница
RPS около процента является шумом этого эксперимента, не доказанным speedup.
Это S-входы и регулируемая частота клиентов, не максимальная физическая ёмкость.
На16клиентах получены штатные32ответа503; из этого нельзя вывести «максимум16
пользователей». Большая часть запросов — простые HTTP/SQL операции, не тяжёлый NFP.

Общий retained audit: **11схем,14436 completed records (включая preview identity),
12147 сохранённых placement geometries,6661history**, все PASS. Дополнительно
отдельный geometry-profile audit:10saved layouts/255actual geometries PASS.
Терминальные строки старых pilots не удалялись.

Полные endpoint/outcome CSV, p50/p90/p95/p99/max/n, time series и observer source
сохранены вне Git в test-artifacts. Их источник — реальные HTTP, не usecase loop.
В sample metrics поле SQL `active` старых runner означает число соединений БД,
а не только state=active; query/transaction ages учитываются отдельно. Последний
cgroup probe может гоняться со штатным завершением контейнера; фиксируется
unavailable_after_container_exit и отдельно проверяется exit/OOM, без сокрытия
ошибок живого процесса. Peak — наблюдавшийся cgroup peak до последнего sample.

Инструменты: Go1.26.2, golangci-lint2.13.2, Node24.22.2/npm11.19.1,
PostgreSQL16/PostGIS3.4. `git diff --check` PASS. Git HEAD/branch прежние;
новых production-файлов этап18 нет, схемы/миграции/рабочий.env не менялись.

## Machine-readable evidence

- [stage18-final-ramp](/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages17_19/stage18-final-ramp/summary.json)
- [stage18-high-verified](/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages17_19/stage18-high-verified/summary.json)
- [stage18-controlled-before](/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages17_19/stage18-controlled-before/summary.json)
- [stage18-controlled-after](/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages17_19/stage18-controlled-after/summary.json)
- [stage18-geometry](/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages17_19/stage18-geometry/summary.json)
- [stage18-all-retained-audit](/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages17_19/stage18-all-retained-audit/test.log)

## Дополнительные комбинации

`TestStage18IndependentOwners` под race:2пользователя/2процесса, одновременно
одинаковый UUID, save+preview+keep, cross-process replay, отсутствие чужих данных.
8сохранённых размещений/4history, independent WKB PASS. Mixed ramp измеряет вкладки
одной сессии, а этот отдельный сценарий — положительную независимость владельцев.

`TestStage18HeterogeneousOccupied`:50фигур трёх типов (15rect/20concave+hole/15triangle),
4угла, keep_existing=true и заранее созданное поHTTP препятствие. Native default
NFP Greedy:200 за1.713s, все50 добавлены, старое размещение сохранено; все51actual
placements и50новых вdurable result прошли Go/PostGIS WKB. UUID replay без второго
history. Это отдельный M/L-профиль, не часть15minsoak или before/after таблицы.
