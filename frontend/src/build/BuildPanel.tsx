import type { BBox } from "@/components/map/BBoxDrawer";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { MousePointerSquareDashed, Network, X } from "lucide-react"

type BuildPanelProps = {
    bbox: BBox | null
    drawMode: boolean
    onToggleDraw: () => void
    onClearBBox: () => void
}

export function BuildPanel({ bbox, drawMode, onToggleDraw, onClearBBox }: BuildPanelProps) {

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
                        variant={drawMode ? "default" : "secondary"}
                        className="h-10"
                        onClick={onToggleDraw}
                    >
                        <MousePointerSquareDashed data-icon="inline-start" />
                        {drawMode ? 'Отменить выделение' : 'или выделите область на карте'}
                    </Button>

                    {bbox && (
                        <div className="flex items-start justify-between gap-2 rounded-md border bg-muted/50 p-3">
                            <dl className="grid grid-cols-[auto_1fr] gap-x-3 gap-y-1 font-mono text-xs tabular-nums">
                                <dt className="text-muted-foreground">С</dt><dd>{bbox.north}</dd>
                                <dt className="text-muted-foreground">Ю</dt><dd>{bbox.south}</dd>
                                <dt className="text-muted-foreground">З</dt><dd>{bbox.west}</dd>
                                <dt className="text-muted-foreground">В</dt><dd>{bbox.east}</dd>
                            </dl>
                            <Button variant="ghost" size="icon" aria-label="Сбросить область" onClick={onClearBBox}>
                                <X />
                            </Button>
                        </div>
                    )}

                </div>
                <Button className="h-12">
                    <Network data-icon="inline-start" />
                    Построить граф
                </Button>
            </div>
        </div>
    )
}