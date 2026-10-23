import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { MousePointerSquareDashed, Network } from "lucide-react"

export function BuildPanel() {
    return (
        <div>
            <div className="flex flex-col gap-1 border-b px-4 py-4">
                <h2 className="text-lg font-semibold">Построение графа</h2>
                <p className="text-sm text-muted-foreground">
                    Выберите регион или задайте границы области анализа.
                </p>
            </div>

            <div className="flex flex-col p-4 gap-4">
                <div className="flex flex-col gap-2">
                    <Input className="h-10" placeholder="Введите регион" />


                    <Button
                        variant={"secondary"}
                        className={'h-10'}
                    >
                        <MousePointerSquareDashed data-icon="inline-start" />
                        {'или выделите область на карте'}
                    </Button>

                </div>
                <Button className="h-12">
                    <Network data-icon="inline-start" />
                    Построить граф
                </Button>
            </div>
        </div>
    )
}