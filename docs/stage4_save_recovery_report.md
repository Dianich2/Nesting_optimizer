# Этап 4 — восстановление сохраняемого автораскроя

## База и scope

HEAD перед началом: `31554515dad95d79af7c8372579f020cfda0eff8`, ветка `wip/geometry-editor-v2`. Дерево было чистым, `git diff --check` пустым, база совпала с ТЗ. Изменены только save recovery во frontend, его типы/тесты и этот отчёт. Backend, авторизация, миграции, рабочие .env и пользовательская БД не изменялись. Коммиты, push, merge, переключения ветки, reset/clean не выполнялись.

## Исходная проблема и план

До исправления `runNesting` пробрасывал любой 400/401/403/404/409/422/429, а `Workspace.run` безусловно вызывал clearOperation в catch. Воспроизведение: первый save реально коммитится, его ответ теряется; после reload повтор с тем же UUID получает 409. Старый catch удалял sessionStorage, не проверяя SQL-статус первого запроса. Отдельный E2E на старом коммите не запускался; причина установлена по исходникам, новый E2E воспроизводит эту последовательность и защищает инвариант сохранения UUID.

План перед правками: (1) выделить save-классификатор; (2) заменить пробрасывание всех 4xx на явный outcome; (3) при ошибке повтора выполнить один ограниченный GET; (4) объединить обработку результата POST/GET и загрузку authoritative placements; (5) сохранить подтверждение COMMIT при сбое синхронизации; (6) доказать поведение unit/E2E и повторить backend/PostGIS/race.

## Аудит сервера и переходы

| Место / ответ | Что известно | Новое действие |
|---|---|---|
| Первый POST: точный `400 validation_error`, message=`validation failed`, непустые details | RunNestingInput.Validate выполняется до Reserve; у нового UI-запроса нет предыдущего неразрешённого запроса с этим UUID | definitive_rejection: реальная ошибка, pending снимается |
| Другой 400 или та же ошибка на повторе | Нельзя доказать исход ранее отправленного запроса | Один GET status, сохранять ключ до terminal truth |
| Auth/проверка поверхности до Reserve, 401/403/404 | Отказ текущему POST не доказывает исход первого POST | GET status; его auth/network-отказ оставляет pending |
| 409 conflict_error повтора | Первый request fingerprint мог уже занимать ключ | GET первоначальной операции, исходный request не заменяется формой |
| Ошибка geometry/budget после Reserve | Owner обычно фиксирует failed после rollback | Подтверждение через SQL status |
| nesting_stale/cancelled/interrupted/timeout/overloaded | Прикладная ошибка может относиться к текущему вызову/ожидателю | GET status, без очистки только по HTTP-коду |
| nesting_pending | Первый worker может продолжать работу | GET status либо сохранение pending |
| TCP/Abort/Timeout, proxy без кода, неожиданный internal_error | Неопределённый исход | Первый POST: pending; повтор: одна попытка GET; нет нового POST/UUID |
| POST 200 либо GET completed + result | Сохранение подтверждено | Результат и confirmed=true фиксируются локально; GET placements заменяет локальный список |
| completed, затем GET placements не удался | COMMIT подтверждён, синхронизация нет | UUID + request + confirmed=true остаются в storage, показано «Раскрой сохранён…», доступны проверка/загрузка |
| GET failed с причиной | SQL подтвердил терминальный отказ | Показать причину, снять pending, размещения/историю не менять |
| GET running / unknown / недоступен | Исход не доказан; unknown может быть до Reserve | Сохранить UUID/request/scope, новый write блокируется |
| DELETE 204 / 404 | Только принятие отмены / отсутствие активной записи в данный момент | Не считать доказательством отмены или отсутствия COMMIT; проверять статус |

`runNesting` возвращает outcome `synchronized`, `confirmed`, `rejected` или `unknown`. Классификатор ошибок отдельно различает definitive_rejection/reconcile/uncertain. Статус SQL имеет приоритет над ошибкой нового POST. Никакой политики «закрыть ожидание» для save не добавлено.

## Авторизация и scope

`AuthProvider.authorizedRequest` создаёт новые Headers с новым Bearer, сохраняя исходные options.body, Idempotency-Key и signal. После 401 refresh и повтор используют тот же UUID и тело. При 401 самого refresh session сбрасывается; helper получает ошибку, но pending не очищает. При повторном входе того же owner старый ключ доступен; другой owner/project/surface отбрасываются readPendingNesting и isPendingSaveScope. Общая авторизация не менялась.

