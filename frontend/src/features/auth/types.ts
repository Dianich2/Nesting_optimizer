
export interface User {
    id: number
    login: string
    email: string
    first_name: string
    last_name: string
    created_at: string
    updated_at: string
}

export interface RegisterRequest {
    login: string
    email: string
    password: string
    first_name: string
    last_name: string
}

export interface LoginRequest {
    identifier: string
    password: string
}

export interface TokenPair {
    access_token: string
    refresh_token: string
}
