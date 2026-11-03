import { useState, type FormEvent } from "react";
import { searchPlaces, type Place } from "@/api/geocode";
import type { BBox } from "@/components/map/BBoxDrawer";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Loader2, MapPin, MousePointerSquareDashed, Network, Search, X } from "lucide-react"

const KM_PER_DEGREE = 111.32

function bboxSizeKm(b: BBox) {
    const midLat = ((b.north + b.south) / 2) * Math.PI / 180
    return {
        width: (b.east - b.west) * KM_PER_DEGREE * Math.cos(midLat),
        height: (b.north - b.south) * KM_PER_DEGREE,
    }
}

const formatKm = (v: number) => Math.round(v).toLocaleString('ru-RU')

function formatRange(from: number, to: number, positive: string, negative: string) {
    const deg = (v: number) => `${Math.abs(v).toLocaleString('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}°`
    if ((from >= 0) === (to >= 0)) return `${deg(from)} – ${deg(to)} ${from >= 0 ? positive : negative}`
    return `${deg(from)} ${from >= 0 ? positive : negative} – ${deg(to)} ${to >= 0 ? positive : negative}`
}

type BuildPanelProps = {
    bbox: BBox | null
    regionName?: string
    drawMode: boolean
    building: boolean
    error: string | null
    onToggleDraw: () => void
    onClearBBox: () => void
    onBuild: () => void
    onCancelBuild: () => void
    onSelectPlace: (p: Place) => void
}
export function BuildPanel({ bbox, regionName, drawMode, building, error, onToggleDraw, onClearBBox, onBuild, onCancelBuild, onSelectPlace}: BuildPanelProps) {
    const [query, setQuery] = useState('')
    const [places, setPlaces] = useState<Place[]>([])
    const [searching, setSearching] = useState(false)
    const [searchError, setSearchError] = useState<string | null>(null)



    async function handleSearch(event: FormEvent) {
        event.preventDefault()
        const clean = query.trim()
        if (!clean || searching) return
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

                    <form onSubmit={handleSearch} className="flex gap-2">
                        <Input
                            value={query}
                            onChange={(e) => setQuery(e.target.value)}
                            className="h-10 flex-1" placeholder="Введите регион" />
                        <Button
                            type="submit"
                            size="icon"
                            variant={searching ? "secondary" : "default"}
                            className="relative size-10 shrink-0 duration-300"
                            aria-busy={searching}
                            aria-label="Найти регион"
                        >
                            <Search className={`absolute transition-all duration-200 ${searching ? 'scale-50 opacity-0' : 'scale-100 opacity-100'}`} />
                            <Loader2 className={`absolute animate-spin transition-opacity duration-200 ${searching ? 'opacity-100' : 'opacity-0'}`} />
                        </Button>
                    </form>

                    {searchError && <p className="animate-in fade-in duration-200 text-xs text-destructive" role="alert">{searchError}</p>}

                    {places.map((place, i) => (
                        <button
                            key={place.id}
                            type="button"
                            style={{ animationDelay: `${i * 40}ms` }}
                            className="animate-in fade-in slide-in-from-top-1 duration-200 [animation-fill-mode:both] rounded-md border p-2 text-left text-sm transition-colors hover:bg-muted"
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
                        <div className="animate-in fade-in slide-in-from-top-1 duration-200 flex flex-col gap-3 pt-2">
                            <div className="flex items-center gap-2">
                                <MapPin className="size-4 shrink-0 text-primary" />
                                <p className="min-w-0 flex-1 truncate font-medium" title={regionName}>
                                    {regionName ?? 'Выделенная область'}
                                </p>
                                <Button variant="ghost" size="icon" className="size-7" aria-label="Сбросить область" onClick={onClearBBox}>
                                    <X />
                                </Button>
                            </div>
                            <dl className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-1.5 text-sm">
                                <dt className="text-muted-foreground">Размер</dt>
                                <dd className="tabular-nums">
                                    {formatKm(bboxSizeKm(bbox).width)} × {formatKm(bboxSizeKm(bbox).height)} км
                                </dd>
                                <dt className="text-muted-foreground">Широта</dt>
                                <dd className="tabular-nums">{formatRange(bbox.south, bbox.north, 'с. ш.', 'ю. ш.')}</dd>
                                <dt className="text-muted-foreground">Долгота</dt>
                                <dd className="tabular-nums">{formatRange(bbox.west, bbox.east, 'в. д.', 'з. д.')}</dd>
                            </dl>
                        </div>
                    )}

                </div>
                <div className="flex gap-2">
                    <Button className="h-12 flex-1" disabled={!bbox || building} onClick={onBuild}>
                        {building
                            ? <Loader2 data-icon="inline-start" className="animate-spin" />
                            : <Network data-icon="inline-start" />}
                        {building ? 'Строим граф...' : 'Построить граф'}
                    </Button>
                    {building && (
                        <Button
                            variant="secondary"
                            className="animate-in fade-in slide-in-from-right-2 duration-200 h-12"
                            onClick={onCancelBuild}
                        >
                            Отменить
                        </Button>
                    )}
                </div>

              

                {error && (
                    <p className="text-sm text-destructive" role="alert">{error}</p>
                )}
            </div>
        </div>
    )
}