import type { LogisticsNode } from "@/api/types";
import { kindLabel, UNKNOWN_KIND } from "@/lib/kinds";

type NodeTypeStatisticsProps = {
    nodes: LogisticsNode[] | null;
};

export function NodeTypeStatistics({ nodes }: NodeTypeStatisticsProps) {
    const kindCounts = new Map<string, number>();

    for (const node of nodes ?? []) {
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
        <div className="mt-6 border-t pt-4">
            <h3 className="mb-3 text-sm font-semibold">
                Узлы по типам
            </h3>

            <div className="space-y-3">
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
            </div>

            {nodes === null && (
                <p className="text-sm text-muted-foreground">
                    Сначала постройте граф.
                </p>
            )}
        </div>
    );
}