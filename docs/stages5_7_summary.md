# Этапы 5–7 — завершены локально

Основа: `ce28255259f389241a028ff54938326076978897`, `wip/geometry-editor-v2`. Production Go/Frontend и миграции не изменялись; добавлены тесты, измерители и документация. Рабочая БД/контейнеры не перезапускались. Commit/push/merge не выполнялись.

## Этап 5 — PASS

[Отчёт](stage5_real_process_reliability_report.md). Два настоящих PID, независимые пулы, реальный HTTP/JWT/PostGIS. 10 обычных и3 race-матрицы: в10обычных1600duplicate clients/400independent jobs. Один Optimize на одинаковый ключ; проверены save/preview, keep modes, cross-process cancel/status, SQL network loss, kill/restart, COMMIT/lost response/stale. Дополнительные свежие full SQL/race повторы после этапов6/7 также PASS. Выявлена суммарная конкуренция4 при worker2×2; local admission не глобален.

Изменения: real_process_test.go и database gate общего project_resources_test.go. Обязательные свежие Go/vet/lint/race, frontend build/lint/72unit и полный PostGIS — PASS. Playwright SKIP: frontend не менялся.

## Этап 6 — PASS с измеренными границами

[Отчёт](stage6_geometry_reliability_report.md). 282 paired measurements (47cases×3before/after), CPU/alloc profiles,75 optimizer cancellation и45 native-stage cancellation measurements,20 production-config server-validation cases. Actual native deadline сохраняет старую раскладку/историю. Regression/vet/lint/race, полный PostGIS и SQL race — свежие PASS.

Новые geometry/cancellation/server-validation/native-deadline tests, строгий CSV summarizer. Выявлено ограничение существующей эвристики radial_without_seed:0placements; математика не переписывалась. Для occupied50: NFP median3.544s, Baseline18.553s на данном fixture;500 ограниченно отменяется, успешный worst-case500/120s не доказан. Исходного SVG зависания нет. Earcut/UnionMany внутри одного вызова не прерываются, hard cancellation не обещается.

## Этап 7 — PASS с эксплуатационными решениями для владельца

[Отчёт](stage7_operational_reliability_report.md). Три свежих bounded growth запуска:30000 UC/SQL операций,28500 Optimize,1500точных preview saves. Это не HTTP throughput benchmark. Квота10, expiry/eviction без пересчёта,5010payload cleanup/run,7600preview terminal rows/run сохранены. В двух финальных повторах SKIP LOCKED при удерживаемой строке и committed save после preview TTL. UC p95 39.704–41.837ms для малого fixture. Strict CSV24aggregate snapshots, readonly diagnostic guard и negative paths PASS.

Исправлен fail-open integration helper: пустой glob миграций/неверный cwd теперь отказ до fixture writes, все12таблиц должны быть в private schema. Начальный setup FAIL разобран в отчёте; production defect/потеря данных не заявляются. Добавлены growth test, агрегированные SQL diagnostics и summarizer, обновлены README/verification.

Свежий финальный gate: Go full count1/vet/lint PASS, selected4packages race count1 PASS (native319.611s), full PostGIS22top-level PASS, реальный parent+children SQL race3tests PASS; frontend build/lint/72unit PASS. Playwright и full growth10000под race отдельно SKIP с объяснением в отчёте.

Владельцу нужно выбрать topology одного/нескольких API и общий ресурсный бюджет; согласовать бессрочность UUID/storage либо будущий versioned retention contract. Не внедрялись новый global limiter, удаление terminal UUID, retention migration, внешняя очередь или production telemetry service. Нужна отдельная проверка точного пользовательского SVG и ручная UI-приёмка recovery после потери ответа.

## Артефакты и ревью

Логи, CSV282geometry/24operational snapshots, CPU/mem profiles и source backup+SHA-256 сохранены вне /tmp:
`/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages5_7`.

Все17изменённых/новых файлов перечислены в финальном Git snapshot отчёта7. Обычный git diff --stat не включает untracked tests/reports/scripts; они включены в архив. Изменения оставлены для независимого локального review, автоматического продолжения новых этапов нет.
