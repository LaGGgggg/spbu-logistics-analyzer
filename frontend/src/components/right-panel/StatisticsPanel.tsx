import type { LogisticsGraph } from "@/api/types";

const transportTypes = [
    { mode: "rail", label: "Ж/д", color: "bg-slate-700" },
    { mode: "sea", label: "Море", color: "bg-blue-700" },
    { mode: "river", label: "Река", color: "bg-sky-500" },
];

type StatisticsPanelProps = {
    graph: LogisticsGraph | null;
};

export function StatisticsPanel({ graph }: StatisticsPanelProps) {
    const totalLength = graph
        ? graph.edges.reduce((sum, edge) => sum + edge.distance_km, 0)
        : null;

    const averageLength =
        graph && graph.edges.length > 0 && totalLength !== null
        ? totalLength / graph.edges.length
        : null;

    const connectedPairs = new Set<string>();

    for (const edge of graph?.edges ?? []) {
        if (edge.from_node_id === edge.to_node_id) continue;

        connectedPairs.add(
            JSON.stringify([edge.from_node_id, edge.to_node_id])
        );

        if (!edge.directed) {
            connectedPairs.add(
                JSON.stringify([edge.to_node_id, edge.from_node_id])
            );
        }
    }

    const nodeCount = graph?.nodes.length ?? 0;

    const density = graph && nodeCount > 1
        ? connectedPairs.size / (nodeCount * (nodeCount - 1))
        : null;

    const neighbors = new Map<string, string[]>();

    for (const node of graph?.nodes ?? []) {
        neighbors.set(node.id, []);
    }

    for (const edge of graph?.edges ?? []) {
        neighbors.get(edge.from_node_id)?.push(edge.to_node_id);
        neighbors.get(edge.to_node_id)?.push(edge.from_node_id);
    }


    const visited = new Set<string>();
    let componentCount = 0;

    for (const nodeId of neighbors.keys()) {
        if (visited.has(nodeId)) continue;

        componentCount++;

        const stack = [nodeId];
        visited.add(nodeId);

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
                        <p className="mt-1 font-mono text-lg font-semibold">
                            {graph ? graph.nodes.length.toLocaleString("ru-RU") : "—"}
                        </p>
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
                        Компонент связности
                    </p>

                    <p className="mt-1 font-mono text-lg font-semibold">
                        {graph ? componentCount.toLocaleString("ru-RU") : "—"}
                    </p>
                </div>

            </div>


        </section>
    );
}