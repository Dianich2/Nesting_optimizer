# Этап 2: координация жизненного цикла preview

База: `wip/geometry-editor-v2`, HEAD `8fa599b1264c7e84404a2d08b7f30856737559be`. Перед началом дерево чистое. Изменения локальные; commit/push/reset/clean не выполнялись. Рабочая БД не изменялась.

## Решение

Прежний mutex координировал только один handler. Теперь PostgreSQL атомарно резервирует `(project_surface_id, operation_key)` в отдельной таблице `nesting_preview_operations`. Владение поверхностью проверяется usecase; FK привязывает scope к проекту и владельцу. Два процесса получают один ownership token, второй ожидает готовый результат без вычислительного слота. Fingerprint использует существующую нормализацию углов. Порядок разных лекал остаётся значимым входом оптимизатора.

Переходы: unknown → running → completed/failed; причины failed включают отмену, timeout, исходную ошибку. Lease фиксирован deadline не более двух минут. Истечение обнаруживается статусным запросом или ограниченным фоновым проходом. Старый ключ не перезапускает вычисление. Store готового preview и CAS completed выполняются в одной короткой проектной транзакции. Отмена и completed конкурируют на одной строке операции; только один исход может победить. Геометрический расчёт не удерживает транзакцию/соединение.

В response операции хранится компактная ссылка (preview_id/expiry/метрики), геометрия восстанавливается из PostgreSQL TTL-кэша. TTL 15 минут и квота 10 готовых записей сохранены. Вытеснение/удаление даёт failed/preview_expired при чтении статуса. Компактные tombstone остаются до удаления поверхности, запрещая новое вычисление старого ключа. Ранее готовые keyed-preview читаются без нового Optimize.

## API

POST `.../nesting/preview` + canonical UUID Idempotency-Key — claim/replay. Одинаковые параметры возвращают тот же preview_id, конфликт — 409 conflict_error.
GET `.../nesting/operations/{key}?kind=preview` — status unknown/running/completed/failed; completed содержит прежний NestingResult с preview_id, placements, metrics и preview_expires_at. failed содержит code/message.
DELETE того же маршрута с `kind=preview` — общий cancellation request, 204 пока running (повтор до финализации тоже 204), 404 nesting_not_running для неизвестного или терминального ключа. По умолчанию kind=save; namespaces независимы даже при одинаковом UUID. Неизвестный kind — 400. Чужие ресурсы — 404.

Frontend хранит key/kind/owner/project/surface/request до POST. Неоднозначный ответ сохраняет pending, доступны проверка, повтор того же ключа и запрос отмены. Подтверждённый completed восстанавливает proposal и открывает панель сохранения. Save использует исходные параметры и preview_id. Отмена не считается завершённой по одному лишь 204.

## Доказательства

- Два независимых handler/usecase с отдельными registry/лимитерами и общей SQL-БД: параллельные POST, один Optimize, одинаковый preview_id, конфликт других параметров, replay через третий новый handler без Optimize.
- DELETE обрабатывается вторым независимым handler; наблюдение статуса — другим usecase. Проверены чужой владелец и разделение UUID save/preview.
- Lease и смерть исполнителя моделируются истечением записи; повтор ключа не пересчитывает.
- Пять конкурентных cancel/commit сценариев проверяют согласованность результата и наличия готовой записи.
- Cleanup SKIP LOCKED проверяется с удерживаемой транзакцией строки и повтором после COMMIT; разные поверхности и legacy replay проверены отдельно.
- Playwright теряет ответ фактически завершённого preview, обновляет вкладку, получает статус и сохраняет точную геометрию. Счётчик preview POST остаётся 1.

## Ограничения и ручная приёмка

Лимитер разных задач остаётся локальным: два процесса могут одновременно выполнять разные ключи. Дубли ждут с лимитом 32 на процесс и опрашивают каждые 100 мс; монитор исполнителя также 100 мс. Горутин без deadline не добавлено. Отмена earcut/simplefeatures кооперативная между вызовами. Отдельные sqlx-пулы/OS-процессы в тесте не поднимались: доказательство использует два независимых API handler/usecase и реальную SQL-синхронизацию.

Для ручной проверки запустить два тестовых API на одной nesting_test, отправить keyed POST первому и GET/DELETE с kind=preview второму; повторить после перезапуска первого. Проверить смену аккаунта/поверхности с pending, TTL/вытеснение, stale перед сохранением. Новая миграция проверена только в выделенной nesting_test; рабочую мигрировать отдельно после ревью.

## Файлы

