import { useEffect, useMemo, useState } from 'react'
import { Loader2 } from 'lucide-react'
import { getGraphMetrics } from '@/api/graph'
import { getApiErrorMessage } from '@/api/errors'
import type { GraphMetrics, LogisticsGraph, TransportMode } from '@/api/types'
import { SegmentedControl } from '@/components/ui/segmented-control'
import { MODES, MODE_LABEL } from '@/lib/modes'
import { formatMetric, METRICS, metricDisplayValue, rankMetrics, type MetricKey } from '@/lib/metrics'

type ModeFilter = TransportMode | 'all'

type LoadState = {
    metrics?: GraphMetrics
    error?: string
}

type MetricsPanelProps = {
    graph: LogisticsGraph | null
    onSelectNode: (id: string) => void
}

type Bin = { from: number; to: number; count: number }

const BIN_COUNT = 10
const TOP_SIZE = 10

const METRIC_OPTIONS = METRICS.map((metric) => ({ value: metric.key, label: metric.label, hint: metric.hint }))

const MODE_OPTIONS: { value: ModeFilter; label: string }[] = [
    { value: 'all', label: 'Все' },
    ...MODES.map((mode) => ({ value: mode, label: MODE_LABEL[mode] })),
]

function buildBins(key: MetricKey, values: number[]): Bin[] {
    if (values.length === 0) return []

    const max = values.reduce((a, b) => Math.max(a, b))
    const min = key === 'degree' ? values.reduce((a, b) => Math.min(a, b)) : 0
    const span = max - min

    if (key === 'degree' && span < BIN_COUNT) {
        const bins = Array.from({ length: span + 1 }, (_, i) => ({ from: min + i, to: min + i, count: 0 }))
        for (const v of values) bins[v - min].count++
        return bins
    }

    const width = (span || 1) / BIN_COUNT
    const bins = Array.from({ length: BIN_COUNT }, (_, i) => ({ from: min + i * width, to: min + (i + 1) * width, count: 0 }))
    for (const v of values) bins[Math.min(Math.floor((v - min) / width), BIN_COUNT - 1)].count++
    return bins
}

function binLabel(key: MetricKey, bin: Bin) {
    if (key === 'degree' && bin.from === bin.to) return `Степень ${bin.from}`
    return `${formatMetric(key, key === 'degree' ? Math.ceil(bin.from) : bin.from)} – ${formatMetric(key, key === 'degree' ? Math.floor(bin.to) : bin.to)}`
}

function Histogram({ metricKey, bins, total }: { metricKey: MetricKey; bins: Bin[]; total: number }) {
    const [hovered, setHovered] = useState<number | null>(null)
    const maxCount = bins.reduce((max, bin) => Math.max(max, bin.count), 1)
    const active = hovered === null ? null : bins[hovered]

    return (
        <div>
            <p className="mb-2 text-xs text-muted-foreground tabular-nums">
                {active ? `${binLabel(metricKey, active)}: ${active.count} узл.` : `Всего узлов: ${total}`}
            </p>
            <div className="flex h-24 items-end gap-0.5 border-b" onMouseLeave={() => setHovered(null)}>
                {bins.map((bin, i) => (
                    <div
                        key={i}
                        role="img"
                        aria-label={`${binLabel(metricKey, bin)}: ${bin.count} узл.`}
                        className="flex h-full min-w-0 flex-1 items-end"
                        onMouseEnter={() => setHovered(i)}
                    >
                        <div
                            className={`w-full rounded-t bg-primary transition-[height,opacity] duration-300 ease-out motion-reduce:transition-none ${hovered !== null && hovered !== i ? 'opacity-40' : ''}`}
                            style={{ height: `${(bin.count / maxCount) * 100}%`, minHeight: bin.count > 0 ? 2 : 0 }}
                        />
                    </div>
                ))}
            </div>
            <div className="mt-1 flex justify-between text-xs text-muted-foreground tabular-nums">
                <span>{formatMetric(metricKey, metricKey === 'degree' ? bins[0].from : 0)}</span>
                <span>{formatMetric(metricKey, metricKey === 'degree' ? Math.floor(bins[bins.length - 1].to) : 1)}</span>
            </div>
        </div>
    )
}

