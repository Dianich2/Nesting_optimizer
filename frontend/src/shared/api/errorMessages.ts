import { ApiError } from './http'

const FIELD_NAMES: Record<string, string> = {
    name: 'Название',
    geometry: 'Геометрия',
    scale: 'Масштаб геометрии',
    login: 'Логин',
    email: 'Электронная почта',
    password: 'Пароль',
    first_name: 'Имя',
    last_name: 'Фамилия',
    identifier: 'Логин или электронная почта',
}

function getFieldErrorMessage(field: string, code: string): string {
    const name = FIELD_NAMES[field] ?? 'Поле'

    switch (code) {
        case 'required':
            return `${name}: обязательное поле.`

        case 'too_short':
            return `${name}: слишком короткое значение.`

        case 'too_long':
            return `${name}: слишком длинное значение.`

        case 'invalid':
            return `${name}: некорректное значение.`

        case 'already_exists':
            return `${name}: такое значение уже используется.`

        default:
            return `${name}: проверьте введённое значение.`
    }
}

export function getApiErrorMessage(
    error: unknown,
    context: 'login' | 'register' | 'general' = 'general',
): string {
    if (!(error instanceof ApiError)) {
        return 'Не удалось связаться с сервером. Проверьте подключение.'
    }

    if (error.code === 'nesting_cancelled') return 'Расчёт отменён. Сохранённая раскладка не изменилась.'
    if (error.code === 'nesting_timeout') return 'Расчёт превысил лимит времени. Попробуйте меньше деталей или более быстрый алгоритм.'
    if (error.code === 'nesting_not_running') return 'Сервер ещё не начал расчёт или уже завершил его.'
    if (error.details.some((detail) => detail.field === 'quantity' && detail.message.includes('500'))) return 'За один запуск можно разместить не более 500 деталей.'
    if (error.details.some((detail) => detail.field === 'allowed_rotations' && detail.message.includes('36'))) return 'Укажите не более 36 допустимых углов.'
    if (context === 'login' && error.status === 401) {
        return 'Неверный логин, электронная почта или пароль.'
    }

    if (context === 'register' && error.status === 409) {
        if (error.message === 'login already exists') {
            return 'Пользователь с таким логином уже существует.'
        }

        if (error.message === 'email already exists') {
            return 'Пользователь с такой электронной почтой уже существует.'
        }

        return 'Логин или электронная почта уже используются.'
    }

    if (
        error.status === 409 &&
        /project (surface|pattern) is used by active placements/.test(error.message)
    ) {
        return 'Сначала удалите размещения, использующие этот экземпляр. Сейчас можно изменить только название.'
    }

    const placementReason = error.details.find(({ field }) => field === 'placement')?.code
    switch (placementReason) {
        case 'placement_version_conflict':
            return 'Размещение изменилось после загрузки. Обновите раскладку.'
        case 'placement_in_hole':
            return 'Лекало попадает в отверстие поверхности.'
        case 'placement_outside_surface':
            return 'Лекало выходит за границу поверхности.'
        case 'placement_overlap':
            return 'Лекало пересекает другое размещение.'
    }

    if (error.status === 404) {
        if (error.message === 'project pattern not found')
            return 'Лекало проекта недоступно или удалено.'
        if (error.message === 'project surface not found')
            return 'Поверхность проекта недоступна или удалена.'
        if (error.message === 'project not found') return 'Проект недоступен или удалён.'
    }
    if (error.status === 503) return 'Сервер временно занят. Попробуйте позже.'

    if (error.status === 400 && error.details.length > 0) {
        return error.details
            .map((detail) => getFieldErrorMessage(detail.field, detail.code))
            .join(' ')
    }

    switch (error.status) {
        case 400:
            return 'Проверьте правильность введённых данных.'
        case 401:
            return 'Необходимо войти в систему.'
        case 403:
            return 'У вас нет доступа к этому действию.'
        case 404:
            return 'Запрашиваемый объект не найден.'
        case 409:
            return 'Не удалось выполнить действие из-за конфликта данных.'
        case 429:
            return 'Слишком много запросов. Попробуйте позже.'
        default:
            return 'Не удалось выполнить запрос. Попробуйте позже.'
    }
}
