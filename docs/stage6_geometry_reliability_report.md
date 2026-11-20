# Этап 6 — надёжность геометрии и измерения

HEAD: ce28255259f389241a028ff54938326076978897, wip/geometry-editor-v2. Перед этапом были только собственные изменения этапа 5: общий test DB gate, real_process_test.go, stage5 report и сводка. Они сохранены. Production-алгоритмы, параметры 500/36/2 минуты и SQL-координация не менялись.

## План / изменения

План: карта native вызовов → положительные/отрицательные fixtures → pilot → три paired before/after повтора → отдельная отмена/pprof → server validation и native SQL timeout → полный gate перед этапом 7.

Новые файлы этапа: internal/nesting/geometry_experiment_test.go (47 cases/metrics), cancellation_measurement_test.go (пять алгоритмов, pprof, точный rotation barrier), internal/geometry/nfp/stage_cancellation_experiment_test.go (реальные stages и budget edges), internal/usecase/nesting/geometry_fixture_validation_test.go (20 production-config cases + queue deadline), internal/integration/native_deadline_test.go (native 500/36 deadline, old layout/history/replay), scripts/summarize_geometry_experiment.py (строгая проверка 6 логов и CSV/таблицы), этот отчёт/сводка.

## Карта вычислений

| Узел | Размерность / стоимость / память | ctx и неделимость |
|---|---|---|
| Input / SQL source | 500 instances, 36 submitted angles; SQL snapshot <=100000 points, 4096 open-ring vertices/polygon, 64 holes, 2000 placements | Budget до материализации повторных геометрий; короткая project transaction |
| Earcut | O(v²) worst-case, arrays O(v), triangles O(v+holes) | ctx до/после Earcut и deviation, в выходном обходе; сам Earcut не прерывается |
| Minkowski | Cartesian product triangles <=65536 до allocation; каждая пара двух треугольников даёт малый convex hull | ctx между парами/пачками, в outer loop sums; не goroutine escape |
| UnionMany | Сложность зависит от пересечений/промежуточного контура; партия <=64 geometries НЕ означает <=64 vertices | ctx до/после каждого library call; крупный поздний union может оставаться неделимым |
| NFP cache | 2048 entries / 8 MiB учтённой геометрии + Go overhead, на run | Exact-bit key и clone; неполный отменённый результат не публикуется |
| Candidates | <=200000; генерация/фильтрация растут с количеством occupied/angles/vertices; временные slices/maps | ctx в NFP angle/occupied loops и выборе кандидата; Baseline generateCandidates проверяется на границах вызова |
| Feasibility | Bounds filter, Transform/Validate/CoveredBy/InteriorsIntersect; много временных геометрий | Native library calls неделимы; ctx между управляемыми шагами |
| GA/SA/Hybrid | Число decoder/iterations умножает стоимость feasibility | ctx в управляемых шагах; production random не менялся |
| Server / persist | validateResult, short final snapshot CAS; atomic history/placements/operation result | Нет write transaction во время Optimize; timeout не создаёт partial layout |

## Среда и метод

Go 1.26.2 linux/amd64. Отдельный созданный этим тестом контейнер nesting-stage6-run: network=none, readonly root, executable tmpfs /work, CPU quota=2, GOMAXPROCS=2, hard memory=1 GiB, swap disabled, GOMEMLIMIT=512 MiB, pid limit128, life sleep<=3600s. PostgreSQL в нём не запускался. Пользовательские контейнеры не останавливались.

Исходный ce28255 экспортирован git archive в /tmp/nesting-stage6-baseline; в него добавлен только тот же тестовый измеритель. Никакие старые патчи не применялись к рабочему дереву. Порядок before1/after1/before2/after2/before3/after3, каждый — новый test process, -count=1, timeout300s. Тяжёлых параллельных тестов во время paired measurements не было. Посторонние процессы хоста не контролировались: это не bare-metal benchmark.

282 строки = 47 fixtures × 6 запусков: 261 completed, 21 ожидаемый bounded timeout. Из completed 30 — явно диагностический radial_without_seed с нулём placements. Все непустые completed проходят ValidatePolygon/CoveredBy/InteriorsIntersect; все timeout возвращают ноль partial placements.

