import type { LogisticsEdge } from "@/api/types";

export function uniqueConnections(
    edges: LogisticsEdge[]
): LogisticsEdge[] {
    const connections = new Map<string, LogisticsEdge>();

    for (const edge of edges) {
        const [from, to] = [
            edge.from_node_id,
            edge.to_node_id,
        ].sort();

        const key = JSON.stringify([edge.mode, from, to]);
        const existing = connections.get(key);

        if (!existing || edge.distance_km < existing.distance_km) {
            connections.set(key, edge);
        }
    }

    return Array.from(connections.values());
}