В E2E синтетический 401 повтора вызывает настоящий refresh браузерной сессии, затем повтор того же POST получает синтетический 409. Оба запроса проверяются на равенство исходных key/body. Отдельный сценарий реального отказа refresh=401 и повторного ручного входа не выполнялся; он остаётся в ручной приёмке и проверен анализом AuthProvider, unit-политикой сохранения и scope-валидатором.

Старый pending без kind остаётся save. Добавлено опциональное `confirmed: true` только для save: при сбое placements оно сохраняет факт подтверждённой записи после reload. Некорректное confirmed и сочетание confirmed+preview отклоняются валидатором. Новая запись UUID не создаётся на retry: берутся original current.request/current.key. Preview-классификатор и mayClosePreview не изменялись.

## Синхронизация и ресурсы

После подтверждения размещения не добавляются поверх старого списка. Единственный источник списка — GET placements; локальный список заменяется целиком. History revision инициирует чтение серверной истории, Undo/Redo сбрасываются лишь после подтверждения COMMIT. При неизвестном/failed исходе размещения и история сохраняются. При сбое синхронизации editing/new write блокируются и показан подтверждённый факт сохранения. Pending удаляется после успешной синхронизации либо достоверного отказа.

POST timeout — 130 секунд, одиночный status GET — 10 секунд, GET placements — 10 секунд. Новые polling/таймеры/автоматические POST не добавлены. Auth refresh остаётся существующим общим механизмом. В pending-save после reload доступен запрос серверной отмены, однако 204/404 не очищают ключ и не называются терминальным исходом.

## Изменённые файлы

| Файл | Назначение |
|---|---|
| frontend/src/features/nesting/saveRecovery.ts | Классификация ошибок save, outcome и guard owner/project/surface/kind |
| frontend/src/features/nesting/runNesting.ts | Явные outcomes, ограниченный reconciliation, единая синхронизация, confirmed-маркер в типе/валидаторе pending |
| frontend/src/pages/NestingPage.tsx | Удаление безусловного catch-clear; применение подтверждений и синхронизации; сохранение исходного request; подтверждённый COMMIT в storage; запрос отмены pending-save |
| frontend/tests/nesting-run.test.mjs | Сохранены положительные keep_existing/sync/timeout тесты, адаптированы к outcomes; прежние неверные утверждения о terminal 4xx заменены матрицей ниже |
| frontend/tests/save-recovery.test.mjs | Матрица ошибок и SQL-статусов с проверкой UUID/state, failure/confirmation, сохранение при network/auth/status errors |
| frontend/tests/nesting-pending.test.mjs | Совместимость confirmed-save после reload и запрет его смешения с preview |
| frontend/e2e/nesting.spec.mjs | Реальный pre-Reserve отказ, unknown+DELETE404+same UUID, COMMIT с потерей ответа, 401-refresh-409, GET/placements failures, точное preview save с потерей ответа |
| docs/stage4_save_recovery_report.md | Этот отчёт |

## Тесты и границы доказательства

Unit-тесты проверяют исходный pre-claim 400; retry 401/403/409; stale/cancelled/interrupted/timeout/overloaded/pending; proxy 500/502/503/504 и internal_error; TypeError/Abort/Timeout; SQL unknown/running/unavailable/completed/failed; confirmed при неудачной синхронизации; legacy kind и чужой scope. Считаются реальные POST-вызовы helper: они не повторяются автоматически. Сравниваются исходные request/key и удержание/освобождение состояния.

Playwright использует отдельного пользователя и API на **nesting_test**:

1. Реальный backend 400 до Reserve: при валидной форме перехватчик меняет только отправляемое тело на quantity=501, API возвращает настоящий отказ; pending освобождается.
2. Первый сетевой POST не достигает backend: GET unknown и DELETE404 не удаляют UUID; повтор отправляет тот же key/body.
3. Повтор реально создаёт COMMIT и одну историю; его ответ теряется. Синтетический временный GET503 удерживает pending для reload.
4. После reload перехватчик подменяет только ответы нового POST: 401, настоящий auth refresh, затем 409. База успешного сохранения не подменяется. Оба POST имеют исходные key/body. Следующий GET503 тоже синтетический, ключ сохраняется.
5. Настоящий SQL completed восстанавливается; одно прерывание GET placements проверяет confirmed=true и понятный режим «сохранено, синхронизация не удалась». Следующая проверка получает реальный список. Проверены три сохранённых размещения и одна история (две строки таблицы вместе с заголовком).
6. Предпросмотр восстанавливается после потери preview-ответа. Save с этим preview_id реально коммитится, его ответ теряется; GET status восстанавливает результат, геометрия совпадает точно, история/размещения не дублируются. Число preview POST остаётся 1; backend TestNestingSavesExactPreviewOnce отдельно проверяет счётчик Optimize.

