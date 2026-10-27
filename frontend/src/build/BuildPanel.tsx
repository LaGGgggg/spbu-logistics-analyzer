import { useState, type FormEvent } from "react";
import { searchPlaces, type Place } from "@/api/geocode";
import type { BBox } from "@/components/map/BBoxDrawer";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Loader2, MousePointerSquareDashed, Network, X } from "lucide-react"

type BuildPanelProps = {
    bbox: BBox | null
    drawMode: boolean
    building: boolean
    error: string | null
    onToggleDraw: () => void
    onClearBBox: () => void
    onBuild: () => void
    onSelectPlace: (p: Place) => void
}
export function BuildPanel({ bbox, drawMode, building, error, onToggleDraw, onClearBBox, onBuild, onSelectPlace}: BuildPanelProps) {
    const [query, setQuery] = useState('')
    const [places, setPlaces] = useState<Place[]>([])
    const [searching, setSearching] = useState(false)
    const [searchError, setSearchError] = useState<string | null>(null)



    async function handleSearch(event: FormEvent) {
        event.preventDefault()
        const clean = query.trim()
        if (!clean) return
        setSearching(true)
        setSearchError(null)
        setPlaces([])

        try {
            const found = await searchPlaces(clean)
            setPlaces(found)
            if (found.length === 0) setSearchError('Ничего не найдено')
        } catch (e) {
            setSearchError(e instanceof Error ? e.message : 'Ошибка поиска')
        } finally {
            setSearching(false)
        }
    }



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

                    <form onSubmit={handleSearch}>
                        <Input
                            value={query}
                            onChange={(e) => setQuery(e.target.value)}
                            disabled={searching}
                            className="h-10" placeholder="Введите регион и нажмите Enter" />
                    </form>

                    {searching && <p className="text-xs text-muted-foreground">Ищем...</p>}
                    {searchError && <p className="text-xs text-destructive" role="alert">{searchError}</p>}

                    {places.map((place) => (
                        <button
                            key={place.id}
                            type="button"
                            className="rounded-md border p-2 text-left text-sm hover:bg-muted"
                            onClick={() => { onSelectPlace(place); setPlaces([]) }}
                        >
                            {place.name}
                        </button>
                    ))}

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
                <Button className="h-12" disabled={!bbox || building} onClick={onBuild}>
                    {building
                        ? <Loader2 data-icon="inline-start" className="animate-spin" />
                        : <Network data-icon="inline-start" />}
                    {building ? 'Строим граф...' : 'Построить граф'}
                </Button>

                {!bbox && (
                    <p className="text-xs text-muted-foreground">Найдите регион или выделите область на карте</p>
                )}

                {error && (
                    <p className="text-sm text-destructive" role="alert">{error}</p>
                )}
            </div>
        </div>
    )
}