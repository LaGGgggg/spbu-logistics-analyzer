import type { LogisticsEdge, LogisticsGraph, LogisticsNode, TransportMode } from '@/api/types'
import { UNKNOWN_KIND } from '@/lib/kinds'
import { uniqueConnections } from '@/lib/connections'

export type DisplayGraph = {
    nodes: LogisticsNode[]
    edges: LogisticsEdge[]
}

export function computeDisplay(
    graph: LogisticsGraph,
    visibleModes: Set<TransportMode>,
    hiddenKinds: Set<string>,
    showEdges: boolean,
): DisplayGraph {
    const nodes = graph.nodes.filter((n) => {
        const modeOk = !visibleModes || n.modes.some((m) => visibleModes.has(m))
        const kindOk = !hiddenKinds || !hiddenKinds.has(n.kind ?? UNKNOWN_KIND)
        return modeOk && kindOk
    })

    if (!showEdges) return { nodes, edges: [] }

    const visibleIds = new Set(nodes.map((n) => n.id))
    const edges = uniqueConnections(graph.edges.filter((e) =>
        (!visibleModes || visibleModes.has(e.mode)) &&
        visibleIds.has(e.from_node_id) &&
        visibleIds.has(e.to_node_id),
    ))

    return { nodes, edges }
}
