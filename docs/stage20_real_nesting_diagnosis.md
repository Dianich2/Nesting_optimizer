# Этап20: фактическая диагностика автораскроя

Исходный HEAD `1d405973c77f892a4bca888d8aa1eef51d055a1c`, ветка `wip/geometry-editor-v2`, дерево в начале чистое. Изменения оставлены локально, без commit/push/merge. Архив исходного HEAD и SHA256 сохранены вне `/tmp` в каталоге артефактов `stages20_22` (см. сводку).

## Вывод и подтверждённые причины

Непроходимость двух простых деталей на тестовом стенде не воспроизведена: 74 случая preview/save прошли трижды (обычный, профилируемый и окончательный бинарники). Это **не доказательство исправления конкретного пользовательского проекта**: его геометрия и запрос отсутствуют. Исторический timeout и отсутствовавшие до ручного обновления таблицы — разные причины. Сложная занятая поверхность учитывает все препятствия.

Подтверждён отдельный дефект backend: принятая отмена во время ожидания SQL-блокировки возвращала HTTP500/internal_error. `lib/pq` может вернуть SQLSTATE57014 вместо `context.Canceled`; `runCoordinated` игнорировал причину отменённого дочернего контекста, а requestContext ещё был активен. Отдельный тест удерживает строку проекта, доказывает ожидание через `pg_blocking_pids`, принимает DELETE и получает ошибку. Test-only Go overlay старой строки воспроизвёл **500 для save и preview**. После минимального изменения `operations.go` причина координируемого контекста сохраняется до cleanup: HTTP409/nesting_cancelled, terminal failed, повтор UUID терминален, записей размещений/истории нет. Прямая SQL-проверка обоих вариантов выполнена10раз с race detector. COMMIT-wins recovery сохранён: успешный сохранённый результат по-прежнему восстанавливается deferred проверкой.

Startup до HTTP и janitor теперь read-only проверяет 12 обязательных таблиц/критические столбцы `current_schema()`, с пределом5с; несовместимость останавливает API с инструкцией обновления. Автоматических миграций нет. Проверка текущей и намеренно неполной **собственной** схемы PASS. Это минимальный guard, не аудит всех индексов/ограничений Goose. README обновлён.

UI показывает состояние, алгоритм, количество, прошедшие секунды без процентов, размещено/запрошено/неразмещено и время ответа. Успешный нулевой результат явно объясняется. Preview не меняет сохранённое, его сохранение использует точный preview_id без Optimize. Указаны120с расчёта и число учитываемых placements. Сообщения timeout/complexity уточнены. Старый блок результатов уже был доступен вне вкладки свойств; гипотеза «результат вообще не показывался вне этой вкладки» не подтверждена.

## Изоляция и метод

Рабочая `nesting_optimizer`, её `.env`, данные и службы не изменялись. Создан отдельный PostgreSQL16/PostGIS3.4, БД `nesting_test`, пользователь `nesting_lab`, собственные Docker ID/network/volume, loopback-порт, уникальные `test_<digits>` схемы. До записи wrapper и fixture проверяют actual database/user/PostGIS/схему с12таблицами и применёнными SQL миграциями. Только свои контейнеры останавливаются по проверенному ID. Новый тестовый Docker создан намеренно; утверждение «Docker вообще не менялся» было бы неверно.

DB: CPU1/RAM1GiB/swapoff/pids128. HTTP load parent+children: CPU2/RAM1536MiB/swapoff/pids256, read-only исходники, ограниченный tmpfs; wrapper контролирует WAL512MiB, DB2GiB, длительные idle transactions и deadline25мин. Browser API: CPU1/RAM768MiB. DSN/JWT/пароли не включены в отчёты.

## Полная матрица реального HTTP/JWT/Go/PostGIS

74 независимые операции: у каждой записаны UUID, preview_id, HTTP, SQL state, requested/placed/unplaced, latency, placements/history до/после в `stage20-matrix-final/cases.json` и `test.log`. Каждый POST повторён тем же UUID с точным сравнением ответа. Preview до apply не меняет placements/history; exact apply создаёт одну историю и replay её не дублирует. На конце проверены все сохранённые результаты, реальные Go Transform→WKB→PostGIS coverage/interior intersection. keep=false заменяет; keep=true сохраняет исходные IDs/позиции и добавляет новое.

