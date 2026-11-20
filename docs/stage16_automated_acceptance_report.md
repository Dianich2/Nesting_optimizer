# Этап 16 — автоматическая браузерная приёмка

Основа: `3d49cbbf1120fb9527ef6e295f9772ab58798579`, ветка `wip/geometry-editor-v2`. Этап начат после PASS gates 14 и 15. Предыдущие локальные изменения сохранены. Production frontend/CSS не менялись.

## Финальные браузерные прогоны — PASS

| Прогон | API | Passed | Retry / skip / flaky | Время Playwright |
|---|---|---:|---|---:|
| final1 | настоящий Go с `-race` | 14 | 0 / 0 / 0 | 78.569015 s |
| final2 | тот же race binary | 14 | 0 / 0 / 0 | 76.030953 s |

Оба прогона используют одну итоговую версию: SHA-256 manifest всех 14 изменённых/новых Go/JS/Python source files проверен после второго запуска. В каждом новая private schema, новые пользователи и собственные API/Vite/Chrome. В обоих API logs отсутствуют DATA RACE/race detected. Trace выключен, workers=1, retries=0. JSON reports подтверждают actual `passed`, единственную попытку и retry=0 для всех 28 результатов. Дополнительный normal-API pilot также прошёл 14/14, но не заменяет эти два финальных запуска.

## Изолированная среда

Создан собственный `nesting-stage16-db`, PostGIS 16-3.4: CPU2, hard RAM1GiB, swap выключен, pids128; порт опубликован только на 127.0.0.1. Перед каждым запуском read-only SQL проверяет фактические current_database/current_user/PostGIS. Go helper повторяет gate: строго `nesting_test`, user postgres, PostGIS3.4; затем новая `test_<timestamp>` schema и проверка 12 private tables до fixture writes. `.env` не читается.

API использует настоящий router, JWT, browser cookies/CSRF/refresh, repositories/transactions, native factories с production Defaults. GOMAXPROCS2, workers2/queue4, pool8/4, GOMEMLIMIT512MiB (soft), bcrypt4 только для тестовых регистраций. API и отдельный stats listener слушают собственные случайные 127.0.0.1 порты. Vite — localhost5188, strictPort, reuseExistingServer=false. Wrapper проверяет PID/database/schema/origin/readiness и останавливает только свой API PID. Playwright завершает свой Vite/Chrome.

Hard cap относится к тестовой DB. Браузерный functional run не выдаётся за cgroup-capped нагрузочный benchmark; native нагрузка этапа15 отдельно ограничена CPU/RAM/deadline.

Google Chrome155.0.8059.84, Playwright1.65.1, Node24.22.2, Go1.26.2. Firefox/WebKit executablePath проверены: движки не установлены, **SKIP**, ничего не скачивалось.

## Сценарии и доказательства

| Область | Фактическая проверка |
|---|---|
| Auth / owner isolation | UI register/login, настоящий refresh200 после reload; anonymous401; чужой JWT получает404 на project/placements/PATCH, SQL hash не меняется |
| Строгая геометрия | Диагональные треугольники: one-ULP overlap create/update409 именно `placement_overlap`; настоящее shared edge201; SQL counts/layout_hash точно сохраняются после отказов |
| Версия размещения | Реальный внешний PATCH с X-Placement-Version, затем запись старой UI-версии:409 `placement_version_conflict`, новая SQL pose сохраняется, alert виден |
| Keyboard / focus | Space выбирает SVG placement; Enter/Space открывают confirm; начальный фокус Cancel, Tab/ShiftTab внутри; Escape/Enter закрывают без удаления, фокус возвращается на trigger |
| Пять алгоритмов | Каждый запускается через UI: один Optimize, один preview_id, exact save с этим id; все poses/geometry равны preview; SQL placement1/history1; UUID replay не вычисляет и не дублирует history; reload подтверждает данные |
| Редактор / проект | Настоящий SVG import с contours/holes, изменение вершин, Undo/Redo, три независимые копии, удаление оригинала с сохранением экземпляра, fullscreen geometry editor / панели |
| Manual layout | Настоящие drag/rotate, координаты, ошибки overlap/outside/hole, Undo/Redo, удаление/отмена удаления |
| Recovery | Реальные preview/save responses теряются после выполнения; reload/status/UUID replay; unknown/running/failed/completed, confirmed при отказе GET placements; native Hybrid cancellation после Reserve, unknown/DELETE404 до Reserve |
| До Reserve | Валидный UI request намеренно изменён до quantity501; настоящий API отвечает400, pending очищается; frontend501/37angles блокируются до POST |
| Реальный stale | Вторая вкладка действительно добавляет placement; save первой получает настоящий409 `nesting_stale`, один UUID, чужая запись сохраняется, Optimize не повторяется |
| Экспорт | SVG/JSON проверяются по содержимому: paths, точные geometry coordinates, poses и конечность координат, а не только событие download |
| Responsive / темы | 390/700/768/1366/1440 × light/dark; canvas/controls, скрытые и раскрытые панели, root horizontal overflow ≤2px; реальные import/editor/preview/export на каждом варианте |
| Touch emulation | Chrome hasTouch/isMobile, реальные эмулированные touch pointer events, tap, drag, pan и zoom через touch-кнопки; сохранённая SQL pose проверена по доставленным координатам |

