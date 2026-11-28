import { LogOut, User } from 'lucide-react'
import type { AuthUser } from '@/api/auth'
import { Button } from '@/components/ui/button'

type UserMenuProps = {
    user: AuthUser | null
    onLogin: () => void
    onLogout: () => void
}

function initials(email: string) {
    const name = email.split('@')[0]
    const parts = name.split(/[._-]+/).filter(Boolean)
    const letters = parts.length > 1 ? parts[0][0] + parts[1][0] : name.slice(0, 2)
    return letters.toUpperCase()
}

export function UserMenu({ user, onLogin, onLogout }: UserMenuProps) {
    if (!user) {
        return (
            <Button className="h-10" variant="secondary" onClick={onLogin}>
                <User />
                Войти
            </Button>
        )
    }

    return (
        <div className="animate-in fade-in duration-200 flex items-center gap-2">
            <span
                className="flex size-9 shrink-0 items-center justify-center rounded-full bg-accent text-xs font-medium text-accent-foreground"
                aria-hidden
            >
                {initials(user.email)}
            </span>
            <span className="hidden max-w-48 truncate text-sm md:inline" title={user.email}>
                {user.email}
            </span>
            <Button variant="ghost" size="icon" className="size-10" aria-label="Выйти" title="Выйти" onClick={onLogout}>
                <LogOut />
            </Button>
        </div>
    )
}