| Случай | Алгоритм | keep | Тип | requested/placed/unplaced | HTTP/state | мс | placements до→после | history до→после |
|---|---|---|---|---|---|---:|---|---|
|empty|baseline|False|save|1/1/0|200/completed|48.739|0→1|0→1|
|empty|baseline|False|preview + exact apply|1/1/0|200/completed|8.750|0→1|0→1|
|empty|baseline|False|save|2/2/0|200/completed|9.277|0→2|0→1|
|empty|baseline|False|preview + exact apply|2/2/0|200/completed|8.630|0→2|0→1|
|empty|baseline|False|save|5/5/0|200/completed|17.184|0→5|0→1|
|empty|baseline|False|preview + exact apply|5/5/0|200/completed|12.642|0→5|0→1|
|concave|baseline|False|save|2/2/0|200/completed|8.735|0→2|0→1|
|concave|baseline|False|preview + exact apply|2/2/0|200/completed|15.091|0→2|0→1|
|hole-one|baseline|False|save|2/2/0|200/completed|38.667|1→2|0→1|
|hole-one|baseline|False|preview + exact apply|2/2/0|200/completed|23.340|1→2|0→1|
|hole-five|baseline|False|save|2/2/0|200/completed|35.918|5→2|0→1|
|hole-five|baseline|False|preview + exact apply|2/2/0|200/completed|14.727|5→2|0→1|
|complex-24|baseline|False|save|2/2/0|200/completed|25.018|24→2|0→1|
|complex-24|baseline|False|preview + exact apply|2/2/0|200/completed|16.801|24→2|0→1|
|empty|baseline|True|save|1/1/0|200/completed|8.700|0→1|0→1|
|empty|baseline|True|preview + exact apply|1/1/0|200/completed|35.064|0→1|0→1|
|empty|baseline|True|save|2/2/0|200/completed|48.106|0→2|0→1|
|empty|baseline|True|preview + exact apply|2/2/0|200/completed|16.400|0→2|0→1|
|empty|baseline|True|save|5/5/0|200/completed|15.840|0→5|0→1|
|empty|baseline|True|preview + exact apply|5/5/0|200/completed|16.725|0→5|0→1|
|concave|baseline|True|save|2/2/0|200/completed|11.781|0→2|0→1|
|concave|baseline|True|preview + exact apply|2/2/0|200/completed|23.993|0→2|0→1|
|hole-one|baseline|True|save|2/2/0|200/completed|18.780|1→3|0→1|
|hole-one|baseline|True|preview + exact apply|2/2/0|200/completed|18.874|1→3|0→1|
|hole-five|baseline|True|save|2/2/0|200/completed|27.914|5→7|0→1|
|hole-five|baseline|True|preview + exact apply|2/2/0|200/completed|43.647|5→7|0→1|
|complex-24|baseline|True|save|2/2/0|200/completed|69.099|24→26|0→1|
|complex-24|baseline|True|preview + exact apply|2/2/0|200/completed|64.279|24→26|0→1|
|empty|nfp_greedy|False|save|1/1/0|200/completed|12.375|0→1|0→1|
|empty|nfp_greedy|False|preview + exact apply|1/1/0|200/completed|11.096|0→1|0→1|
|empty|nfp_greedy|False|save|2/2/0|200/completed|18.402|0→2|0→1|
|empty|nfp_greedy|False|preview + exact apply|2/2/0|200/completed|20.308|0→2|0→1|
|empty|nfp_greedy|False|save|5/5/0|200/completed|19.550|0→5|0→1|
|empty|nfp_greedy|False|preview + exact apply|5/5/0|200/completed|20.837|0→5|0→1|
|concave|nfp_greedy|False|save|2/2/0|200/completed|12.055|0→2|0→1|
|concave|nfp_greedy|False|preview + exact apply|2/2/0|200/completed|10.444|0→2|0→1|
|hole-one|nfp_greedy|False|save|2/2/0|200/completed|15.398|1→2|0→1|
|hole-one|nfp_greedy|False|preview + exact apply|2/2/0|200/completed|11.839|1→2|0→1|
|hole-five|nfp_greedy|False|save|2/2/0|200/completed|12.311|5→2|0→1|
|hole-five|nfp_greedy|False|preview + exact apply|2/2/0|200/completed|13.325|5→2|0→1|
|complex-24|nfp_greedy|False|save|2/2/0|200/completed|25.789|24→2|0→1|
|complex-24|nfp_greedy|False|preview + exact apply|2/2/0|200/completed|26.481|24→2|0→1|
|empty|nfp_greedy|True|save|1/1/0|200/completed|11.610|0→1|0→1|
|empty|nfp_greedy|True|preview + exact apply|1/1/0|200/completed|8.688|0→1|0→1|
|empty|nfp_greedy|True|save|2/2/0|200/completed|20.636|0→2|0→1|
|empty|nfp_greedy|True|preview + exact apply|2/2/0|200/completed|10.339|0→2|0→1|
|empty|nfp_greedy|True|save|5/5/0|200/completed|18.276|0→5|0→1|
|empty|nfp_greedy|True|preview + exact apply|5/5/0|200/completed|20.246|0→5|0→1|
|concave|nfp_greedy|True|save|2/2/0|200/completed|9.403|0→2|0→1|
|concave|nfp_greedy|True|preview + exact apply|2/2/0|200/completed|11.449|0→2|0→1|
|hole-one|nfp_greedy|True|save|2/2/0|200/completed|15.893|1→3|0→1|
|hole-one|nfp_greedy|True|preview + exact apply|2/2/0|200/completed|20.497|1→3|0→1|
|hole-five|nfp_greedy|True|save|2/2/0|200/completed|20.762|5→7|0→1|
|hole-five|nfp_greedy|True|preview + exact apply|2/2/0|200/completed|19.080|5→7|0→1|
|complex-24|nfp_greedy|True|save|2/2/0|200/completed|70.354|24→26|0→1|
|complex-24|nfp_greedy|True|preview + exact apply|2/2/0|200/completed|53.790|24→26|0→1|
|occupied|baseline|True|save|2/2/0|200/completed|14.062|1→3|0→1|
|occupied|baseline|True|preview + exact apply|2/2/0|200/completed|17.523|1→3|0→1|
|occupied|nfp_greedy|True|save|2/2/0|200/completed|14.417|1→3|0→1|
|occupied|nfp_greedy|True|preview + exact apply|2/2/0|200/completed|16.944|1→3|0→1|
|occupied|genetic|True|save|2/2/0|200/completed|615.991|1→3|0→1|
|occupied|genetic|True|preview + exact apply|2/2/0|200/completed|614.304|1→3|0→1|
|occupied|simulated_annealing|True|save|2/2/0|200/completed|472.472|1→3|0→1|
|occupied|simulated_annealing|True|preview + exact apply|2/2/0|200/completed|466.743|1→3|0→1|
|occupied|hybrid|True|save|2/2/0|200/completed|1097.755|1→3|0→1|
|occupied|hybrid|True|preview + exact apply|2/2/0|200/completed|1094.011|1→3|0→1|
|full|baseline|True|save|1/0/1|200/completed|12.511|2→2|0→1|
|full|baseline|True|preview + exact apply|1/0/1|200/completed|7.925|2→2|0→1|
|too-small|baseline|True|save|1/0/1|200/completed|9.714|0→0|0→1|
|too-small|baseline|True|preview + exact apply|1/0/1|200/completed|14.439|0→0|0→1|
|full|baseline|True|save|2/0/2|200/completed|10.596|2→2|0→1|
|full|baseline|True|preview + exact apply|2/0/2|200/completed|13.045|2→2|0→1|
|too-small|baseline|True|save|2/0/2|200/completed|10.657|0→0|0→1|
|too-small|baseline|True|preview + exact apply|2/0/2|200/completed|10.652|0→0|0→1|