Fixtures: rectangle, L-concave, 1/2 holes, occupied, valid close points/coordinates≈100000, radial concave, 10/50/100/500, 32/64/128 vertices и 36 angles. Для сравнительной GA/SA/Hybrid матрицы фиксирован seed20261120 и существующий comparison profile (GA population8/generations10; SA max100); это НЕ production throughput. Отдельно TestGeometryFixturesPassServerValidation выполняет все 5 алгоритмов на 4 fixtures с production defaults (GA30/50, SA1000), проверяет настоящий server validateResult и отклонение испорченной позы.

Budget эксперимента: обычно10s, quantity50/100 до40s, quantity500 до3s, vertices32/64/128 до5s. Production deadline120s не менялся. Успешное завершение 500 за120s НЕ доказано; timeout по3s — тест управляемого отказа, не production benchmark результата.

Время wall/CPU — фактическое на вызов, TotalAlloc/Mallocs — разность process counters на attempt (включая малую instrumentation), RSS — process cumulative max, а не отдельный peak каждого fixture. Goroutines sampled каждые2ms, поэтому краткие пики могут не попасть в выборку. Cache hits вычислены outer calls−delegate misses только для completed; timeout hits=null. Три повтора дают descriptive median/range, а не p99/95% CI или SLA. Нативный код до/после идентичен, никаких speedup/regression claims на основе шумовых различий.

## Основные числа

На occupied radial12: Baseline50 median18.553s, NFP50 median3.544s; Baseline100 получает timeout40s, NFP100 median13.448s с100 размещениями; оба500 отменяются на3s без partial result. Это только данный fixture. Максимальный process cumulative RSS33400KiB, sampled goroutines4; container cgroup memory.peak=72802304 bytes (включая tmpfs/бинарники), oom/oom_kill=0. Cgroup peak не равен per-operation heap.

CPU/alloc profile одного реального Baseline50: wall19.38s, CPU samples21.21s. mallocgc cumulative27.63% samples; LineString.IsSimple cumulative22.35%; alloc_space sampling: rtree.bulkInsert21.38%, LineString.IsSimple9.45%, toSimpleFeaturesRing7.11%. Перечисленные cumulative проценты перекрываются — складывать их нельзя. Примерно12.6 GB TotalAlloc при небольшом RSS означает интенсивное создание краткоживущих объектов, не 12.6 GB resident memory. Оптимизация prepared geometry/повторной валидации требует отдельного proof эквивалентности; в этом этапе проверку геометрии не отключали.

## Отмена

Каждый алгоритм: 15 измерений, 500 instances / 36 angles /20 obstacles. Замер от фактического ctx cancel до возврата, не включая20ms ожидания timer. Это тестовый fixed profile; default-config cancellation дополнительно проходит в существующем LiveCancellation suite.

| Алгоритм/стадия | n | min ms | median ms | max ms | max интервал ctx checks полного вызова ms |
|---|---:|---:|---:|---:|---:|
| baseline | 15 | 0.011091 | 0.030648 | 0.584394 | 0.000000 |
| nfp_greedy | 15 | 0.011452 | 0.106982 | 0.530913 | 0.000000 |
| genetic | 15 | 0.014948 | 0.037511 | 0.284267 | 0.000000 |
| simulated_annealing | 15 | 0.011652 | 0.025428 | 1.342468 | 0.000000 |
| hybrid | 15 | 0.023946 | 0.061206 | 2.178608 | 0.000000 |
| earcut | 15 | 11.310348 | 13.882330 | 19.615917 | 17.880818 |
| Minkowski_series | 15 | 0.001873 | 0.008396 | 0.398454 | 2.299737 |
| Union_batches | 15 | 1.046298 | 4.276122 | 6.204818 | 18.955198 |

