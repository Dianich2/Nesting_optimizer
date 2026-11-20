# Этап 13 — BLOCKED

HEAD `21eecdbc2a5fa35ddb59028622db3274f8580cb6`, ветка `wip/geometry-editor-v2`. Рабочее дерево содержит собственные изменения этапа 11 и исследование/тесты этапа 12; полный финальный status приведён в сводке/отчёте 12.

Этап 13 **не начинался**: этап 12 не прошёл обязательный независимый PostGIS oracle. Минимальные валидные треугольники дают Go InteriorsIntersect=false, PostGIS=true и два точных рациональных proper crossings. Подробности, воспроизведение и варианты решения — в [stage12_geometry_invariants_report.md](stage12_geometry_invariants_report.md).

Условие ТЗ: начинать 13 только после успешных 11–12; при неразрешённом FAIL критического теста остановиться. Новые браузерные проверки не могут исправить или доказать отсутствие этого геометрического расхождения. Predicate/precision contract не менялся ради перехода дальше.

| Запланированный набор | Факт |
|---|---|
| Реальный React/API/PostGIS,5 алгоритмов,preview/exactsave/recovery | BLOCKED,не запускался |
| Viewports390/700/768/1366/1440 и обе темы | BLOCKED,новых прогонов нет |
| hasTouch/isMobile/tap/keyboard/focus/export failures/logout scope | BLOCKED,не реализовывались/не запускались |
| Два финальных полных Playwright прогона 13 | BLOCKED,0 |
| Firefox/WebKit | SKIP;доступность заново не исследовалась,ничего не скачивалось |
| UI/CSS изменения 13 | Нет |

Браузерные PASS этапа 10 не переиспользуются как результат 13. Новые E2E users/API/listeners/frontend для 13 не создавались. Frontend build/lint/72unit свежими запускались в 12 и прошли,но это не Playwright-приёмка 13. Нагрузка/скриншоты 13 не имитировались.

После отдельного решения по строгой топологии: восстановить зелёный geometry gate12,прогнать весь corpus по PostGIS,повторить полные проверки 12,затем выполнить 13 в собственных private schemas. Не подключаться к рабочей БД и не читать .env. Original user SVG по-прежнему отсутствует.

Будущая ручная приёмка владельца: реальные размеры холста/панелей,удобство добавления/drag/zoom,читаемость,ощущение скорости,настоящий touch-device и исходный SVG. Выводов об удобстве не делалось. Commit/push/merge/миграций/UI-правок нет.


## Финальный Git snapshot

`git diff --check` PASS. HEAD/ветка неизменны;никакихcommit/push/merge. Нижеобычныйstatдляtrackedфайлов(новыеfixtures/tests/reportsвнёмнеотображаются),затемполныйstatus.

```text
 README.md                                          |   4 +
 backend/internal/nesting/anchors.go                |  78 +++++--
 backend/internal/nesting/anchors_test.go           |  11 +-
 backend/internal/nesting/bootstrap.go              |  11 +-
 backend/internal/nesting/bootstrap_test.go         |  11 +-
 backend/internal/nesting/candidate_from_anchors.go |  56 +++--
 .../nesting/candidate_from_anchors_test.go         |  12 +-
 backend/internal/nesting/generate_candidates.go    |  63 +++---
 backend/internal/nesting/nfp_candidates.go         | 247 ++++++++-------------
 backend/internal/nesting/place_one_pattern.go      |  25 ++-
 backend/internal/nesting/place_one_pattern_nfp.go  |   5 +-
 docs/verification.md                               |  10 +
 12 files changed, 272 insertions(+), 261 deletions(-)
```

```text
 M README.md
 M backend/internal/nesting/anchors.go
 M backend/internal/nesting/anchors_test.go
 M backend/internal/nesting/bootstrap.go
 M backend/internal/nesting/bootstrap_test.go
 M backend/internal/nesting/candidate_from_anchors.go
 M backend/internal/nesting/candidate_from_anchors_test.go
 M backend/internal/nesting/generate_candidates.go
 M backend/internal/nesting/nfp_candidates.go
 M backend/internal/nesting/place_one_pattern.go
 M backend/internal/nesting/place_one_pattern_nfp.go
 M docs/verification.md
?? backend/internal/geometry/simplefeatures/contact_oracle_regression_test.go
?? backend/internal/geometry/simplefeatures/testdata/stage12_contact_counterexample.json
?? backend/internal/integration/candidate_budget_test.go
?? backend/internal/integration/geometry_oracle_test.go
?? backend/internal/nesting/candidate_budget.go
?? backend/internal/nesting/candidate_budget_regression_test.go
?? backend/internal/nesting/geometry_invariants_test.go
?? backend/internal/nesting/testdata/stage12_property12_counterexample.json
?? backend/internal/usecase/nesting/invariant_corpus_test.go
?? docs/stage11_candidate_budget_safety_report.md
?? docs/stage12_geometry_invariants_report.md
?? docs/stage13_automated_acceptance_report.md
?? docs/stages11_13_summary.md
```
