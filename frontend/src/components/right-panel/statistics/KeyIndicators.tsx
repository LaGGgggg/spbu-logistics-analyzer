import { useMemo } from "react";
import type { LogisticsGraph } from "@/api/types";

function computeTopology(graph: LogisticsGraph) {
    const pairs = new Set<string>();
    const neighbors = new Map<string, string[]>();

    for (const node of graph.nodes) neighbors.set(node.id, []);

    for (const edge of graph.edges) {
        neighbors.get(edge.from_node_id)?.push(edge.to_node_id);
        neighbors.get(edge.to_node_id)?.push(edge.from_node_id);

        if (edge.from_node_id === edge.to_node_id) continue;

        pairs.add(`${edge.from_node_id}|${edge.to_node_id}`);
        if (!edge.directed) pairs.add(`${edge.to_node_id}|${edge.from_node_id}`);
    }

    const nodeCount = graph.nodes.length;
    const density = nodeCount > 1 ? pairs.size / (nodeCount * (nodeCount - 1)) : null;

    const visited = new Set<string>();
    let componentCount = 0;

    for (const startId of neighbors.keys()) {
        if (visited.has(startId)) continue;

        componentCount++;
        const stack = [startId];
        visited.add(startId);

        while (stack.length > 0) {
            const currentId = stack.pop()!;
            for (const neighborId of neighbors.get(currentId) ?? []) {
                if (!visited.has(neighborId)) {
                    visited.add(neighborId);
                    stack.push(neighborId);
                }
            }
        }
    }

    return { density, componentCount };
}

type KeyIndicatorsProps = {
    graph: LogisticsGraph | null;
};

export function KeyIndicators({ graph }: KeyIndicatorsProps) {
    const totalLength = graph
        ? graph.edges.reduce((sum, edge) => sum + edge.distance_km, 0)
        : null;

    const averageLength =
        graph && graph.edges.length > 0 && totalLength !== null
            ? totalLength / graph.edges.length
            : null;

    const topology = useMemo(
        () => (graph ? computeTopology(graph) : null),
        [graph],
    );
    const density = topology?.density ?? null;
    const componentCount = topology?.componentCount ?? 0;

    return (
        <section>
            <h2 className="text-sm font-semibold">
                Ключевые показатели
            </h2>

            <div className="mt-3 grid grid-cols-2 gap-2">
                <div className="rounded-md border bg-muted/50 p-3">
                    <p className="text-xs text-muted-foreground">
                        Узлов
                    </p>

                    <p className="mt-1 font-mono text-lg font-semibold">
                        {graph ? graph.nodes.length.toLocaleString("ru-RU") : "—"}
                    </p>
                </div>

                <div className="rounded-md border bg-muted/50 p-3">
                    <p className="text-xs text-muted-foreground">
                        Рёбер
                    </p>

                    <p className="mt-1 font-mono text-lg font-semibold">
                        {graph ? graph.edges.length.toLocaleString("ru-RU") : "—"}
                    </p>
                </div>

                <div className="rounded-md border bg-muted/50 p-3">
                    <p className="text-xs text-muted-foreground">
                        Сумма длин рёбер
                    </p>

                    <p className="mt-1 font-mono text-lg font-semibold">
                        {totalLength !== null
                            ? totalLength.toLocaleString("ru-RU", {
                                maximumFractionDigits: 1,
                            })
                            : "—"}

                        {totalLength !== null && (
                            <span className="ml-1 text-xs font-normal">км</span>
                        )}
                    </p>
                </div>

                <div className="rounded-md border bg-muted/50 p-3">
                    <p className="text-xs text-muted-foreground">
                        Средняя длина ребра
                    </p>

                    <p className="mt-1 font-mono text-lg font-semibold">
                        {averageLength !== null
                            ? averageLength.toLocaleString("ru-RU", {
                                maximumFractionDigits: 1,
                            })
                            : "—"}

                        {averageLength !== null && (
                            <span className="ml-1 text-xs font-normal">км</span>
                        )}
                    </p>
                </div>

                <div className="rounded-md border bg-muted/50 p-3">
                    <p className="text-xs text-muted-foreground">
                        Плотность графа
                    </p>

                    <p className="mt-1 font-mono text-lg font-semibold">
                        {density !== null
                            ? density.toLocaleString("ru-RU", {
                                maximumFractionDigits: 4,
                            })
                            : "—"}
                    </p>
                </div>

                <div className="rounded-md border bg-muted/50 p-3">
                    <p className="text-xs text-muted-foreground">
                        Компонентов связности
                    </p>

                    <p className="mt-1 font-mono text-lg font-semibold">
                        {graph ? componentCount.toLocaleString("ru-RU") : "—"}
                    </p>
                </div>

            </div>
        </section>
    );
}