export function MetricsPanel({ graph, onSelectNode }: MetricsPanelProps) {
    const [metricKey, setMetricKey] = useState<MetricKey>('degree')
    const [mode, setMode] = useState<ModeFilter>('all')
    const [cache, setCache] = useState<Record<string, LoadState>>({})
    const [previous, setPrevious] = useState<GraphMetrics | undefined>()

    const graphId = graph?.id
    const requestKey = graphId && mode !== 'all' ? `${graphId}:${mode}` : null

    useEffect(() => {
        if (!graphId || mode === 'all') return

        const key = `${graphId}:${mode}`
        if (cache[key]) return

        const controller = new AbortController()

        getGraphMetrics(graphId, mode, controller.signal)
            .then((metrics) => setCache((prev) => ({ ...prev, [key]: { metrics } })))
            .catch((e) => {
                if (!controller.signal.aborted) setCache((prev) => ({ ...prev, [key]: { error: getApiErrorMessage(e) } }))
            })

        return () => controller.abort()
    }, [graphId, mode, cache])

    const current = requestKey ? cache[requestKey] : undefined
    const loading = requestKey !== null && !current
    const error = current?.error
    const metrics = requestKey ? (current?.metrics ?? (loading ? previous : undefined)) : graph?.metrics

    const namesById = useMemo(() => new Map((graph?.nodes ?? []).map((n) => [n.id, n.name])), [graph])
    const ranked = useMemo(() => (metrics ? rankMetrics(metrics) : null), [metrics])
    const table = ranked?.[metricKey]

    const rows = useMemo(() => {
        if (!table) return []
        return [...table.byNode.entries()]
            .map(([id, metric]) => ({ id, rank: metric.rank, display: metricDisplayValue(metricKey, metric) }))
            .sort((a, b) => a.rank - b.rank)
    }, [table, metricKey])

    const bins = useMemo(() => buildBins(metricKey, rows.map((row) => row.display)), [metricKey, rows])
    const description = METRICS.find((m) => m.key === metricKey)?.hint
    const hasLinks = ranked ? [...ranked.degree.byNode.values()].some((metric) => metric.raw > 0) : false
    const allZero = rows.length > 0 && rows.every((row) => row.display === 0)

    function changeMode(next: ModeFilter) {
        setPrevious(metrics)
        setMode(next)
    }

    if (!graph) {
        return <p className="text-sm text-muted-foreground">Постройте граф, чтобы увидеть метрики.</p>
    }

    return (
        <section className="flex flex-col gap-4">
            <SegmentedControl label="Метрика" options={METRIC_OPTIONS} value={metricKey} columns={2} onChange={setMetricKey} />

            <p className="min-h-15 text-sm text-muted-foreground">
                {description}
                {metricKey !== 'degree' && '. Значения показаны как доля от лучшего узла.'}
            </p>

            <SegmentedControl label="Вид транспорта" options={MODE_OPTIONS} value={mode} onChange={changeMode} />

            {error && <p className="text-sm text-destructive" role="alert">{error}</p>}

            {!loading && !error && !metrics && (
                <p className="text-sm text-muted-foreground">Метрики для этого графа недоступны.</p>
            )}

            {!loading && !error && metrics && rows.length === 0 && (
                <p className="text-sm text-muted-foreground">Для этого вида транспорта узлов нет.</p>
            )}

            {!loading && !error && rows.length > 0 && !hasLinks && (
                <p className="text-sm text-muted-foreground">
                    В этом слое {rows.length} узл., но нет ни одной связи, поэтому метрики не имеют смысла.
                </p>
            )}

            {!error && rows.length > 0 && (hasLinks || loading) && (
                <div className={`flex flex-col gap-4 transition-opacity duration-200 ${loading ? 'opacity-50' : ''}`}>
                    <div className="border-t pt-4">
                        <h3 className="mb-3 flex items-center gap-2 text-sm font-semibold">
                            Распределение
                            {loading && <Loader2 className="size-3.5 animate-spin text-muted-foreground" />}
                        </h3>
                        <Histogram metricKey={metricKey} bins={bins} total={rows.length} />
                    </div>

                    <div className="border-t pt-4">
                        {allZero ? (
                            <p className="text-sm text-muted-foreground">
                                У всех узлов значение 0: кратчайшие пути идут напрямую, минуя другие узлы.
                            </p>
                        ) : (
                            <>
                                <h3 className="mb-2 text-sm font-semibold">Топ-{Math.min(TOP_SIZE, rows.length)} узлов</h3>
                                <ol key={`${metricKey}:${mode}`} className="animate-in fade-in duration-300 flex flex-col gap-0.5">
                                    {rows.slice(0, TOP_SIZE).map((row) => (
                                        <li key={row.id}>
                                            <button
                                                type="button"
                                                className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm transition-colors hover:bg-muted"
                                                onClick={() => onSelectNode(row.id)}
                                            >
                                                <span className="w-5 shrink-0 text-muted-foreground tabular-nums">{row.rank}</span>
                                                <span className="min-w-0 flex-1 truncate">{namesById.get(row.id) || 'Без названия'}</span>
                                                <span className="font-medium tabular-nums">{formatMetric(metricKey, row.display)}</span>
                                            </button>
                                        </li>
                                    ))}
                                </ol>
                            </>
                        )}
                    </div>
                </div>
            )}
        </section>
    )
}
