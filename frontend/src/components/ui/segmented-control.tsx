type SegmentedOption<T extends string> = {
    value: T
    label: string
    hint?: string
}

type SegmentedControlProps<T extends string> = {
    options: SegmentedOption<T>[]
    value: T
    onChange: (value: T) => void
    columns?: number
    label: string
}

const GAP = '0.25rem'

export function SegmentedControl<T extends string>({ options, value, onChange, columns = options.length, label }: SegmentedControlProps<T>) {
    const rows = Math.ceil(options.length / columns)
    const index = Math.max(options.findIndex((option) => option.value === value), 0)
    const column = index % columns
    const row = Math.floor(index / columns)

    return (
        <div className="rounded-lg bg-muted p-1">
            <div
                role="group"
                aria-label={label}
                className="relative grid gap-1"
                style={{ gridTemplateColumns: `repeat(${columns}, minmax(0, 1fr))` }}
            >
                <span
                    aria-hidden
                    className="absolute top-0 left-0 rounded-md bg-card shadow-sm transition-transform duration-300 ease-out motion-reduce:transition-none"
                    style={{
                        width: `calc((100% - ${GAP} * ${columns - 1}) / ${columns})`,
                        height: `calc((100% - ${GAP} * ${rows - 1}) / ${rows})`,
                        transform: `translate(calc(${column} * (100% + ${GAP})), calc(${row} * (100% + ${GAP})))`,
                    }}
                />
                {options.map((option) => (
                    <button
                        key={option.value}
                        type="button"
                        title={option.hint}
                        aria-pressed={option.value === value}
                        className={`relative h-8 truncate rounded-md px-2 text-sm transition-colors duration-200 ${option.value === value
                            ? 'font-medium text-foreground'
                            : 'text-muted-foreground hover:text-foreground'}`}
                        onClick={() => onChange(option.value)}
                    >
                        {option.label}
                    </button>
                ))}
            </div>
        </div>
    )
}
