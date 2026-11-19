import { useState, type FormEvent } from 'react'
import { AlertCircle, Eye, EyeOff, Loader2 } from 'lucide-react'
import { login, register, type AuthUser } from '@/api/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { SegmentedControl } from '@/components/ui/segmented-control'

type Mode = 'login' | 'register'

type Errors = { email?: string; password?: string; form?: string }

const MODE_OPTIONS: { value: Mode; label: string }[] = [
    { value: 'login', label: 'Вход' },
    { value: 'register', label: 'Регистрация' },
]

const MIN_PASSWORD_LENGTH = 8

type AuthFormProps = {
    onSuccess: (user: AuthUser) => void
}

export function AuthForm({ onSuccess }: AuthFormProps) {
    const [mode, setMode] = useState<Mode>('login')
    const [email, setEmail] = useState('')
    const [password, setPassword] = useState('')
    const [showPassword, setShowPassword] = useState(false)
    const [pending, setPending] = useState(false)
    const [errors, setErrors] = useState<Errors>({})

    function switchMode(next: Mode) {
        setMode(next)
        setErrors({})
    }

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault()
        if (pending) return

        const next: Errors = {}
        if (!/^\S+@\S+\.\S+$/.test(email.trim())) next.email = 'Введите корректный email'
        if (password.length < MIN_PASSWORD_LENGTH) next.password = `Пароль должен содержать не менее ${MIN_PASSWORD_LENGTH} символов`
        setErrors(next)
        if (next.email || next.password) return

        setPending(true)
        try {
            const user = mode === 'login' ? await login(email, password) : await register(email, password)
            onSuccess(user)
        } catch (e) {
            setErrors({ form: e instanceof Error ? e.message : 'Не удалось выполнить вход. Попробуйте ещё раз.' })
        } finally {
            setPending(false)
        }
    }

    return (
        <div className="flex w-full max-w-sm flex-col gap-8">
            <div className="flex flex-col gap-2">
                <h1 className="text-2xl font-semibold tracking-tight text-balance">
                    {mode === 'login' ? 'Вход в аккаунт' : 'Создание аккаунта'}
                </h1>
                <p className="text-sm leading-relaxed text-muted-foreground">
                    {mode === 'login'
                        ? 'Продолжите анализ логистической сети вашего региона.'
                        : 'Стройте графы инфраструктуры и находите ключевые узлы.'}
                </p>
            </div>

            <SegmentedControl label="Вход или регистрация" options={MODE_OPTIONS} value={mode} onChange={switchMode} />

            <form onSubmit={handleSubmit} noValidate className="flex flex-col gap-5">
                {errors.form && (
                    <div role="alert" className="animate-in fade-in duration-200 flex gap-2 rounded-md border border-destructive/30 bg-destructive/5 px-3 py-2.5 text-sm text-destructive">
                        <AlertCircle className="mt-0.5 size-4 shrink-0" aria-hidden />
                        <p className="leading-relaxed">{errors.form}</p>
                    </div>
                )}

                <div className="flex flex-col gap-2">
                    <Label htmlFor="auth-email">Email</Label>
                    <Input
                        id="auth-email"
                        type="email"
                        autoComplete="email"
                        autoFocus
                        placeholder="name@company.ru"
                        className="h-10"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        aria-invalid={!!errors.email}
                        aria-describedby={errors.email ? 'auth-email-error' : undefined}
                    />
                    {errors.email && (
                        <p id="auth-email-error" className="text-xs text-destructive">
                            {errors.email}
                        </p>
                    )}
                </div>

                <div className="flex flex-col gap-2">
                    <Label htmlFor="auth-password">Пароль</Label>
                    <div className="relative">
                        <Input
                            id="auth-password"
                            type={showPassword ? 'text' : 'password'}
                            autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
                            className="h-10 pr-10"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            aria-invalid={!!errors.password}
                            aria-describedby={errors.password ? 'auth-password-error' : undefined}
                        />
                        <button
                            type="button"
                            onClick={() => setShowPassword((s) => !s)}
                            className="absolute inset-y-0 right-0 flex w-10 items-center justify-center text-muted-foreground transition-colors hover:text-foreground"
                            aria-label={showPassword ? 'Скрыть пароль' : 'Показать пароль'}
                        >
                            {showPassword ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
                        </button>
                    </div>
                    {errors.password ? (
                        <p id="auth-password-error" className="text-xs text-destructive">
                            {errors.password}
                        </p>
                    ) : (
                        mode === 'register' && (
                            <p className="text-xs text-muted-foreground">Не менее {MIN_PASSWORD_LENGTH} символов</p>
                        )
                    )}
                </div>

                <Button type="submit" disabled={pending} className="h-10 w-full">
                    {pending && <Loader2 className="animate-spin" data-icon="inline-start" />}
                    {mode === 'login' ? 'Войти' : 'Зарегистрироваться'}
                </Button>
            </form>

            <p className="text-sm text-muted-foreground">
                {mode === 'login' ? 'Нет аккаунта? ' : 'Уже есть аккаунт? '}
                <button
                    type="button"
                    onClick={() => switchMode(mode === 'login' ? 'register' : 'login')}
                    className="font-medium text-primary hover:underline"
                >
                    {mode === 'login' ? 'Зарегистрироваться' : 'Войти'}
                </button>
            </p>
        </div>
    )
}
