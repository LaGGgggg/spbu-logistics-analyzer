export type AuthUser = {
    email: string
}

const STORAGE_KEY = 'magistral.user'

function readStoredUser(): AuthUser | null {
    try {
        const raw = localStorage.getItem(STORAGE_KEY)
        if (!raw) return null
        const parsed = JSON.parse(raw) as Partial<AuthUser>
        return typeof parsed.email === 'string' ? { email: parsed.email } : null
    } catch {
        return null
    }
}

function storeUser(user: AuthUser | null) {
    try {
        if (user) localStorage.setItem(STORAGE_KEY, JSON.stringify(user))
        else localStorage.removeItem(STORAGE_KEY)
    } catch {
        return
    }
}

export function getCurrentUser(): AuthUser | null {
    return readStoredUser()
}

export async function login(email: string, password: string): Promise<AuthUser> {
    if (!email || !password) throw new Error('Неверный email или пароль.')
    const user = { email: email.trim().toLowerCase() }
    storeUser(user)
    return user
}

export async function register(email: string, password: string): Promise<AuthUser> {
    return login(email, password)
}

export async function logout() {
    storeUser(null)
}
