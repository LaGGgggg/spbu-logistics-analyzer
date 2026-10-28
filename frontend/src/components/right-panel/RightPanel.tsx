import { Button } from "@/components/ui/button";
import { useState } from "react";
import { StatisticsPanel } from "./StatisticsPanel";
import type { LogisticsGraph } from "@/api/types";

type RightPanelProps = {
    graph: LogisticsGraph | null;
};

export function RightPanel({ graph }: RightPanelProps) {
    const [activeTab, setActiveTab] = useState("statistics");

    return (
        <div className="p-4">
            <div className="flex gap-1 rounded-lg bg-muted p-1">
                <Button
                    variant={activeTab === "statistics" ? "secondary" : "ghost"}
                    className={`flex-1 ${
                        activeTab === "statistics"
                            ? "bg-zinc-700 text-white hover:bg-zinc-700 hover:text-white"
                            : "hover:bg-transparent hover:text-inherit"
                    }`}
                    onClick={() => setActiveTab("statistics")}
                >
                    Статистика
                </Button>

                <Button
                    variant={activeTab === "metrics" ? "secondary" : "ghost"}
                    className={`flex-1 ${
                        activeTab === "metrics"
                            ? "bg-zinc-700 text-white hover:bg-zinc-700 hover:text-white"
                            : "hover:bg-transparent hover:text-inherit"
                    }`}
                    onClick={() => setActiveTab("metrics")}
                >
                    Метрики
                </Button>

                <Button
                    variant={activeTab === "node" ? "secondary" : "ghost"}
                    className={`flex-1 ${
                        activeTab === "node"
                            ? "bg-zinc-700 text-white hover:bg-zinc-700 hover:text-white"
                            : "hover:bg-transparent hover:text-inherit"
                    }`}
                    onClick={() => setActiveTab("node")}
                >
                    Узел
                </Button>
            </div>

            <div className="mt-4">
                {activeTab === "statistics" && <StatisticsPanel graph={graph} />}
                {activeTab === "metrics" && <p>Блок 2</p>}
                {activeTab === "node" && <p>Блок 3</p>}
            </div>
        </div>
    );
}