| Файл | Назначение |
|---|---|
| `backend/internal/integration/preview_coordination_test.go` | Межэкземплярная отмена, lease, UUID namespaces, гонки COMMIT/cleanup, legacy и TTL. |
| `backend/migrations/20261120120000_preview_operation_lifecycle.sql` | Новая аддитивная таблица и индекс lease; Down удаляет только новую таблицу. |
| `README.md` | Контракт статуса/отмены, координация, совместимость и ограничения. |
| `backend/docs/docs.go` | Сгенерированная документация Swagger параметра kind. |
| `backend/docs/swagger.json` | Сгенерированная документация Swagger параметра kind. |
| `backend/docs/swagger.yaml` | Сгенерированная документация Swagger параметра kind. |
| `backend/internal/app/container/container.go` | Подключение отдельного репозитория preview-операций. |
| `backend/internal/integration/nesting_lifecycle_test.go` | Два независимых handler/usecase для параллельных POST и восстановления. |
| `backend/internal/integration/project_resources_test.go` | Подключение preview-coordinator к fixture и явный kind отмены. |
| `backend/internal/repository/postgres/nesting_operation_repository.go` | Общий SQL lease/CAS-механизм для двух раздельных таблиц; компактная ссылка на готовый preview. |
| `backend/internal/repository/postgres/nesting_preview_repository.go` | Ограниченная фоновая финализация истёкших lease. |
| `backend/internal/transport/http/errors/handler.go` | HTTP 404 для истёкшего/вытесненного preview. |
| `backend/internal/transport/http/handler/nesting.go` | Выбор типа статуса и отказ от локального singleflight. |
| `backend/internal/transport/http/handler/nesting_execution.go` | Раздельная durable-отмена по kind; удалены поля старого singleflight. |
| `backend/internal/transport/http/handler/nesting_preview_execution.go` | Удалён заменённый PostgreSQL локальный singleflight. |
| `backend/internal/usecase/nesting/operations.go` | Общая координация с callback атомарной финализации; поведение save сохранено. |
| `backend/internal/usecase/nesting/preview.go` | Атомарный Store/Finish, восстановление геометрии и публичного статуса. |
| `backend/internal/usecase/nesting/run.go` | Claim до расчёта; replay; совместимость с готовыми preview прежней версии. |
| `docs/verification.md` | Инструкция проверки двух экземпляров и браузерного восстановления. |
| `frontend/e2e/nesting.spec.mjs` | Потеря реального ответа preview, reload, статус, сохранение той же геометрии без второго POST. |
| `frontend/src/features/nesting/runNesting.ts` | Тип и валидация pending preview с изоляцией владельца/проекта/поверхности. |
| `frontend/src/pages/NestingPage.tsx` | Сохранение ключа до POST, восстановление после reload, повтор/отмена, исходные параметры при save. |
| `frontend/src/shared/api/errorMessages.ts` | Сообщение об истёкшем/вытесненном preview. |
| `frontend/tests/nesting-pending.test.mjs` | Валидация kind, чужой scope и запрет save-payload в preview. |

## Команды и результаты

| Команда | Результат |
|---|---|
| backend: `go test ./...` | PASS; интеграция без DSN пропускается, отдельный реальный прогон ниже |
| backend: `go vet ./...` | PASS |
| backend: `golangci-lint run ./...` | PASS, 0 issues |
| backend: `go test -race ./internal/usecase/nesting ./internal/transport/http/handler ./internal/nesting ./internal/geometry/nfp` | PASS; полный native nesting прогон 323.805 с, повтор без изменений использует кэш Go |
| `NESTING_TEST_DSN=postgres://postgres:***@127.0.0.1:5432/nesting_test?sslmode=disable`, integration binary `-test.v` | PASS, реальная PostgreSQL/PostGIS, отдельные схемы |
| integration binary, собранный `go test -race -c ./internal/integration`, `-test.run Preview -test.v` | PASS; два независимых handler, cancellation/lease, TTL, legacy, COMMIT/cleanup |
| frontend: `npm run build` | PASS |
| frontend: `npm run lint` | PASS |
| root: `node --test --test-isolation=none frontend/tests/*.test.mjs` | PASS, 33 теста |
| frontend: `NESTING_API_TARGET=http://127.0.0.1:18082 E2E_CHROME=/usr/bin/google-chrome npm run test:e2e -- --reporter=line` | PASS, отдельный API на nesting_test; интерфейс, preview recovery/reload и точное сохранение |
| `git diff --check` | PASS |

Интеграция запускается скомпилированными тестовыми бинарниками внутри `nesting-reliability-final`, cwd `/work/internal/integration`; миграции читаются из `/work/migrations`. Это тот же Go test suite, а не имитация SQL. Для race используется совместимый загрузчик libc; рабочий контейнер nesting_postgres не затрагивался.

`docs/stage2_preview_lifecycle_report.md` — этот отчёт. Полные git status/stat и сырые журналы сохранены вместе с резервной копией вне /tmp в каталоге stage2-preview-lifecycle.