Интервал между настоящими production ctx.Err-проверками включает native функцию, allocations и scheduling; это proxy неделимого участка, НЕ точная изолированная длительность каждой library функции и не верхняя теоретическая граница. Таймерный cancel может попасть в подготовку, поэтому отдельно есть детерминированный rotation barrier на3-м из36 углов (10 race-повторов), реальная серия Minkowski, Union batches и ранний cancelled context. Queue deadline20ms проверяет освобождение слота; native SQL deadline50ms на500/36 проверяет сохранение старого placement99, ноль истории и отсутствие повторного Optimize. Отмена перед commit/timeout↔commit повторяется OS-process harness в PostGIS regression.

## Выявленное ограничение качества

Радиальное вогнутое лекало без bbox-corner на пустом прямоугольнике даёт0 размещений во всех5 алгоритмах. Анализ: vertex-anchor candidates не гарантируют подходящего первого касания. Это воспроизведено и на ce28255, не регрессия новых тестов; пустой ответ геометрически корректен, но качество неудовлетворительно. Этот case оставлен в таблице, а не выдан за успех раскладки. Для реальной нагрузки отдельный occupied seed позволяет начать размещение. Расширение anchor-set/inner-fit — отдельное изменение эвристики, автоматически не внедрено.

## Полная таблица парных измерений

