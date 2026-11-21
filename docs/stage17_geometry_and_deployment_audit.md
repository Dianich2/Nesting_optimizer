# Этап 17 — аудит геометрии и deployment readiness

Основа: `585614491453ebd6067d8fce89b0c9430115bbc0`, ветка `wip/geometry-editor-v2`. Исходное дерево чистое. Ранее завершённые этапы14–16 не откатывались. До исправлений сохранён полный `git archive HEAD` вне `/tmp`: baseline-5856144.tar, SHA256 `738bd937bd83e4fbcaa948261644e668ec7142df005cb92790de46d44b5b3301`.

## Подтверждённые дефекты и исправления

| Риск | Failure-before | Исправление / pass-after |
|---|---|---|
| R1 CoveredBy | Валидная деталь входит в треугольное отверстие: legacy CoveredBy=true, независимый rational oracle=false | Общий bounded strictCoveredBy, сохранены касания, rect fast path и интерфейс; минимизировано до9вершин, JSON сохранён; exact/Go/PostGIS согласованы |
| R2 complexity | Реальные валидные1024×1024vertex геометрии превышают work1m;10HTTPcreate/update/preview/save/replay возвращали500 | Существующий validation_error/geometry.too_long, HTTP400; old pose/history сохраняются, повтор UUID не вычисляется |
| R2 exact cause | После4096exact calls edges() заменял ErrComplexity на InvalidPolygon | Причина сохраняется; отдельная регрессия на валидном near-collinear triangle |
| R2 durable cause | HTTP400 уже корректен, но errors.Is теряется при первом durable-return и replay | Canonical persisted complexity rehydration сохраняет sentinel без новой схемы/public code; остальные validation errors не приобретают ложную причину |
| R3 production defaults | Clean child APP_ENV=production принимал dev-defaults | Validate после Load: distinct32-byte JWT secrets, explicit DB password, verify-full SSL, HTTPS origin, TTL/port/bcrypt/worker/queue validation; local/dev сохраняются |
| DSN | Reserved characters в пароле/IPv6/user меняли DSN или делали его невалидным | net/url + UserPassword + JoinHostPort, round-trip test |
| Startup error | Invalid port driver error раскрывал synthetic secret marker | Безопасное Error() соединения, cause доступна errors.Is; значения/DSN не выводятся |
| HTTP input | Настоящий loopback Content-Length4MiB+1 возвращал500 | Узкое исправление framework body-limit→413/validation_error; неизвестные реальные internal errors остаются500 |

## Геометрические основания

Для валидных regular-closed полигонов coverage нарушается, если часть exterior boundary внутреннего полигона лежит снаружи внешнего, либо внутренности входят в отверстие поверхности. Проверяются exact AABB, сертифицированные orientation signs, proper crossings, точки/интервалы контакта и strict overlap с outer holes. Координаты intersection не строятся и не округляются. Работа/точные операции используют прежний общий bounded strictState; границы бюджетов не увеличены. При исчерпании — error, не silent acceptance.

Независимый test-only inclusion oracle использует rational horizontal slabs по всем входным vertex/intersection Y и включение открытых X-интервалов. Production predicate не вызывается как собственный эталон. Матрица154cases×4variants=616: triangles/concave, holes, gap/penetration oneULP, collinear vertices, open/closed/reversed, очень малые/большие dyadic scales, negative coordinates и actual Transform0/30/45/90. Actual manual Transform дополнительно проверен через WKB и HTTPcreate/update; reject409 сохраняет SQL pose.

Fresh563-case corpus:29.751s, TotalAlloc19,189,921,144B cumulative (неRSS); server563positive/1589negative; PostGIS1016actualpolygons/698pairs. Metrics и poses **не изменились ни в одном** из563 результатов относительно принятого15. Полный WKB oracle/controls завершился, не только первые500.

Pair benchmarks: один CPU/RAM512MiB runner,3×100ms, validation у old/new. Triangle-in ~12.6→2.7µs, holes-touch ~21→7.4µs, holes-contained ~20→6.7µs; полные ns/B/allocs сохранены в logs. Одиночный corpus wall не выдаётся за статистический speedup. Нет нового O(N²) rational hot path: AABB+filtered predicates, точный fallback только при несертифицированных знаках/контактах, прежние work/exact caps.