empty=300×200/деталь40×30/q1,2,5; occupied=одна исходная деталь/все5алгоритмов; concave=L; hole-one/hole-five=отверстие и1/5placements; complex-24=вогнутая600×400 с отверстием и24размещениями; full=полностью заполненная; too-small=10×10. Углы0/90, complex-24 четыре угла. Успех=0 при full/too-small ожидаем и не является500. Максимум первого прогона969,384мс, профилируемого1117,565мс, финального1097,755мс. Эти wall-time — характеристики данного стенда, не SLA и не оценка точности оптимума.

## Измерение фаз

Observer добавлен только Go overlay в test binary; производственные алгоритмы, RNG, геометрия, лимиты не менялись. `phases.json` содержит дельты накопительных таймеров. Начальный snapshot **включает** buildProblem; atomic commit включает CAS reload; внешняя commit revalidation измерена отдельно. Для preview строк фазы включают последующий exact apply, а HTTP latency — только исходный preview. Суммировать перекрывающиеся фазы как независимые нельзя.

Сложная поверхность24placements, прямой save двух деталей, отдельный профилируемый прогон:

| Алгоритм/keep | HTTP мс | snapshot incl.build | build | Optimize | validateResult | commit revalidation | atomic incl.CAS |
|---|---:|---:|---:|---:|---:|---:|---:|
|baseline/False|15.261|2.5243|0.0034|4.9026|0.0357|0.0304|4.8665|
|baseline/True|74.482|2.7682|0.0035|52.2163|0.1235|0.5560|5.6734|
|nfp_greedy/False|20.861|4.4281|0.0034|5.5514|0.0309|0.0222|8.4515|
|nfp_greedy/True|72.184|2.9333|0.0027|57.3740|0.1425|0.1450|8.2698|