| Fixture / algorithm / q / v / h / angles | Before wall ms median [min,max] | After wall ms median [min,max] | After CPU ms median | Alloc MiB median per attempt | Sampled goroutines max | Process cumulative RSS KiB max | Placed | Outcomes |
|---|---:|---:|---:|---:|---:|---:|---|---|
| 36_angles / nfp_greedy / 2 / 4 / 0 / 36 | 15.449 [15.398,16.103] | 18.347 [17.661,19.750] | 19.110 | 11.20 | 3 | 33316 | 2 | completed:3 |
| close_points / baseline / 6 / 5 / 0 / 2 | 15.979 [14.722,19.470] | 15.550 [15.249,15.724] | 16.827 | 11.04 | 3 | 24236 | 6 | completed:3 |
| close_points / genetic / 6 / 5 / 0 / 2 | 233.487 [226.291,274.924] | 227.951 [227.210,231.361] | 258.294 | 167.60 | 3 | 24236 | 6 | completed:3 |
| close_points / hybrid / 6 / 5 / 0 / 2 | 474.330 [471.327,481.252] | 491.629 [476.637,502.064] | 538.368 | 353.43 | 3 | 24236 | 6 | completed:3 |
| close_points / nfp_greedy / 6 / 5 / 0 / 2 | 14.933 [11.296,17.026] | 11.174 [10.808,11.765] | 12.210 | 7.00 | 3 | 24236 | 6 | completed:3 |
| close_points / simulated_annealing / 6 / 5 / 0 / 2 | 252.025 [251.847,253.011] | 253.130 [239.016,263.521] | 275.638 | 188.07 | 3 | 24236 | 6 | completed:3 |
| concave / baseline / 6 / 6 / 0 / 2 | 30.806 [30.681,36.801] | 31.178 [28.346,31.465] | 28.602 | 19.20 | 3 | 15276 | 6 | completed:3 |
| concave / genetic / 6 / 6 / 0 / 2 | 501.630 [475.385,515.756] | 479.235 [425.259,484.939] | 439.757 | 295.71 | 3 | 16980 | 6 | completed:3 |
| concave / hybrid / 6 / 6 / 0 / 2 | 932.627 [858.243,986.961] | 840.920 [826.358,952.610] | 906.514 | 599.36 | 3 | 24236 | 6 | completed:3 |
| concave / nfp_greedy / 6 / 6 / 0 / 2 | 27.666 [25.845,28.246] | 23.726 [22.500,24.863] | 23.496 | 13.66 | 3 | 15276 | 6 | completed:3 |
| concave / simulated_annealing / 6 / 6 / 0 / 2 | 510.164 [501.961,534.134] | 508.182 [472.267,517.870] | 486.202 | 320.43 | 3 | 17584 | 6 | completed:3 |
| occupied / baseline / 6 / 4 / 0 / 2 | 12.531 [10.873,13.112] | 12.820 [11.085,12.915] | 13.162 | 7.59 | 3 | 24236 | 6 | completed:3 |
| occupied / genetic / 6 / 4 / 0 / 2 | 173.296 [167.335,197.463] | 165.977 [161.200,173.803] | 179.717 | 113.95 | 3 | 24236 | 6 | completed:3 |
| occupied / hybrid / 6 / 4 / 0 / 2 | 424.734 [359.315,437.341] | 364.515 [357.209,379.156] | 396.276 | 249.30 | 3 | 24236 | 6 | completed:3 |
| occupied / nfp_greedy / 6 / 4 / 0 / 2 | 6.357 [6.288,6.523] | 6.637 [6.304,6.850] | 7.334 | 4.14 | 3 | 24236 | 6 | completed:3 |
| occupied / simulated_annealing / 6 / 4 / 0 / 2 | 204.026 [183.911,214.401] | 183.844 [179.873,191.154] | 199.376 | 127.62 | 3 | 24236 | 6 | completed:3 |
| one_hole / baseline / 6 / 4 / 1 / 2 | 25.696 [21.786,27.361] | 28.930 [26.878,30.206] | 29.326 | 15.40 | 3 | 24236 | 6 | completed:3 |
| one_hole / genetic / 6 / 4 / 1 / 2 | 530.339 [528.390,629.088] | 529.913 [506.908,593.399] | 563.076 | 330.84 | 3 | 24236 | 6 | completed:3 |
| one_hole / hybrid / 6 / 4 / 1 / 2 | 1142.377 [1077.357,1163.831] | 1163.220 [1140.544,1222.099] | 1246.679 | 706.46 | 3 | 24236 | 6 | completed:3 |
| one_hole / nfp_greedy / 6 / 4 / 1 / 2 | 15.836 [15.495,17.222] | 17.412 [16.457,18.030] | 16.965 | 9.25 | 3 | 24236 | 6 | completed:3 |
| one_hole / simulated_annealing / 6 / 4 / 1 / 2 | 618.495 [598.739,695.429] | 607.943 [597.318,634.534] | 651.895 | 371.81 | 3 | 24236 | 6 | completed:3 |
| quantity / baseline / 10 / 12 / 0 / 2 | 675.296 [649.026,751.557] | 665.437 [651.231,686.348] | 733.232 | 441.65 | 3 | 33316 | 10 | completed:3 |
| quantity / baseline / 50 / 12 / 0 / 2 | 18411.444 [17770.498,19028.200] | 18552.892 [18131.990,18619.364] | 20304.306 | 12024.71 | 3 | 33316 | 50 | completed:3 |
| quantity / baseline / 100 / 12 / 0 / 2 | 39999.165 [39998.806,39999.612] | 39996.233 [39995.633,39999.342] | 43783.718 | 25332.16 | 3 | 33316 | 0 | timeout:3 |
| quantity / baseline / 500 / 12 / 0 / 2 | 2997.086 [2996.503,2999.142] | 2996.588 [2994.765,2999.560] | 3315.769 | 1919.34 | 3 | 33316 | 0 | timeout:3 |
| quantity / nfp_greedy / 10 / 12 / 0 / 2 | 268.866 [266.700,270.302] | 283.459 [257.820,293.409] | 304.618 | 150.19 | 3 | 33316 | 10 | completed:3 |
| quantity / nfp_greedy / 50 / 12 / 0 / 2 | 3603.262 [3594.122,3623.256] | 3544.352 [3539.331,3672.568] | 3974.268 | 2277.76 | 3 | 33316 | 50 | completed:3 |
| quantity / nfp_greedy / 100 / 12 / 0 / 2 | 14297.055 [13912.652,15181.982] | 13447.871 [13302.953,13960.122] | 15193.740 | 8488.75 | 3 | 33316 | 100 | completed:3 |
| quantity / nfp_greedy / 500 / 12 / 0 / 2 | 2999.597 [2996.156,3003.191] | 2998.068 [2995.420,3000.147] | 3369.882 | 1794.36 | 3 | 33316 | 0 | timeout:3 |
| radial_without_seed / baseline / 6 / 12 / 0 / 2 | 1.110 [1.047,1.143] | 1.051 [1.044,1.085] | 1.096 | 1.06 | 2 | 24236 | 0 | completed:3 |
| radial_without_seed / genetic / 6 / 12 / 0 / 2 | 372.455 [337.500,384.800] | 303.063 [301.957,305.109] | 342.929 | 263.25 | 3 | 24236 | 0 | completed:3 |
| radial_without_seed / hybrid / 6 / 12 / 0 / 2 | 693.198 [682.562,841.210] | 715.014 [680.222,822.716] | 764.073 | 559.18 | 3 | 33316 | 0 | completed:3 |
| radial_without_seed / nfp_greedy / 6 / 12 / 0 / 2 | 1.108 [1.067,1.124] | 1.117 [1.104,1.140] | 1.155 | 1.09 | 2 | 24236 | 0 | completed:3 |
| radial_without_seed / simulated_annealing / 6 / 12 / 0 / 2 | 396.839 [388.489,413.267] | 365.744 [359.216,421.094] | 405.094 | 295.91 | 3 | 24236 | 0 | completed:3 |
| rectangle / baseline / 6 / 4 / 0 / 2 | 8.899 [8.745,8.916] | 9.109 [9.010,9.893] | 8.747 | 5.62 | 3 | 12104 | 6 | completed:3 |
| rectangle / genetic / 6 / 4 / 0 / 2 | 139.763 [132.083,148.334] | 132.469 [126.735,133.243] | 119.897 | 82.32 | 3 | 14276 | 6 | completed:3 |
| rectangle / hybrid / 6 / 4 / 0 / 2 | 281.522 [263.670,292.219] | 268.519 [264.606,277.469] | 254.258 | 175.75 | 3 | 15276 | 6 | completed:3 |
| rectangle / nfp_greedy / 6 / 4 / 0 / 2 | 4.668 [4.374,4.959] | 4.372 [4.320,4.434] | 4.584 | 3.13 | 3 | 12872 | 6 | completed:3 |
| rectangle / simulated_annealing / 6 / 4 / 0 / 2 | 144.224 [128.407,159.112] | 143.717 [135.741,146.005] | 132.854 | 92.35 | 3 | 14276 | 6 | completed:3 |
| two_holes / baseline / 6 / 4 / 2 / 2 | 32.867 [30.342,37.791] | 33.239 [31.585,34.146] | 35.395 | 21.19 | 3 | 24236 | 6 | completed:3 |
| two_holes / genetic / 6 / 4 / 2 / 2 | 707.495 [684.474,740.005] | 716.160 [710.696,748.180] | 773.125 | 455.62 | 3 | 24236 | 6 | completed:3 |
| two_holes / hybrid / 6 / 4 / 2 / 2 | 1679.731 [1529.235,1890.423] | 1519.707 [1515.076,1565.736] | 1655.147 | 972.37 | 3 | 24236 | 6 | completed:3 |
| two_holes / nfp_greedy / 6 / 4 / 2 / 2 | 19.188 [18.924,21.076] | 20.275 [20.103,20.963] | 21.819 | 12.33 | 3 | 24236 | 6 | completed:3 |
| two_holes / simulated_annealing / 6 / 4 / 2 / 2 | 842.722 [802.153,972.844] | 793.182 [790.555,815.698] | 856.207 | 512.26 | 3 | 24236 | 6 | completed:3 |
| vertices / nfp_greedy / 2 / 32 / 0 / 2 | 212.659 [209.916,225.153] | 220.886 [206.699,226.036] | 246.140 | 113.23 | 3 | 33316 | 2 | completed:3 |
| vertices / nfp_greedy / 2 / 64 / 0 / 2 | 1230.700 [1025.446,1272.752] | 1025.183 [1021.938,1040.665] | 1195.528 | 485.20 | 3 | 33316 | 2 | completed:3 |
| vertices / nfp_greedy / 2 / 128 / 0 / 2 | 4999.088 [4832.800,4999.460] | 4830.467 [4794.065,4997.087] | 5726.471 | 2187.70 | 3 | 33316 | 0,2 | completed:2, timeout:1 |

