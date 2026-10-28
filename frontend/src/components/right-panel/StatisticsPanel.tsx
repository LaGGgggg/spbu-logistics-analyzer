import type { LogisticsGraph } from "@/api/types";
import { kindLabel, UNKNOWN_KIND } from "@/lib/kinds";

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

    const transportStats = transportTypes.map((transport) => {
        const edges = graph?.edges.filter(
            (edge) => edge.mode === transport.mode
        ) ?? [];

        const totalLength = edges.reduce(
            (sum, edge) => sum + edge.distance_km,
            0
        );

        return {
            ...transport,
            edgeCount: edges.length,
            totalLength,
        };
    });

    const transportEdgeCount = transportStats.reduce(
        (sum, transport) => sum + transport.edgeCount,
        0
    );

    const kindCounts = new Map<string, number>();

    for (const node of graph?.nodes ?? []) {
        const kind = node.kind || UNKNOWN_KIND;

        kindCounts.set(
            kind,
            (kindCounts.get(kind) ?? 0) + 1
        );
    }

    const nodeTypeStats = Array.from(kindCounts, ([kind, count]) => ({
        kind,
        count,
    })).sort((a, b) => b.count - a.count);

    const maxKindCount = nodeTypeStats[0]?.count ?? 0;

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
                        Компонент связности
                    </p>

                    <p className="mt-1 font-mono text-lg font-semibold">
                        {graph ? componentCount.toLocaleString("ru-RU") : "—"}
                    </p>
                </div>

            </div>

            <div className="mt-6 border-t pt-4">

                <div className="mb-3 flex items-center justify-between">
                    <h3 className="text-sm font-semibold">
                        По видам транспорта
                    </h3>

                    <span className="text-xs text-muted-foreground">
                        рёбер · км
                    </span>
                </div>

                <div className="space-y-3">
                    {transportStats.map((transport) => (
                        <div key={transport.mode}>
                            <div className="flex items-center justify-between gap-2">
                                <span className="text-sm">
                                    {transport.label}
                                </span>

                                <span className="font-mono text-xs text-muted-foreground">
                                    {graph
                                        ? `${transport.edgeCount.toLocaleString("ru-RU")} · ${transport.totalLength.toLocaleString("ru-RU", {
                                            maximumFractionDigits: 1,
                                        })
                                        } км`
                                        : "—"}
                                </span>
                            </div>

                            <div className="mt-1 h-2 overflow-hidden rounded-full bg-muted">
                                <div
                                    className={`h-full rounded-full ${transport.color}`}
                                    style={{
                                        width: `${transportEdgeCount > 0
                                            ? (transport.edgeCount / transportEdgeCount) * 100
                                            : 0
                                            }%`,
                                    }}
                                />
                            </div>
                        </div>
                    ))}
                </div>
            </div>

            <div className="mt-6 border-t pt-4">
                <h3 className="mb-3 text-sm font-semibold">
                    Узлы по типам
                </h3>

                {nodeTypeStats.map((item) => (
                    <div
                        key={item.kind}
                        className="grid grid-cols-[100px_1fr_36px] items-center gap-2"
                    >
                        <span className="text-xs">
                            {kindLabel(item.kind)}
                        </span>

                        <div className="h-2 overflow-hidden rounded-full bg-muted">
                            <div
                                className="h-full rounded-full bg-slate-600"
                                style={{
                                    width: `${maxKindCount > 0
                                        ? (item.count / maxKindCount) * 100
                                        : 0
                                        }%`,
                                }}
                            />
                        </div>

                        <span className="text-right font-mono text-xs">
                            {item.count.toLocaleString("ru-RU")}
                        </span>
                    </div>
                ))}

                {!graph && (
                    <p className="text-sm text-muted-foreground">
                        Сначала постройте граф.
                    </p>
                )}
            </div>

            <div className="mt-6 border-t pt-4">
                <h3 className="mb-3 text-sm font-semibold">
                    Качество данных
                </h3>

                <div className="grid grid-cols-2 gap-x-4 gap-y-2 text-xs">
                    <div className="flex justify-between gap-2">
                        <span className="text-muted-foreground">Кандидатов</span>
                        <span className="font-mono">
                            {graph
                                ? graph.coverage.candidates.toLocaleString("ru-RU")
                                : "—"}
                        </span>
                    </div>

                    <div className="flex justify-between gap-2">
                        <span className="text-muted-foreground">Принято</span>
                        <span className="font-mono">
                            {graph
                                ? graph.coverage.accepted.toLocaleString("ru-RU")
                                : "—"}
                        </span>
                    </div>

                    <div className="flex justify-between gap-2">
                        <span className="text-muted-foreground">Отклонено</span>
                        <span className="font-mono">
                            {graph
                                ? graph.coverage.rejected.toLocaleString("ru-RU")
                                : "—"}
                        </span>
                    </div>

                    <div className="flex justify-between gap-2">
                        <span className="text-muted-foreground">Предупреждений</span>
                        <span className="font-mono">
                            {graph
                                ? (graph.coverage.warnings?.length ?? 0).toLocaleString("ru-RU")
                                : "—"}
                        </span>
                    </div>
                </div>

                {(graph?.coverage.warnings?.length ?? 0) > 0 && (
                    <div className="mt-3 rounded-md border border-amber-200 bg-amber-50 p-3 text-amber-950">
                        <p className="mb-2 text-xs font-semibold">
                            Предупреждения источника
                        </p>

                        <ul className="list-disc space-y-1 pl-4 text-xs">
                            {(graph?.coverage.warnings ?? []).map((warning, index) => (
                                <li key={`${index}-${warning}`}>
                                    {warning}
                                </li>
                            ))}
                        </ul>
                    </div>
                )}
            </div>

        </section>
    );
}