Общий CPU профиль74операций: ValidatePolygon59,21% cumulative samples, CoveredBy25,54%, IsSimple28,71%; проценты вложенные и не складываются. Более тяжёлый отдельный случай50деталей3видов, отверстие,4угла, существующая деталь: HTTP200 за1859,147мс;51placement,1history; повтор и WKB всех50результатов PASS; GET status17,910мс во время расчёта. Его CPU профиль: Optimize77,53%, ValidatePolygon26,97%, InteriorsIntersect8,43%, CoveredBy6,74% cumulative. Это указывает на стоимость строгой геометрии, но не даёт оснований отключать validation или менять oracle. Длительный120с пользовательский timeout на её данных не воспроизведён.

## Ошибки, статусы и recovery

| Сценарий | HTTP/код/SQL | Доказательство |
|---|---|---|
| Простые2детали, занятая/пустая, Baseline/NFP |200/completed|74-case matrix + real React stage20 |
| Все5алгоритмов|200/completed, одна история|matrix occupied + stage16 browser exact save |
| Нет подходящей позиции|200/completed, placed0|full/too-small + browser zero |
| Отмена на SQL lock ДО|500/internal_error, failed|cancel-http-before overlay+SQL audit |
| Отмена на SQL lock ПОСЛЕ|409/nesting_cancelled, failed|cancel-http-after + SQL race10повторов |
| Нативный deadline|408/nesting_timeout, failed, старые placements на месте|TestNativeDeadlinePreservesSavedLayout; lifecycle timeout-at-commit |
| Queue full/duplicate waiters|503/nesting_overloaded или nesting_pending|TestRealProcessLifecycle duplicate-storm/local-capacity; mixed-load |
| Source changed/конкурентная запись|409/nesting_stale, без partial writes|snapshot/concurrent-commit SQL + browser stale-tab |
| Невалидный запрос/complexity|400/validation_error или complexity details|validation/candidate-budget/real-complexity SQL + browser |
| Server unavailable / unknown / running|нет подтверждённого outcome; старый UUID|node recovery + browser transport fault handling |
| Lost response after COMMIT|status completed, один run, без Optimize replay|real-process lost-response + browser save recovery |
| Cancel before COMMIT, restart, crash reserve/compute|terminal failure либо restored completed; без частичного save|real-process lifecycle + SQL cancellation/lease suites |
| Expired preview|404/preview_expired, старый UUID не пересчитывается|SQL preview retention/restart |
|401 refresh / повторная авторизация|штатный refresh; pending привязан кuser/project/surface|browser recovery + frontend unit |

Первый общий SQL прогон FAIL на конкурентной отмене (500); последующий isolated PASS не был принят за достаточное доказательство. Детерминированный lock-test подтвердил причину, исправленную выше. Следующий тест expired-replay в первом прогоне тоже упал после преждевременного выхода из capacity case с недочитанными запросами; на целиком завершённых прогонах не воспроизвёлся. Первые логи сохранены.

## Команды и scope

`python3 scripts/run_stages18_19_lab.py --metadata /tmp/nesting-2022-lab.json --binary /tmp/nesting2022-final-v2.test --test TestStage20RealMatrix --output <artifact>/stage20-matrix-final` — PASS74.
Аналогичный wrapper: `TestStage20SchemaGuard` PASS; `TestStage20CancelDuringSQLLockHTTP` PASS2; старый test-only overlay FAIL2 ожидаемо. `python3 /tmp/run-nesting2022-cancel.py` — PASS10повторов обоих вариантов под race.
`python3 scripts/run_stage16_browser.py --label stage20 --lab-metadata /tmp/nesting-2022-lab.json --api-binary /tmp/nesting20-profiled.test stage20.spec.mjs` — PASS1. Итоговые общие проверки в сводке.

Изменены: app startup/schema.go, nesting/operations.go, интеграционные тесты stage20/диагностика real_process, NestingPage/result messaging, README; тестовые wrappers/overlay. Миграций и изменений оптимизаторов нет. Лимиты backend120с/frontend130с сохранены. Фоновая очередь — отдельная будущая задача, сейчас не внедрена. Для ручной приёмки нужны исходная пользовательская геометрия, точный новый UUID/HTTP/status и выбранные параметры: текущие синтетические тесты не заменяют этого.