CSV geometry-raw.csv содержит все282 измерения, bytes, allocs, CPU, cache, metrics, RSS и исходы. Рядом сохраняются сырые логи и CPU/mem profiles; scripts/summarize_geometry_experiment.py отвергает неполные/failed логи, неправильные count и partial timeout.

## Не выполнено / границы

Точного пользовательского SVG зависания нет; его исправление не заявляется. Не выполнялся full Union accepted максимума65536 пар, 4096×500 geometry и полностью заполненный200000-candidate workload: это опасное сочетание без установленного предела памяти library. Граничные guard-проверки4096/4097 и64/65 и отказ Cartesian product выполнялись; earcut4096 измерялся, но это не полный NFP4096. Вся production-config матрица100/500 для GA/SA/Hybrid не запускалась; вместо неё small production fixtures, medium fixed-profile и cancellation stress. Не определён worst-case latency произвольной допустимой геометрии; chunk64 не даёт hard cancel.

## Команды / gate

Основные: NESTING_GEOMETRY_EXPERIMENT=1 /work/{before,after}.test -test.run '^TestGeometryReliabilityExperiment$' -test.count=1 -test.v -test.timeout=300s (6 PASS); 3 cancellation runs каждого explicit suite (PASS); pprof -test.cpuprofile/-test.memprofile и go tool pprof -top/-alloc_space (PASS). Все процессы в указанном bounded container.

