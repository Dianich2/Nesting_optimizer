# Этап22: числовые поля и приёмка интерфейса

База `1d405973c77f892a4bca888d8aa1eef51d055a1c`, `wip/geometry-editor-v2`, в начале дерево чистое. Изменения локальные. Gate21 PASS до начала22; окончательная общая регрессия выполняется после исправлений всех найденных случаев. Рабочие БД/данные/службы не менялись. Использованы собственные Docker API и PostGIS `nesting_test`, их ID и миграции проверены до записи.

## Подтверждённые дефекты и изменения

- Количества немедленно преобразовывались через Number('')→0. Теперь `Record<number,string>`, пустое значение/placeholder0, raw draft и `quantityDraft.ts` с валидацией перед POST. Пусто/0 пропускают лекало; хотя бы одна положительная quantity нужна. Только целые0..500, safe integer, сумма≤500. −1/1.5/NaN/501/overflow отклоняются. NaN проверяется unit-тестом: HTML number input сам очищает такой текст.
- Шаг сетки холста и геометрического редактора имел ту же проблему. Хранится строка, пустую можно стереть/набрать; до валидного шага≥0,01 используется последнее валидное число. На холсте есть пояснение ошибочного значения.
- Width/height геометрического редактора, LibraryPicker.scale, PlacementProperties X/Y/rotation **уже были строковыми**. Их не переписывали. Реальный браузер проверяет Ctrl+A/Backspace/новый ввод во всех этих полях; итоговые quantity POST2 и pose PATCH90° проходят API. Диалог масштабирования открывается штатным маршрутом `/projects/:id/editor`.
- Кнопки автораскроя в grid, gap10px, width100%, высота≥40px; aside padding16px, перенос длинного текста и focus styles. Есть объяснение preview/прямого save/120с/препятствий, точная кнопка «Сохранить этот предпросмотр». Сохранение preview не пересчитывает результат.
- При320px/zoom200% выявлен overflow не кнопок, а глобального body min-width320 и шапки. Для страницы раскроя адресно снят min-width; шапка переносится, включая промежуточные720CSSpx. Остальные страницы не переделывались. Строка результата сделана13px/1.4, чтобы сохранять полезную высоту холста. Left/right panels по-прежнему сворачиваются независимо, панели прокручиваются.

## Проверки

`quantity-draft.test.mjs`:2 unit теста с большим набором входов — пусто/0/leading zeros, −1,1.5,NaN,501,unsafe integer, sum500/501. Вместе с3draft tests полный frontend unit suite77тестов. Подробные exit codes в сводке.

`stage22.spec.mjs` numeric: реальный браузер Ctrl+A→Backspace→2; несколько неверных вводов и sum501 дают понятное сообщение и0nesting POST. Валидный запрос содержит только выбранный pattern_id/quantity2, SQL2placements. Существующие X/Y/angle очищаются и вводятся заново, PATCH200; grid step5, library scale2, editor width25/height15/grid5 очищаются без подстановки нуля. Пилоты с неверным названием кнопки закрытия и маршрутом редактора были FAIL теста; исправлены по существующему UI, не обходом.

`stage22.spec.mjs` layout:8ширин320/390/700/768/1024/1366/1440/1920 ×2темы ×3масштаба=48. Длинные русские/английские названия, реальные проект/поверхность/лекала. Для масштабов Playwright задаёт CSS viewport=floor(physicalWidth/zoom), CSS height=floor(1000/zoom), DPR=zoom. Это **эквивалент viewport/DPR для layout**, а не автоматизация нативного меню zoom Chrome. Нативный browser zoom остаётся ручной проверкой. Между контекстами передаётся актуальное storageState с новой refresh cookie (секреты не записаны в отчёт).

Замеры `getBoundingClientRect`: gap≥8 (фактически10), scrollWidth≤clientWidth+1 у document/страницы/холста/панелей, обе кнопки доступны для обычного Playwright trial click, сняты48screenshots. Снимки и JSON замеров сохраняются вне `/tmp` с финальным отчётом. Первый пилот проверял только panel overflow и пропустил global body overflow; screenshot выявил недостаточность этого assertion. Проверка усилена до document scrollWidth, исправления прошли48/48. Скриншоты не равны ручной визуальной UX-приёмке; пересечения всех произвольных label геометрически не доказаны.

Полная19-test browser regression включает: обе темы, SVG/DXF export, библиотечные копии/редактор/проект; drag/rotation/undo/redo/reload; все5nativeалгоритмов exact preview save с SQL/Optimize counter; stale-tab; fake transport faults поверх реального API для потерянного ответа/unknown/pending/refresh; дополнительно новые real20/21/22. Production HTTP/JWT/Go/PostGIS не заменены mock алгоритмом. Transport interception используется именно для сценариев потери ответа, а не доказательства геометрии.

Старые assertions текста quantity обновлены под конкретный новый текст. Touch-тест закрывает открываемую после success панель штатной кнопкой. Ранние полные запуски со старыми ожиданиями не были засчитаны. Сохраняются их FAIL logs. Появлявшийся пустой экран при dev navigation исследуется безопасной диагностикой client errors/network без cookies/headers/storage; итоговая классификация указана в сводке.

## Команды и ограничения

Gate22: `python3 scripts/run_stage16_browser.py --label stage22-gate3 --lab-metadata /tmp/nesting-2022-lab.json --api-binary /tmp/nesting2022-final-race.test stage22.spec.mjs` PASS2; затем расширенный numeric PASS1. Итоговые19-test запуски/retries0, Go/PostGIS/race/build/lint/unit — см. сводку.

Проверить вручную: физический zoom Chrome150/200%, реальные touch/другие браузеры, читабельность при длинном названии и высокой плотности placements, исходный проблемный проект. .env, миграции, алгоритмы,120/130с и strict predicates сохранены. Автоматического исправления всех возможных пользовательских SVG или фоновой очереди не добавлялось.
