import type { LogisticsGraph } from "@/api/types";
import { KeyIndicators } from "./KeyIndicators";
import { TransportStatistics } from "./TransportStatistics";
import { NodeTypeStatistics } from "./NodeTypeStatistics";
import { DataQuality } from "./DataQuality";

type StatisticsPanelProps = {
    graph: LogisticsGraph | null;
};

export function StatisticsPanel({ graph }: StatisticsPanelProps) {
    return (
        <div>
            <KeyIndicators graph={graph} />
            <TransportStatistics edges={graph?.edges ?? null} />
            <NodeTypeStatistics nodes={graph?.nodes ?? null} />
            <DataQuality coverage={graph?.coverage ?? null} />
        </div>
    );
}