Go full regression, vet/lint, frontend72 и PostGIS выполняются свежими, финальные результаты добавлены ниже. Для -race указан -count=1 (не кэш) и timeout20m. Explicit heavy experiment/profile по умолчанию SKIP в общем go test, но реально исполнены отдельными командами выше. Служебный subprocess worker SKIP в parent — ожидаемая ветка, не пропуск SQL-инвариантов. Frontend не менялся; Playwright на этапе6 отдельно не запускался (SKIP по условию задачи).

## Ручной checklist

Запустить fixtures в указанном capped container, проверить completed/timeout отдельно, проверить все holes/occupied через server validation. Перед SQL-тестами gate current_database/current_user/postgis_version, только nesting_test. Для реального проблемного SVG сначала сохранить исходник и параметры, затем ограниченный одиночный run/CPU/alloc/cancel; не запускать много500 jobs одновременно. Не менять рабочую БД, source geometry или алгоритм ради зелёного результата.

## Финальный gate этапа 6 — PASS

`go test -count=1 ./...`, `go vet ./...`, `golangci-lint run ./...` — свежие PASS (0 issues). `go test -race -count=1 -timeout=20m ./internal/usecase/nesting ./internal/transport/http/handler ./internal/nesting ./internal/geometry/nfp` — PASS без кэша: usecase96.331s, native nesting318.370s. Новый rotation boundary — 10 дополнительных race-повторов PASS. Полный актуальный PostGIS-набор — PASS; native deadline/OS-process suite дополнительно повторён под race. Frontend build/lint и72 unit-теста — свежие PASS. Production-код и пользовательские данные не затронуты.

В этапе6 также усилен test-only real_process_test.go: контрольные COUNT больше не игнорируют SQL-ошибки, каждый такой запрос имеет deadline5s. Повторены SQL race и полный integration после этой правки.

```text
 backend/internal/integration/project_resources_test.go | 10 +++++++---
 1 file changed, 7 insertions(+), 3 deletions(-)
```

```text
 M backend/internal/integration/project_resources_test.go
?? backend/internal/geometry/nfp/stage_cancellation_experiment_test.go
?? backend/internal/integration/native_deadline_test.go
?? backend/internal/integration/real_process_test.go
?? backend/internal/nesting/cancellation_measurement_test.go
?? backend/internal/nesting/geometry_experiment_test.go
?? backend/internal/usecase/nesting/geometry_fixture_validation_test.go
?? docs/stage5_real_process_reliability_report.md
?? docs/stage6_geometry_reliability_report.md
?? docs/stages5_7_summary.md
?? scripts/
```
