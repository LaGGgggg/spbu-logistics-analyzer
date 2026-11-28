import { useEffect } from 'react'
import { ArrowLeft } from 'lucide-react'
import type { AuthUser } from '@/api/auth'
import { Button } from '@/components/ui/button'
import { Logo } from '@/header/Logo'
import { AuthForm } from './AuthForm'

type AuthPageProps = {
    onSuccess: (user: AuthUser) => void
    onClose: () => void
}

export function AuthPage({ onSuccess, onClose }: AuthPageProps) {
    useEffect(() => {
        function handleKeyDown(event: KeyboardEvent) {
            if (event.key === 'Escape') onClose()
        }
        document.addEventListener('keydown', handleKeyDown)
        return () => document.removeEventListener('keydown', handleKeyDown)
    }, [onClose])

    return (
        <div
            role="dialog"
            aria-modal="true"
            aria-label="Вход в аккаунт"
            className="animate-in fade-in duration-200 fixed inset-0 z-[1100] grid grid-cols-1 bg-card lg:grid-cols-[minmax(480px,1fr)_1.2fr]"
        >
            <div className="flex flex-col overflow-y-auto px-8 py-6">
                <div className="flex items-center justify-between gap-4">
                    <Logo />
                    <Button variant="ghost" className="h-10" onClick={onClose}>
                        <ArrowLeft data-icon="inline-start" />
                        К карте
                    </Button>
                </div>
                <div className="flex flex-1 items-center justify-center py-12">
                    <AuthForm onSuccess={onSuccess} />
                </div>
            </div>
            <div className="relative hidden lg:block">
                <img
                    src="/images/auth-port.webp"
                    alt="Контейнерный терминал морского порта с высоты"
                    className="absolute inset-0 size-full object-cover"
                />
            </div>
        </div>
    )
}