401/409/503 и abort в прежних recovery routes явно **синтетические**. Они не выдаются за SQL-гонки. Успешные ответы, SQL COMMIT, Optimize и история настоящие. Реальный second-tab stale проверяется отдельно.

## Тестовое инструментирование и диагностика

`browser_api_test.go` дополнен отдельными test-only SQL counts/layout_hash по surface_id: параметризованные SELECT с2s timeout, наружу выводятся только числа/хеш. Эндпоинта нет в production router. При ревью убран захват общей err-переменной: теперь query errors локальны каждому HTTP callback. Runtime DATA RACE до исправления не наблюдалась; это профилактическая корректировка тестового helper.

`bootstrap.spec.mjs` расширен на700/768; новый `stage16.spec.mjs` добавляет strict API, SQL preservation, versions, owner isolation, keyboard, все5 алгоритмов и touch. В Playwright включён screenshot only-on-failure. Wrapper `scripts/run_stage16_browser.py` сохраняет readiness/DB gate/API log/JSON report и browser artifacts; JWT/network traces не архивируются.

Touch pilots не скрыты:

1. Быстрый ручной поток Input.dispatchTouchEvent давал нестабильный следующий zoom/pan tap. Один успешный pilot не был принят: три повтора подтвердили нестабильность драйвера.
2. Использован native `Input.synthesizeScrollGesture`, speed200, preventFling, gestureSourceType=touch, с ожиданием фактического pointerup и render. Его доставленное перемещение включает touch slop; прежнее ожидание ровно10 world units оказалось неверным для этого драйвера.
3. Результат сравнивается с фактически доставленными pointer coordinates. SVG CTM projection отличается от viewport scale примерно на2.35e-6; UI-only bound5e-5 меньше пикселя. Строгие геометрические assertions/Oracle не используют этот допуск.
4. После исправления harness три независимых touch runs PASS2.4/2.5/2.4s, затем оба полных финальных прогона PASS. Production UI defect не подтверждён, его обработчики не переписывались.

Все неудачные pilots/raw logs/screenshots сохранены. Скриншоты390dark и700light визуально просмотрены: это проверка состояния стенда, не доказательство удобства интерфейса.

## Независимый SQL audit

После обоих финальных прогонов read-only audit с повторным DB gate:

| Схема | Users | Projects | Placements | History |
|---|---:|---:|---:|---:|
| final1 | 15 | 13 | 19 | 15 |
| final2 | 15 | 13 | 19 | 15 |

В каждой history: Baseline11, NFP1, GA1, SA1, Hybrid1. Save operations:15completed/1failed; failed — настоящий stale. Лишней истории/replay-дублей нет. SQL snapshots/hash и API poses также проверяются непосредственно в новых E2E сценариях.

## Воспроизведение

```sh
# Из backend; подготовить отдельный собственный nesting_test/PostGIS стенд.
go test -race -c -o /tmp/nesting-stage16-api-race.test ./internal/integration
# Из корня; wrapper работает только с собственным nesting-stage16-db,
# проверяет actual DB и создаёт private schema, не читает .env:
python3 scripts/run_stage16_browser.py --label review \
 --api-binary /tmp/nesting-stage16-api-race.test
```

Фактические финальные labels: final1/final2. Команда Playwright внутри wrapper: workers1, retries0, traceoff, list+JSON reporters. Непустые API race reports отдельно проверены как условие отказа при приёмке. Стенд после приёмки останавливается; рабочий DSN подставлять нельзя.

## Ограничения / ручная проверка

Firefox/WebKit отсутствуют — SKIP. Chrome touch emulation не является физическим телефоном; pinch gesture не объявляется проверенным (zoom проверен touch-кнопками). Исходный пользовательский SVG отсутствует. Владелец отдельно проверяет удобство панелей/drag/zoom, читаемость и контраст тем, ощущение скорости и свой SVG. Скриншоты это не заменяют. Произвольные500 сложных деталей за120s не обещаются.

Финальные backend/frontend gates, cleanup, HEAD/status и backup записаны также в `stages14_16_summary.md`. Commit/push/merge и изменения рабочей DB не выполнялись.

## Gate16 — PASS

Свежие full Go count1/vet/golangci-lint0issues, full PostGIS с606-case WKB corpus и адресный SQL/race parent/children прошли. Frontend build/lint/72units PASS. Native race пяти пакетов повторён на том же production-коде в14/15; новые browser/helper ветки фактически выполнены в обоих полных raceAPI E2E. Production исправлений после14 не было. `git diff --check` и whitespace новых файлов PASS; HEAD/ветка прежние. Четыре собственных runner/DB остановлены, API/stats/Vite listeners закрыты. Рабочий nesting_postgres и прежний nesting-reliability-final running; volumes сохранены. Точные cleanup/resource logs и20-file source backup — в stages14_16.
