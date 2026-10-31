import type { MultiPolygon, Polygon } from 'geojson'
import type { Bounds } from '@/api/types'

type Area = Polygon | MultiPolygon
type Nested = number[] | Nested[]

function eachLon(coords: Nested, visit: (lon: number) => void) {
    if (typeof coords[0] === 'number') visit(coords[0])
    else for (const c of coords as Nested[]) eachLon(c, visit)
}

function shiftWest(coords: Nested): Nested {
    if (typeof coords[0] === 'number') {
        const [lon, lat] = coords as number[]
        return [lon < 0 ? lon + 360 : lon, lat]
    }
    return (coords as Nested[]).map(shiftWest)
}

function lonRange(coords: Nested) {
    let min = Infinity
    let max = -Infinity
    eachLon(coords, (lon) => {
        if (lon < min) min = lon
        if (lon > max) max = lon
    })
    return [min, max]
}

export function normalizeArea(geometry: Area): { geometry: Area; bounds: Bounds } {
    let coords = geometry.coordinates as Nested
    let [west, east] = lonRange(coords)

    if (east - west > 180) {
        coords = shiftWest(coords)
        ;[west, east] = lonRange(coords)
    }

    let south = Infinity
    let north = -Infinity
    const visitLat = (c: Nested) => {
        if (typeof c[0] === 'number') {
            const lat = (c as number[])[1]
            if (lat < south) south = lat
            if (lat > north) north = lat
        } else for (const x of c as Nested[]) visitLat(x)
    }
    visitLat(coords)

    return {
        geometry: { ...geometry, coordinates: coords } as Area,
        bounds: { west, south, east: Math.min(east, 180), north },
    }
}