## Deployment / security

Load проверен настоящими child processes с полностью очищенным env и cwd=tempdir, без чтения рабочего `.env`. Production/prod/staging strict; local/dev/development/test сохраняют development contract. Порты1..65535, access TTL1..1440min, refresh>=access и<=525600min; bcrypt4..31 local,10..16 nonlocal. Nonlocal origin HTTPS без path/userinfo/query/fragment; DB SSL verify-full. Ошибки называют поля, не значения. Это readiness по конкретным рискам, не общий сертификат безопасности deployment; TLS termination/CA/secret entropy/операционные права остаются администратору.

Реальный router/JWT/SQL: expired access, неверная подпись, чужие GET/PATCH/DELETE/status/preview/save/history, UUID kind/surface scoping, insecure Origin, production Secure/HttpOnly/SameSiteStrict cookie, no-store, logout→refresh401. Invalid JSON/quantity501/37angles отвергнуты. Existing password/account-deletion и session tests включены в fullSQL. Security fixtures используют bcrypt4 отдельно от throughput.

## Изоляция и окружение

Новая лаборатория `nesting-1719-0afaf4c3`, собственные Docker network/volume/container IDs, случайные test credentials только в приватных0600/env. Не используется ни nesting_postgres, ни nesting-reliability-final. Обязательная wrapper-проверка ID, active network, volume и loopback port bindings; затем actual nesting_test/usernesting_lab/PostGIS3.4/current_schema. Go helper подтверждает private schema и12migration tables до fixture INSERT.

DB CPU1/hard1GiB; native CPU1/hard512MiB; SQL CPU1/hard768MiB; race CPU1/hard1GiB. API E2E CPU1/hard768MiB с собственным container ID/hostPID. Swap off, pids bounded, дедлайны у каждого процесса. Нагрузка машины около32–35 на32CPU, RAMavailable≈49GiB; поэтому ступени нагрузки дальше будут консервативными. Изначальный own internal network не публиковал port: до fixture writes собственный DB был остановлен и пересоздан на own bridge с тем же own volume, старые ID сохранены. Другие сервисы не менялись.

## Gate 17 — PASS

Full Go/vet/lint и5-package race прошли; полный SQL + explicit coverage154/corpus563/security PASS; frontend build/lint/72units PASS; два полных capped race-API E2E14/14PASS каждый (retries=0); дополнительные SQL real-process и security race checks PASS. Owner-change barrier PASS, последний lint:0issues. SQL opt-in Transition/old Matrix по умолчанию SKIP; новые обязательные coverage/corpus реально выполнены, прежние strict regression default tests тоже. Начало18 — только после окончательного PASS.

Неудачные harness pilots не спрятаны: имя переменной original в новом test конфликтовало с polygon; -1angle по существующему контракту нормализуется, поэтому invalid test заменён реальными37angles; Fiber app.Test обрывает oversized body доHTTP, статус доказан настоящим TCPheader test. Исторические failed logs сохранены отдельно.

Никаких миграций, precision epsilon, изменений NFP/RNG/TTL/deadline/лимитера/SQL lifecycle для зелёного теста. Commit/push/merge/reset/clean не выполнялись. `git diff --check` PASS; полный список файлов/checkpoint будет в сводке.

## Финальное уточнение readiness

При дополнительном чтении выявлен accepted origin с завершающим `/`: реальный
browser-auth сравнивает Origin буквально, поэтому ошибочная настройка не должна
проходить startup gate. Новый clean-child case сначала FAIL (accepted), затем
Validate запрещает любой path; matrix и реальный HTTP foreign-origin test PASS.
Нормализация Origin/ослабление CSRF не вводились. Рабочий.env не читался.
Полный финальный Go/vet/lint/frontend gate повторён после этой поправки (см.summary).
Математика и load-runtime не менялись; helpers строят typed Config без config.Load.

Дополнительно20s bounded differential fuzzing CPU1/RAM512MiB:185825executions,
0расхождений coverage/interiors с независимыми rational slab oracles. Число
executions включает отброшенные invalid-length/degenerate inputs, это не число
проверенных валидных полигонов. Correlated random near-contact/dyadic/1ULP inputs,
не безлимитный fuzzing. Лог сохранён вне/tmp.
