import { Button } from "@/components/ui/button";
import { useState } from "react";

export function RightPanel() {
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
                {activeTab === "statistics" && <p>Блок 1</p>}
                {activeTab === "metrics" && <p>Блок 2</p>}
                {activeTab === "node" && <p>Блок 3</p>}
            </div>
        </div>
    );
}