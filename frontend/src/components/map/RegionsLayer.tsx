import { useEffect, useRef, useState } from 'react'
import { GeoJSON } from 'react-leaflet'
import type { FeatureCollection, MultiPolygon, Polygon as GeoPolygon } from 'geojson'
import type { GeoJSON as LeafletGeoJSON, Layer, Polygon, PathOptions } from 'leaflet'
import type { Place } from '@/api/geocode'
import { normalizeArea } from '@/lib/geo'

const BASE: PathOptions = { color: '#64748b', weight: 0.8, fillOpacity: 0, opacity: 0.5 }
const HOVER: PathOptions = { color: '#2563eb', weight: 1.5, fillColor: '#2563eb', fillOpacity: 0.05, opacity: 0.5 }

type RegionsLayerProps = {
    active: boolean
    hoverFill?: boolean
    lockedId?: number | null
    onSelect: (place: Place) => void
}

export function RegionsLayer({ active, hoverFill = true, lockedId = null, onSelect }: RegionsLayerProps) {
    const [data, setData] = useState<FeatureCollection | null>(null)
    const layerRef = useRef<LeafletGeoJSON | null>(null)
    const activeRef = useRef(active)
    const onSelectRef = useRef(onSelect)
    const lockedRef = useRef(lockedId)
    const hoverFillRef = useRef(hoverFill)
    const ignoreClicksUntilRef = useRef(0)

    useEffect(() => {
        if (!activeRef.current && active) ignoreClicksUntilRef.current = Date.now() + 400
        activeRef.current = active
        onSelectRef.current = onSelect
        lockedRef.current = lockedId
        hoverFillRef.current = hoverFill
    }, [active, hoverFill, lockedId, onSelect])

    useEffect(() => {
        const controller = new AbortController()
        fetch('/regions.geojson', { signal: controller.signal })
            .then((r) => r.json() as Promise<FeatureCollection>)
            .then(setData)
            .catch(() => { })
        return () => controller.abort()
    }, [])


    if (!data) return null

    return (
        <GeoJSON
            ref={layerRef}
            data={data}
            style={BASE}
            onEachFeature={(feature, layer: Layer) => {
                const props = feature.properties ?? {}
                const name: string = props.name_ru ?? props.name ?? props.shapeName ?? 'Регион'

                layer.bindTooltip(name, { sticky: true })


                const isLocked = () => lockedRef.current !== null && Number(feature.id) === lockedRef.current

                layer.on({
                    tooltipopen: () => {
                        if (!activeRef.current || isLocked()) layer.closeTooltip()
                    },
                    mouseover: () => {
                        if (activeRef.current && !isLocked()) (layer as Polygon).setStyle(hoverFillRef.current ? HOVER : { ...HOVER, fillOpacity: 0 })
                    },
                    mouseout: () => layerRef.current?.resetStyle(layer),
                    click: () => {
                        if (!activeRef.current || isLocked() || Date.now() < ignoreClicksUntilRef.current) return
                        const area = normalizeArea(feature.geometry as GeoPolygon | MultiPolygon)
                        onSelectRef.current({
                            id: Number(feature.id ?? props.id ?? Math.random() * 1e9),
                            name,
                            bounds: area.bounds,
                            geometry: area.geometry,
                        })
                    },
                })
            }}
        />
    )
}