import type { LogisticsEdge } from "@/api/types";
import { MODES, MODE_COLOR, MODE_LABEL } from "@/lib/modes";

const transportTypes = MODES.map((mode) => ({
    mode,
    label: MODE_LABEL[mode],
    color: MODE_COLOR[mode],
}));

type TransportStatisticsProps = {
    edges: LogisticsEdge[] | null;
};

export function TransportStatistics({ edges }: TransportStatisticsProps) {
    const transportStats = transportTypes.map((transport) => {
        const transportEdges = edges?.filter(
            (edge) => edge.mode === transport.mode
        ) ?? [];

        const totalLength = transportEdges.reduce(
            (sum, edge) => sum + edge.distance_km,
            0
        );

        return {
            ...transport,
            edgeCount: transportEdges.length,
            totalLength,
        };
    });

    const transportEdgeCount = transportStats.reduce(
        (sum, transport) => sum + transport.edgeCount,
        0
    );

    return (
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
                                {edges !== null
                                    ? `${transport.edgeCount.toLocaleString("ru-RU")} · ${transport.totalLength.toLocaleString("ru-RU", {
                                        maximumFractionDigits: 1,
                                    })
                                    } км`
                                    : "—"}
                            </span>
                        </div>

                        <div className="mt-1 h-2 overflow-hidden rounded-full bg-muted">
                            <div
                                className="h-full rounded-full"
                                style={{
                                    backgroundColor: transport.color,
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
    );
}