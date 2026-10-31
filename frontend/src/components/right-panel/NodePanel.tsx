import { useMemo } from 'react'
import type { LogisticsEdge, LogisticsGraph, LogisticsNode } from '@/api/types'
import { MODE_COLOR, MODE_LABEL } from '@/lib/modes'
import { kindLabel, UNKNOWN_KIND } from '@/lib/kinds'
import { computeGraphMetrics, type NodeMetric } from '@/lib/metrics'

type NodePanelProps = {
    graph: LogisticsGraph | null
    nodeId: string | null
    onSelectNode: (id: string) => void
}

type Neighbor = { edge: LogisticsEdge; other: LogisticsNode }

const ID_LABEL: Record<string, string> = {
    esr: 'Код ЕСР',
    unlocode: 'UN/LOCODE',
    iata: 'IATA',
    icao: 'ICAO',
    osm_node: 'OSM node',
    osm_way: 'OSM way',
    osm_relation: 'OSM relation',
}

const ALWAYS_SHOWN_IDS = ['esr', 'unlocode']

const formatKm = (v: number) => v.toLocaleString('ru-RU', { maximumFractionDigits: 1 })
const formatCoord = (v: number) => v.toLocaleString('ru-RU', { minimumFractionDigits: 5, maximumFractionDigits: 5 })
const formatScore = (v: number) => v.toLocaleString('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 })

type MetricRowProps = {
    label: string
    hint: string
    metric: NodeMetric | null
    total: number
    format?: (metric: NodeMetric) => string
}

function MetricRow({ label, hint, metric, total, format = (m) => formatScore(m.value) }: MetricRowProps) {
    return (
        <div className="flex flex-col gap-1.5">
            <div className="flex items-baseline justify-between gap-2 text-sm">
                <span className="cursor-help underline decoration-dotted decoration-muted-foreground/50 underline-offset-4" title={hint}>
                    {label}
                </span>
                {metric ? (
                    <span className="tabular-nums">
                        <span className="font-semibold">{format(metric)}</span>
                        <span className="text-muted-foreground"> · {metric.rank} место из {total}</span>
                    </span>
                ) : (
                    <span className="text-muted-foreground">граф слишком большой</span>
                )}
            </div>
            <div className="h-1.5 overflow-hidden rounded-full bg-muted">
                <div
                    className="h-full rounded-full bg-primary transition-[width] duration-500"
                    style={{ width: `${(metric?.value ?? 0) * 100}%` }}
                />
            </div>
        </div>
    )
}

export function NodePanel({ graph, nodeId, onSelectNode }: NodePanelProps) {
    const nodesById = useMemo(
        () => new Map((graph?.nodes ?? []).map((n) => [n.id, n])),
        [graph],
    )
    const metrics = useMemo(() => (graph ? computeGraphMetrics(graph) : null), [graph])
    const node = nodeId ? nodesById.get(nodeId) : undefined
    const nodeMetrics = nodeId ? metrics?.byNode.get(nodeId) : undefined

    const neighbors = useMemo(() => {
        if (!graph || !nodeId) return []
        const nearest = new Map<string, Neighbor>()
        for (const edge of graph.edges) {
            if (edge.from_node_id !== nodeId && edge.to_node_id !== nodeId) continue
            const otherId = edge.from_node_id === nodeId ? edge.to_node_id : edge.from_node_id
            const other = otherId === nodeId ? undefined : nodesById.get(otherId)
            if (!other) continue
            const current = nearest.get(otherId)
            if (!current || edge.distance_km < current.edge.distance_km) nearest.set(otherId, { edge, other })
        }
        return [...nearest.values()].sort((a, b) => a.edge.distance_km - b.edge.distance_km)
    }, [graph, nodeId, nodesById])

    if (!node) {
        return (
            <p className="text-sm text-muted-foreground">
                Нажмите на узел на карте, чтобы увидеть подробности.
            </p>
        )
    }

    const ids = node.external_ids ?? []
    const extraIds = ids.filter((ext) => !ALWAYS_SHOWN_IDS.includes(ext.namespace))

    return (
        <section key={node.id} className="animate-in fade-in duration-200 flex flex-col gap-4">
            <div>
                <h2 className="text-base font-semibold leading-tight">{node.name || 'Без названия'}</h2>
                <p className="mt-1 text-sm text-muted-foreground">{kindLabel(node.kind ?? UNKNOWN_KIND)}</p>
            </div>

            <div className="flex flex-wrap gap-1.5">
                {node.modes.map((m) => (
                    <span key={m} className="inline-flex items-center gap-1.5 rounded-full bg-muted px-2.5 py-1 text-xs">
                        <span className="size-2 rounded-full" style={{ backgroundColor: MODE_COLOR[m] }} />
                        {MODE_LABEL[m] ?? m}
                    </span>
                ))}
            </div>

            <dl className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-2 border-t pt-4 text-sm">
                <dt className="text-muted-foreground">Широта</dt>
                <dd className="tabular-nums">{formatCoord(node.coordinates.latitude)}</dd>

                <dt className="text-muted-foreground">Долгота</dt>
                <dd className="tabular-nums">{formatCoord(node.coordinates.longitude)}</dd>

                {node.cargo && (
                    <>
                        <dt className="text-muted-foreground">Груз</dt>
                        <dd>{node.cargo}</dd>
                    </>
                )}

                {ALWAYS_SHOWN_IDS.map((ns) => (
                    <div key={ns} className="contents">
                        <dt className="text-muted-foreground">{ID_LABEL[ns]}</dt>
                        <dd className="break-all tabular-nums">
                            {ids.filter((ext) => ext.namespace === ns).map((ext) => ext.value).join(', ') || '—'}
                        </dd>
                    </div>
                ))}

                {extraIds.map((ext) => (
                    <div key={`${ext.namespace}:${ext.value}`} className="contents">
                        <dt className="text-muted-foreground">{ID_LABEL[ext.namespace] ?? ext.namespace}</dt>
                        <dd className="break-all tabular-nums">{ext.value}</dd>
                    </div>
                ))}
            </dl>

            {metrics && nodeMetrics && (
                <div className="flex flex-col gap-3 border-t pt-4">
                    <div className="flex items-baseline justify-between">
                        <h3 className="text-sm font-semibold">Метрики узла</h3>
                        <span className="text-xs text-muted-foreground">весь граф</span>
                    </div>
                    <MetricRow
                        label="Степень"
                        hint="Сколько узлов связано с этим напрямую"
                        metric={nodeMetrics.degree}
                        total={metrics.total}
                        format={(m) => String(m.raw)}
                    />
                    <MetricRow
                        label="Посредничество"
                        hint="Как часто узел лежит на кратчайшем пути между другими узлами — транзитная роль"
                        metric={nodeMetrics.betweenness}
                        total={metrics.total}
                    />
                    <MetricRow
                        label="Близость"
                        hint="Насколько близко узел в среднем ко всем остальным по длине путей в км"
                        metric={nodeMetrics.closeness}
                        total={metrics.total}
                    />
                    <MetricRow
                        label="PageRank"
                        hint="Важность узла с учётом важности его соседей"
                        metric={nodeMetrics.pagerank}
                        total={metrics.total}
                    />
                </div>
            )}

            <div className="border-t pt-4">
                <h3 className="mb-2 text-sm font-semibold">
                    Соседние узлы <span className="font-normal text-muted-foreground">· {neighbors.length}</span>
                </h3>

                {neighbors.length === 0 && (
                    <p className="text-sm text-muted-foreground">Изолированный узел, рёбер нет.</p>
                )}

                <ul className="flex flex-col gap-0.5">
                    {neighbors.map(({ edge, other }) => (
                        <li key={edge.id}>
                            <button
                                type="button"
                                onClick={() => onSelectNode(other.id)}
                                className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm transition-colors hover:bg-muted"
                            >
                                <span className="size-2 shrink-0 rounded-full" style={{ backgroundColor: MODE_COLOR[edge.mode] }} />
                                <span className="min-w-0 flex-1 truncate">{other.name || 'Без названия'}</span>
                                <span className="text-muted-foreground tabular-nums">{formatKm(edge.distance_km)} км</span>
                            </button>
                        </li>
                    ))}
                </ul>
            </div>
        </section>
    )
}
