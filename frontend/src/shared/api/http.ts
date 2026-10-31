
const API_BASE_URL = '/api/v1'


export interface ApiFieldError {
    field: string
    code: string
    message: string
}

interface ApiErrorBody {
    error?: {
        code?: string
        message?: string
        details?: ApiFieldError[]
    }
}

export class ApiError extends Error {
    public readonly status: number
    public readonly code?: string
    public readonly details: ApiFieldError[]

    constructor(
        status: number,
        code?: string,
        message = 'Ошибка запроса к серверу',
        details: ApiFieldError[] = [],
    ) {
        super(message)

        this.name = 'ApiError'
        this.status = status
        this.code = code
        this.details = details
    }
}


export async function apiRequest<T>(
    path: string,
    options: RequestInit = {},
): Promise<T> {
    const headers = new Headers(options.headers)

    if (options.body != null && !headers.has('Content-Type')) {
        headers.set('Content-Type', 'application/json')
    }

    const response = await fetch(`${API_BASE_URL}${path}`, {
        ...options,
        headers,
    })

    if (!response.ok) {
        let errorBody: ApiErrorBody | null = null

        try {
            errorBody = (await response.json()) as ApiErrorBody
        } catch {
            // Сервер может вернуть ответ без JSON.
        }

        throw new ApiError(
            response.status,
            errorBody?.error?.code,
            errorBody?.error?.message || `Ошибка HTTP ${response.status}`,
            errorBody?.error?.details ?? [],
        )
    }

    if (response.status === 204) {
        return undefined as T
    }

    return (await response.json()) as T
}
