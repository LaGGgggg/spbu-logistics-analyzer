import type { LogisticsGraph } from '@/api/types'

export type NodeMetric = { raw: number; value: number; rank: number }

export type NodeMetrics = {
    degree: NodeMetric
    betweenness: NodeMetric | null
    closeness: NodeMetric | null
    pagerank: NodeMetric
}

export type GraphMetrics = {
    total: number
    byNode: Map<string, NodeMetrics>
}

const MAX_PATH_NODES = 4000
const DAMPING = 0.85
const EPS = 1e-9

type Adjacent = { to: number; w: number }

class MinHeap {
    private keys: number[] = []
    private items: number[] = []

    get size() {
        return this.items.length
    }

    push(key: number, item: number) {
        this.keys.push(key)
        this.items.push(item)
        let i = this.items.length - 1
        while (i > 0) {
            const parent = (i - 1) >> 1
            if (this.keys[parent] <= this.keys[i]) break
            this.swap(i, parent)
            i = parent
        }
    }

    pop(): [number, number] {
        const top: [number, number] = [this.keys[0], this.items[0]]
        const lastKey = this.keys.pop()!
        const lastItem = this.items.pop()!
        if (this.items.length > 0) {
            this.keys[0] = lastKey
            this.items[0] = lastItem
            let i = 0
            for (;;) {
                const l = 2 * i + 1
                const r = l + 1
                let m = i
                if (l < this.items.length && this.keys[l] < this.keys[m]) m = l
                if (r < this.items.length && this.keys[r] < this.keys[m]) m = r
                if (m === i) break
                this.swap(i, m)
                i = m
            }
        }
        return top
    }

    private swap(a: number, b: number) {
        ;[this.keys[a], this.keys[b]] = [this.keys[b], this.keys[a]]
        ;[this.items[a], this.items[b]] = [this.items[b], this.items[a]]
    }
}

function toMetrics(raw: Float64Array): NodeMetric[] {
    let max = 0
    for (const v of raw) if (v > max) max = v

    const order = Array.from(raw.keys()).sort((a, b) => raw[b] - raw[a])
    const ranks = new Array<number>(raw.length)
    order.forEach((idx, pos) => {
        const prev = order[pos - 1]
        ranks[idx] = pos > 0 && Math.abs(raw[prev] - raw[idx]) < EPS ? ranks[prev] : pos + 1
    })

    return Array.from(raw, (v, i) => ({ raw: v, value: max > 0 ? v / max : 0, rank: ranks[i] }))
}

function shortestPaths(adj: Adjacent[][]) {
    const n = adj.length
    const betweenness = new Float64Array(n)
    const closeness = new Float64Array(n)

    for (let s = 0; s < n; s++) {
        const dist = new Float64Array(n).fill(Infinity)
        const sigma = new Float64Array(n)
        const preds: number[][] = Array.from({ length: n }, () => [])
        const stack: number[] = []
        const heap = new MinHeap()

        dist[s] = 0
        sigma[s] = 1
        heap.push(0, s)

        while (heap.size > 0) {
            const [d, v] = heap.pop()
            if (d > dist[v]) continue
            stack.push(v)
            for (const { to, w } of adj[v]) {
                const nd = d + w
                if (nd < dist[to] - EPS) {
                    dist[to] = nd
                    sigma[to] = sigma[v]
                    preds[to] = [v]
                    heap.push(nd, to)
                } else if (Math.abs(nd - dist[to]) <= EPS) {
                    sigma[to] += sigma[v]
                    preds[to].push(v)
                }
            }
        }

        let sum = 0
        for (const v of stack) sum += dist[v]
        const reached = stack.length - 1
        if (sum > 0 && n > 1) closeness[s] = (reached / (n - 1)) * (reached / sum)

        const delta = new Float64Array(n)
        for (let i = stack.length - 1; i >= 0; i--) {
            const w = stack[i]
            for (const v of preds[w]) delta[v] += (sigma[v] / sigma[w]) * (1 + delta[w])
            if (w !== s) betweenness[w] += delta[w]
        }
    }

    return { betweenness, closeness }
}

function pageRank(adj: Adjacent[][]) {
    const n = adj.length
    let rank = new Float64Array(n).fill(1 / n)

    for (let iter = 0; iter < 100; iter++) {
        const next = new Float64Array(n).fill((1 - DAMPING) / n)
        let dangling = 0

        for (let v = 0; v < n; v++) {
            if (adj[v].length === 0) {
                dangling += rank[v]
                continue
            }
            const share = (DAMPING * rank[v]) / adj[v].length
            for (const { to } of adj[v]) next[to] += share
        }

        let diff = 0
        for (let v = 0; v < n; v++) {
            next[v] += (DAMPING * dangling) / n
            diff += Math.abs(next[v] - rank[v])
        }

        rank = next
        if (diff < 1e-10) break
    }

    return rank
}

export function computeGraphMetrics(graph: LogisticsGraph): GraphMetrics {
    const n = graph.nodes.length
    const index = new Map(graph.nodes.map((node, i) => [node.id, i]))
    const adj: Adjacent[][] = Array.from({ length: n }, () => [])
    const neighbors = Array.from({ length: n }, () => new Set<number>())

    for (const edge of graph.edges) {
        const a = index.get(edge.from_node_id)
        const b = index.get(edge.to_node_id)
        if (a === undefined || b === undefined || a === b) continue

        const w = Math.max(edge.distance_km, 1e-6)
        adj[a].push({ to: b, w })
        if (!edge.directed) adj[b].push({ to: a, w })
        neighbors[a].add(b)
        neighbors[b].add(a)
    }

    const degrees = toMetrics(Float64Array.from(neighbors, (set) => set.size))
    const pageranks = toMetrics(pageRank(adj))
    const paths = n <= MAX_PATH_NODES ? shortestPaths(adj) : null
    const betweenness = paths ? toMetrics(paths.betweenness) : null
    const closeness = paths ? toMetrics(paths.closeness) : null

    const byNode = new Map<string, NodeMetrics>()
    graph.nodes.forEach((node, i) => {
        byNode.set(node.id, {
            degree: degrees[i],
            betweenness: betweenness?.[i] ?? null,
            closeness: closeness?.[i] ?? null,
            pagerank: pageranks[i],
        })
    })

    return { total: n, byNode }
}