Синтетические 401/409/503 — доказательство UI recovery при управляемых отказах, а не настоящая межпроцессная SQL-гонка. Успешные COMMIT, placements и история получаются настоящим backend/PostGIS.

## Команды и результаты

- `cd backend && go test ./...`: PASS; без DSN интеграция пропускается, реальный прогон ниже.
- `go vet ./...`: PASS.
- `golangci-lint run ./...`: PASS, 0 issues.
- `go test -race -count=1 ./internal/usecase/nesting ./internal/transport/http/handler ./internal/nesting ./internal/geometry/nfp`: PASS, **без кэша**, native nesting 322.907 с.
- `CGO_ENABLED=0 go test -c -o /tmp/stage4-integration.test ./internal/integration`, затем binary `-test.v` в `nesting-reliability-final`, cwd `/work/internal/integration`, миграции `/work/migrations`. Явный `NESTING_TEST_DSN=postgres://postgres:<REDACTED>@127.0.0.1:5432/nesting_test?sslmode=disable`: PASS полного PostGIS-набора; TestProjectResourcesAPI 43.35 с. Durable restart, rollback, stale, cancel/commit, namespace и точный preview не пропускались.
- `cd frontend && npm run build`: PASS.
- `npm run lint`: PASS.
- Из корня `node --test --test-isolation=none frontend/tests/*.test.mjs`: PASS, 72 теста.
- Playwright: `NESTING_API_TARGET=http://127.0.0.1:18082 E2E_CHROME=/usr/bin/google-chrome npm run test:e2e -- --reporter=line`: отдельный тестовый API/PostGIS; финальный результат ниже.
- `git diff --check`: PASS.

## Ручная приёмка / reset

1. На тестовой поверхности выполнить save, потерять ответ, обновить вкладку; проверить прежний UUID в pending и восстановить SQL completed. История одна.
2. На повторе вернуть 401/409: ключ не исчезает до терминального SQL-статуса. Повторный вход того же пользователя восстанавливает scope, другой аккаунт его не видит.
3. Оборвать status/placements: для неизвестного результата ключ остаётся; для confirmed видно «сохранено», повторная загрузка безопасна. Проверить оба keep_existing.
4. Запросить отмену до Reserve и на границе COMMIT: 404/204 не считать отменой; проверить статус. Новая операция доступна лишь после подтверждения/синхронизации или failed.
5. Сохранить точный preview с потерей ответа: проверить прежний preview_id и отсутствие нового Optimize.
6. Перезагрузка или сброс значений формы не должны удалять pending-save. Не использовать ручную очистку sessionStorage в качестве «reset» неизвестного сохранения. Для save нет «Закрыть ожидание».

## За рамками

Backend SQL-lifecycle, retention tombstones, OS cross-instance kill/restart, глобальный worker-limit, неделимые earcut/simplefeatures и исходное пользовательское зависание не изменялись. Реальный отказ refresh=401 с ручным повторным входом не был отдельным браузерным тестом. Ручная приёмка пользователем ещё требуется. Рабочая БД и её миграции не затрагивались.

## Финальный результат и Git

Playwright актуального дерева: PASS, 40.5 секунды. Проверка реального refresh между синтетическими 401/409 прошла; POST key/body совпали. `git diff --check` пустой. Новые файлы видны в status; обычный --stat считает tracked-файлы.

```text
 frontend/e2e/nesting.spec.mjs               | 141 +++++++++++++++++++++++++---
 frontend/src/features/nesting/runNesting.ts |  71 ++++++++++++--
 frontend/src/pages/NestingPage.tsx          | 120 ++++++++++++++---------
 frontend/tests/nesting-pending.test.mjs     |  10 ++
 frontend/tests/nesting-run.test.mjs         |  43 ++-------
 5 files changed, 288 insertions(+), 97 deletions(-)
```

```text
 M frontend/e2e/nesting.spec.mjs
 M frontend/src/features/nesting/runNesting.ts
 M frontend/src/pages/NestingPage.tsx
 M frontend/tests/nesting-pending.test.mjs
 M frontend/tests/nesting-run.test.mjs
?? docs/stage4_save_recovery_report.md
?? frontend/src/features/nesting/saveRecovery.ts
?? frontend/tests/save-recovery.test.mjs
```
