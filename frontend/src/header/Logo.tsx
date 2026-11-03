import logoUrl from '@/assets/logo.svg'

export function Logo() {
    return (
        <div className="flex items-center gap-3">
            <img src={logoUrl} alt="" className="size-9" />
            <div className="flex flex-col gap-1">
                <div className="text-lg leading-none font-semibold tracking-tight">Магистраль</div>
                <div className="text-xs leading-none text-muted-foreground">анализ логистической сети</div>
            </div>
        </div>
    )
}
