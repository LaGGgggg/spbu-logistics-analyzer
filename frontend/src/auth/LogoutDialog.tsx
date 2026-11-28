import { useEffect, useRef } from 'react'
import { Button } from '@/components/ui/button'

type LogoutDialogProps = {
    open: boolean
    onClose: () => void
    onConfirm: () => void
}

export function LogoutDialog({
    open,
    onClose,
    onConfirm,
}: LogoutDialogProps) {
    const dialogRef = useRef<HTMLDialogElement>(null)

    useEffect(() => {
        const dialog = dialogRef.current
        if (!dialog) return

        if (open && !dialog.open) {
            dialog.showModal()
        } else if (!open && dialog.open) {
            dialog.close()
        }
    }, [open])

    return (
        <dialog
            ref={dialogRef}
            onClose={onClose}
            aria-labelledby="logout-title"
            aria-describedby="logout-description"
            className="fixed inset-0 m-auto w-[calc(100%-2rem)] max-w-sm rounded-xl border border-border bg-card p-6 text-foreground shadow-xl backdrop:bg-black/50"
        >
            <h2 id="logout-title" className="text-lg font-semibold">
                Выйти из аккаунта?
            </h2>

            <p
                id="logout-description"
                className="mt-2 text-sm text-muted-foreground"
            >
                Вы действительно хотите выйти?
            </p>

            <div className="mt-6 flex justify-end gap-2">
                <Button variant="outline" onClick={onClose} >
                    Отмена
                </Button>

                <Button variant="destructive" onClick={onConfirm}>
                    Выйти
                </Button>
            </div>
        </dialog>
    )
}