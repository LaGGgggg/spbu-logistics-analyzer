import type { ReactNode } from "react"
import { useState } from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";

type Props = {
    header: ReactNode,
    left: ReactNode,
    right: ReactNode,
    children: ReactNode
}

export function AppLayout({ header, left, right, children }: Props) {
    const [isRightOpen, setIsRightOpen] = useState(false);

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

                <div
                    className={`relative shrink-0 transition-[width] duration-300 ${
                        isRightOpen ? "w-95" : "w-0"
                    }`}
                >
                    <button
                        type="button"
                        onClick={() => setIsRightOpen(!isRightOpen)}
                        aria-label={isRightOpen ? "Скрыть правую панель" : "Открыть правую панель"}
                        aria-expanded={isRightOpen}
                        className="absolute right-full top-4 z-[1000] rounded-l-lg border bg-card p-2 shadow-sm hover:bg-muted"
                    >
                        {isRightOpen ? (
                            <ChevronRight className="h-5 w-5" />
                        ) : (
                            <ChevronLeft className="h-5 w-5" />
                        )}
                    </button>

                    <div
                        className="h-full overflow-hidden"
                        inert={!isRightOpen}
                    >
                        <aside className="flex h-full w-95 flex-col overflow-y-auto border-l bg-card">
                            {right}
                        </aside>
                    </div>
                </div>
            </div>
        </div>
    )
}