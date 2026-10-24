import { Checkbox } from '@/components/ui/checkbox'
import { Switch } from '@/components/ui/switch'
import type { TransportMode } from '@/api/types'
import { MODES, MODE_COLOR, MODE_LABEL } from '@/lib/modes'
import { kindLabel } from '@/lib/kinds'

type LayersPanelProps = {
    visibleModes: Set<TransportMode>
    counts: Record<TransportMode, number>
    onToggleMode: (mode: TransportMode) => void

    kindCounts: [string, number][]
    hiddenKinds: Set<string>
    onToggleKind: (kind: string) => void

    showEdges: boolean
    onShowEdgesChange: (value: boolean) => void
}

export function LayersPanel({
    visibleModes, counts, onToggleMode,
    kindCounts, hiddenKinds, onToggleKind,
    showEdges, onShowEdgesChange,
}: LayersPanelProps) {
    return (
        <div className="flex flex-col gap-4 border-t p-4">
            <h3 className="text-lg font-semibold">Слои</h3>

            <section className="flex flex-col gap-1">
                <h4 className="text-xs font-medium text-muted-foreground">Вид транспорта</h4>
                {MODES.map((mode) => (
                    <label key={mode} className="flex cursor-pointer items-center gap-2.5 rounded-md py-1">
                        <Checkbox
                            className="size-5"
                            checked={visibleModes.has(mode)}
                            onCheckedChange={() => onToggleMode(mode)}
                        />
                        <span className="size-3 rounded-full" style={{ backgroundColor: MODE_COLOR[mode] }} />
                        <span className="flex-1 text-sm">{MODE_LABEL[mode]}</span>
                        <span className="text-sm tabular-nums">{counts[mode]}</span>
                    </label>
                ))}
            </section>

            <section className="flex flex-col gap-1">
                <h4 className="text-xs font-medium text-muted-foreground">Тип узла</h4>
                {kindCounts.map(([kind, count]) => (
                    <label key={kind} className="flex cursor-pointer items-center gap-2.5 rounded-md py-1">
                        <Checkbox
                            className="size-5"
                            checked={!hiddenKinds.has(kind)}
                            onCheckedChange={() => onToggleKind(kind)}
                        />
                        <span className="flex-1 text-sm">{kindLabel(kind)}</span>
                        <span className="text-sm tabular-nums">{count}</span>
                    </label>
                ))}
            </section>

            <label className="flex cursor-pointer items-center justify-between gap-2.5">
                <span className="text-sm">Показывать связи</span>
                <Switch checked={showEdges} onCheckedChange={onShowEdgesChange} />
            </label>
        </div>
    )
}
