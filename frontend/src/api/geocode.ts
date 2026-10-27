import type { Bounds } from "./types";
import type { GeoJsonObject, Polygon } from "geojson";


export type Place = {
    id: number,
    name: string,
    bounds: Bounds,
    geometry: GeoJsonObject
}

type NominatimItem = {
    place_id: number
    display_name: string
    boundingbox: [string, string, string, string]
    geojson?: GeoJsonObject
}

export async function searchPlaces(query: string, signal?: AbortSignal): Promise<Place[]> {
    const url = new URL('https://nominatim.openstreetmap.org/search')
    url.searchParams.set('q', query)
    url.searchParams.set('format', 'jsonv2')
    url.searchParams.set('limit', '5')
    url.searchParams.set('accept-language', 'ru')
    url.searchParams.set('polygon_geojson', '1')
    url.searchParams.set('polygon_threshold', '0.01')

    const response = await fetch(url, { signal })
    if (!response.ok) throw new Error(`Поиск недоступен (HTTP ${response.status})`)

    const items = (await response.json()) as NominatimItem[]
    return items.map((item) => {
        const [south, north, west, east] = item.boundingbox.map(Number)
        const isArea = item.geojson?.type === 'Polygon' || item.geojson?.type === 'MultiPolygon'

        const rect: Polygon = {
            type: 'Polygon',
            coordinates: [[[west, south], [east, south], [east, north], [west, north], [west, south]]],
        }
        const geometry: GeoJsonObject = isArea ? item.geojson! : rect


        return { id: item.place_id, name: item.display_name, bounds: { west, south, east, north }, geometry }
    })
}

