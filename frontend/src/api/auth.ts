import axios from 'axios'
import { apiClient } from './client'
import { getApiErrorMessage } from './errors'

export type AuthUser = {
    id: string
    email: string
    name: string
}

// The session lives in an HttpOnly cookie, so the server is the only source
// of truth for the signed-in state; localStorage is never trusted.
export async function getCurrentUser(): Promise<AuthUser | null> {
    try {
        const { data } = await apiClient.get<AuthUser>('/auth/me')
        return data
    } catch (e) {
        if (isStatus(e, 401)) return null
        throw e
    }
}

export async function login(email: string, password: string): Promise<AuthUser> {
    if (!email || !password) throw new Error('Неверный email или пароль.')

    try {
        const { data } = await apiClient.post<AuthUser>('/auth/login', { email: email.trim(), password })
        return data
    } catch (e) {
        throw new Error(getApiErrorMessage(e), { cause: e })
    }
}

export async function register(email: string, password: string): Promise<AuthUser> {
    if (!email || !password) throw new Error('Неверный email или пароль.')

    try {
        const { data } = await apiClient.post<AuthUser>('/auth/register', { email: email.trim(), password })
        return data
    } catch (e) {
        throw new Error(getApiErrorMessage(e), { cause: e })
    }
}

export async function logout(): Promise<void> {
    try {
        await apiClient.post('/auth/logout')
    } catch (e) {
        if (isStatus(e, 401)) return
        throw e
    }
}

function isStatus(e: unknown, status: number): boolean {
    return axios.isAxiosError(e) && e.response?.status === status
}
