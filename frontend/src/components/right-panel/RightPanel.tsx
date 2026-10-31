import { Button } from "@/components/ui/button";
import { StatisticsPanel } from "./StatisticsPanel";
import { NodePanel } from "./NodePanel";
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
            <div className="flex gap-1 rounded-lg bg-muted p-1">
                {TABS.map((tab) => (
                    <Button
                        key={tab.value}
                        variant={activeTab === tab.value ? "secondary" : "ghost"}
                        className={`flex-1 ${
                            activeTab === tab.value
                                ? "bg-card text-foreground shadow-sm hover:bg-card"
                                : "text-muted-foreground hover:bg-transparent hover:text-foreground"
                        }`}
                        onClick={() => onTabChange(tab.value)}
                    >
                        {tab.label}
                    </Button>
                ))}
            </div>

            <div className="mt-4">
                {activeTab === "statistics" && <StatisticsPanel graph={graph} />}
                {activeTab === "metrics" && <p>Блок 2</p>}
                {activeTab === "node" && (
                    <NodePanel graph={graph} nodeId={selectedNodeId} onSelectNode={onSelectNode} />
                )}
            </div>
        </div>
    );
}
