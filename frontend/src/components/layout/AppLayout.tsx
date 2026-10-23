import type { ReactNode } from "react"

type Props = {
    header: ReactNode,
    left: ReactNode,
    right: ReactNode,
    children: ReactNode
}

export function AppLayout({ header, left, right, children }: Props) {
    return (
        <div className="flex h-svh flex-col">
            <header className="flex h-18 shrink-0 items-center border-b bg-card px-4">
                {header}
            </header>

            <div className="flex min-h-0 flex-1">
                <aside className="flex w-80 shrink-0 flex-col overflow-y-auto border-r bg-card">
                    {left}
                </aside>

                <main className="relative isolate min-w-0 flex-1">
                    {children}
                </main>

                <aside className="flex w-95 shrink-0 flex-col overflow-y-auto border-l bg-card">
                    {right}
                </aside>
            </div>
        </div>
    )
}