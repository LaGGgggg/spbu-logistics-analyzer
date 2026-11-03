import { SegmentedControl } from "@/components/ui/segmented-control";
import { StatisticsPanel } from "./statistics/StatisticsPanel";
import { NodePanel } from "./NodePanel";
import { MetricsPanel } from "./MetricsPanel";
import type { LogisticsGraph } from "@/api/types";

export type RightTab = "statistics" | "metrics" | "node";

const TABS: { value: RightTab; label: string }[] = [
    { value: "statistics", label: "Статистика" },
    { value: "metrics", label: "Метрики" },
    { value: "node", label: "Узел" },
];

type RightPanelProps = {
    graph: LogisticsGraph | null;
    activeTab: RightTab;
    onTabChange: (tab: RightTab) => void;
    selectedNodeId: string | null;
    onSelectNode: (id: string) => void;
};

export function RightPanel({ graph, activeTab, onTabChange, selectedNodeId, onSelectNode }: RightPanelProps) {
    return (
        <div className="p-4">
            <SegmentedControl label="Раздел панели" options={TABS} value={activeTab} onChange={onTabChange} />

            <div key={activeTab} className="animate-in fade-in duration-200 mt-4">
                {activeTab === "statistics" && <StatisticsPanel graph={graph} />}
                {activeTab === "metrics" && <MetricsPanel graph={graph} onSelectNode={onSelectNode} />}
                {activeTab === "node" && (
                    <NodePanel graph={graph} nodeId={selectedNodeId} onSelectNode={onSelectNode} />
                )}
            </div>
        </div>
    );
}
