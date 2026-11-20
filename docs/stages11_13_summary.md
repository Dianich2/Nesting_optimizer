# Этапы 11–13 — остановка на геометрическом блокере

Основа: `21eecdbc2a5fa35ddb59028622db3274f8580cb6`, `wip/geometry-editor-v2`; исходное дерево чистое. Все изменения локальны, commit/push/merge/rebase и изменения рабочей БД не выполнялись.

| Этап | Статус | Результат |
|---|---|---|
| [11 — candidate budget](stage11_candidate_budget_safety_report.md) | **PASS** | Общие raw/storage200000,overflow-safe preflight,ctx в циклах,без partial;Go/race/PostGIS/Frontend прошли |
| [12 — geometry invariants](stage12_geometry_invariants_report.md) | **BLOCKED** | Go/server523 случая прошли,независимый PostGIS выявил strict-topology FAIL;сохранён минимальный красный regression |
| [13 — automated acceptance](stage13_automated_acceptance_report.md) | **BLOCKED** | Не начинался по предшествующему gate;новых Playwright прогонов 0 |

## Что реализовано и проверено в 11

Старый Baseline создавал всё произведение anchors×vertices×rotations без ctx/budget; direct NFP мог вернуть превышение лимита. На исходном коде bounded tests реально получили 200004/200017 кандидатов без ошибки. Огромный теоретический 603979776-element список не создавался.

Теперь `MaxCandidates=200000` ограничивает raw work и retained candidates на одну попытку размещения,включая все углы и fallback. Дубликаты расходуют work. Preflight делением до умножения предотвращает overflow;NFP использует один sink без больших промежуточных candidate lists. Все опасные add проверяют ctx/limit до роста. Переполнение — явный ErrComplexity/HTTP400,не усечённая раскладка. Прежние малые порядок/score/priority сохранены,bootstrap conservative5seeds/angle сохранён.

Прошли limit±1,MaxInt/zero,отмена внутри vertex/holes loops,нулевой partial,все 5 семейств,40HTTPfirst/replay budget failures на настоящем PostGIS в keep/replace. Старый placement99/history0/ready0 сохраняются,повторыневычисляются. Fresh Go/vet/lint/race4 пакетов,полный PostGIS/SQLrace и frontend build/lint/72unit — PASS. Принятый checkpoint11 отдельно сохранён с SHA-256.

## Почему нельзя принять 12 и начинать 13

В boundedmatrix523 случая/946formedpolygons покрыты 19 классами,5 метаморфнымивариантами,5 алгоритмами и 48seededproperty результатами. Все Go-предикаты,независимаяплощадь/metrics и servervalidateResult принялирезультаты;PostGIS на техже WKB отверг 501-йслучай:property/seed20261121/12/nfp_greedy,второеразмещение.

Минимизация дала два валидных треугольника. Go Relate=`FF2F01212`/InteriorsIntersect=false;PostGIS Relate=`212101212`/interiors=true. Независимая точная rational arithmetic подтверждает 2propercrossings. Различие координатоколо 8.88e−16;ST_Area пересечения округляетсядо 0,поэтому выводоснованнатопологии+точных signs,неплощади. Библиотека simplefeaturesv0.59.0snap-итблизкиеузлы;её precisionmodel расходитсяс ostrictoracle ТЗ.

Та же минимальная regression воспроизводится на исходном 21eecdb. Это ранее существующее поведение предиката,не изменение геометрии новым budget. Production Engine/RNG/NFP math не менялись. Ослабление oracle,tolerancewaiver,запреткасанияилипроизвольный epsilon-shift неприменялись.

**Текущий `go test -count=1 ./...` — FAIL**,потому что новый маленький детерминированный `TestStrictContactOracleRegression` намеренно оставлен красным по умолчанию. Он не скрыт opt-in и не объявлен PASS. Vet/lint и frontend72 — свежие PASS. Новыйобщий race/полный SQLgate12 послекритического FAIL непроводился;Playwright13 непроводился. Отчёт 12 разделяетэтистатусы.

## Дальнейшее решение

Нужен проверенный способ строгой near-contact топологии при сохранении допустимых касаний и ограниченной стоимости:robust/adaptive exact predicate с spatialfilter/budgets либо явное решение о precisioncontract. Отдельная финальная PostGIS-валидациятребуетоценки API/SQL/atomicity. Это не внедрялось наугад. Сначала устранить красный regression и пройти весь oracle/gate12,только затем 13.

Исходныйпользовательский SVG отсутствует. UX/настоящий touch/Firefox/WebKit/новыебраузерныесценарии 13 непроверялись. RuntimePASS этапов 8–10 незаменяютэтипропуски. Потерярабочихданныхневыявлена;рабочаяБДнеиспользовалась.

## Ревью и артефакты

Сохраненыисходнаябаза,принятый stage11snapshot,rawlogs,523-casecorpus,original/minimalcounterexamples,exactrationalproof,PostGISSQLproof ифинальный sourcebackup+SHA-256 вне/tmp:
`/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages11_13`.

Собственные runnercontainers остановлены,тестовых API/сокетовнеоставлено. Workingnesting_postgres ипрежнийвыделенный nesting-reliability-final оставлены running. Ихтома/настройкинеизменялись. Полный Gitstatus и diffstat приведенывотчёте 12;обычный stat неучитывает untrackedfixtures/reports.
