# Этапы 8–10 — завершены локально

Основа: `30aadbccedd56b5a0bb769e59403944b22fda75d`, ветка `wip/geometry-editor-v2`. На входе дерево было чистым. Commit/push/merge/rebase не выполнялись, миграции не добавлялись, рабочая БД и пользовательские контейнеры не перенастраивались.

| Этап | Итог | Основное доказательство |
|---|---|---|
| [8 — кандидаты](stage8_candidate_bootstrap_report.md) | PASS | 4 fail-before → 4 pass-after; все 5 алгоритмов размещают radial fixture; unit/server/HTTP/PostGIS/race |
| [9 — производительность](stage9_geometry_performance_report.md) | PASS, отрицательный эксперимент | Новый профиль, 3 пары/60 measurements, неизменное качество; сомнительный performance-патч снят |
| [10 — браузер](stage10_automated_e2e_acceptance_report.md) | PASS | Два полных свежих прогона одной финальной версии: 7/7 за 49.0s и 7/7 за 47.7s; настоящие React/API/PostGIS |

## Качество раскроя

Вогнутое радиальное лекало без bbox-corner прежде не получало первого положения в пустом листе. Общий bounded bootstrap добавляет до 5 проверяемых позиций на разрешённый угол **только после неуспеха прежних кандидатов**. Сохраняются их приоритет, реальные predicates, ctx и остаточный candidate budget. Все пять семейств используют общий путь; математическая полнота для произвольной геометрии не заявляется.

Проверены отверстия, препятствия, полностью занятая поверхность, неподходящая деталь, касание границ, повороты из списка, большие координаты, неосевой контур, отмена и корректность метрик. В fixed Hybrid fixture качество выросло с 8 до 9 размещений при прежних usedArea=4800/height=60. Изменившиеся строгие NFP golden counts обновлены только после воспроизводимого подтверждения и валидации геометрии; production RNG не менялся.

## Performance evidence

База этапа 9 включает **незакоммиченный принятый этап 8**, сохранённый полным source snapshot + SHA-256. На одинаковом harness и hard limits CPU2/RAM1GiB выполнены pilot, два профиля и три before/after пары: Baseline/NFP, 6/10/50/100 деталей, concave/holes/occupied. Все 60 основных измерений completed/valid, количество размещений, utilization и NFP work совпадают.

Эксперимент убирал повторную конверсию из Bounds после полной валидации. Differential proof дал 101 bitwise совпадение и сохранённое отклонение invalid inputs. Аллокаций стало меньше примерно на 1–2%, но wall не улучшился убедительно; у NFP100 медиана выросла примерно на 10% при пересекающихся диапазонах. **Эксперимент снят**, source/diff/proof сохранены отдельно. Ускорение итоговой версии не заявляется. Новые global cache/очередь/library/retention не внедрялись.

## Browser evidence и исправленные дефекты

Chrome 155 / Playwright 1.65.1, отдельный PostGIS `nesting_test`, private schema и настоящий authenticated API на каждый прогон. Матрица 1440/1366/390 × light/dark проходит SVG import, project copies/editor, manual Undo/Redo, radial preview без occupied, exact preview save с реальным Optimize counter, reload/history и содержимое SVG/JSON. Новая positive матрица не подменяет ответы.

Сохранён старый recovery-flow: потерянные preview/save ответы после реального сервера/COMMIT, unknown/retry одного UUID, pre-reservation реальный 400, отмена, синтетические 401/409 с настоящим refresh, временные status/placements failures. Дополнительно confirmed переживает reload при продолжающемся отказе placements, а реальная вторая вкладка вызывает настоящий nesting_stale без перезаписи чужого placement. Синтетические отказы явно отделены в отчёте.

Исправлены два production frontend дефекта: экранирование HTML pattern логина для флага v; сохранение успешно загруженного контекста поверхности при отказе initial placements read. Последний обнаружен именно новым confirmed-reload E2E: раньше восстановленные placements оставались без холста. До синхронизации ручные записи блокируются. CSS/дизайн не менялись.

Финальный read-only SQL-аудит последней E2E-схемы: 7 тестовых пользователей, 6 проектов, 7 placements, 6 runs; save operations 6 completed/1 failed (stale), preview operations 7 completed. Исходные диагностические FAIL сохранены, не засчитаны как успешные прогоны; автоматических browser retries нет.

## Проверки и оставшиеся границы

После каждого этапа свежие Go test count=1, vet, golangci-lint, race четырёх пакетов, full PostGIS и адресный SQL parent/child race — PASS. Frontend build/lint/72 unit — PASS. Этап 10 дополнительно: 14 успешных финальных compound browser tests. Обычные helper/explicit-heavy SKIP отделены от реально исполненных integration/measurement suites в отчётах. 30 000 SQL operations повторно не запускались ради галочки.

Блокеров для локального ревью нет. Ручная UX-приёмка, исходный пользовательский SVG, Firefox/WebKit/настоящий touch и worst-case произвольных 500 сложных деталей не проверены. Требуется оценка владельцем расположения панелей, размеров/управления и ощущения скорости. Local worker limits и бессрочные terminal UUID остаются существующим контрактом; новые архитектурные решения не принимались.

## Артефакты и ревью

Все 22 изменённых/новых файла перечислены в финальном Git snapshot отчёта 10. Сырые логи, CSV (36 bootstrap/60 performance rows), before/after CPU/mem profiles, rejected patch, screenshots/exports, topology и source backup + SHA-256 сохранены вне /tmp:
`/home/dianich/.codex/visualizations/2026/11/15/01a1dd80-a84e-78d1-8432-4588e1dbf5e8/stages8_10`.

Собственные API/Vite/browser и measurement/E2E DB завершены; рабочий nesting_postgres и прежний выделенный test DB продолжают работать. Собственный E2E volume не удалён. Изменения оставлены в рабочем дереве для независимого ревью. Этапы 11+ не начинались.
