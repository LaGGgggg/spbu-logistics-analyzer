import type { GraphMetrics } from '@/api/types'

export type MetricKey = 'degree' | 'betweenness' | 'closeness' | 'pagerank'

export type NodeMetric = { raw: number; value: number; rank: number }

export type MetricTable = { total: number; byNode: Map<string, NodeMetric> }

export type RankedMetrics = Record<MetricKey, MetricTable>

export const METRICS: { key: MetricKey; label: string; hint: string }[] = [
    { key: 'degree', label: 'Степень', hint: 'Сколько узлов связано с этим напрямую' },
    { key: 'betweenness', label: 'Посредничество', hint: 'Как часто узел лежит на кратчайшем пути между другими узлами' },
    { key: 'closeness', label: 'Близость', hint: 'Насколько близко узел в среднем ко всем остальным по длине путей' },
    { key: 'pagerank', label: 'PageRank', hint: 'Важность узла с учётом важности его соседей' },
]

const EPS = 1e-12

function rankTable(entries: [string, number][]): MetricTable {
    const sorted = [...entries].sort((a, b) => b[1] - a[1])
    const max = sorted[0]?.[1] ?? 0
    const byNode = new Map<string, NodeMetric>()
    let rank = 1

    sorted.forEach(([id, raw], position) => {
        if (position > 0 && sorted[position - 1][1] - raw > EPS) rank = position + 1
        byNode.set(id, { raw, value: max > 0 ? raw / max : 0, rank })
    })

    return { total: sorted.length, byNode }
}

export function rankMetrics(metrics: GraphMetrics): RankedMetrics {
    return {
        degree: rankTable(metrics.power_load.map((item) => [item.node_id, item.degree])),
        betweenness: rankTable(Object.entries(metrics.betweenness)),
        closeness: rankTable(Object.entries(metrics.closeness)),
        pagerank: rankTable(Object.entries(metrics.page_rank)),
    }
}

export function metricDisplayValue(key: MetricKey, metric: NodeMetric) {
    return key === 'degree' ? metric.raw : metric.value
}

export function formatMetric(key: MetricKey, value: number) {
    if (key === 'degree') return value.toLocaleString('ru-RU')
    return value.toLocaleString('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}
