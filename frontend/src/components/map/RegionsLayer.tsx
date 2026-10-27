import { useEffect, useRef, useState } from 'react'
import { GeoJSON } from 'react-leaflet'
import type { FeatureCollection } from 'geojson'
import type { GeoJSON as LeafletGeoJSON, Layer, Polygon, PathOptions } from 'leaflet'
import type { Place } from '@/api/geocode'

const BASE: PathOptions = { color: '#64748b', weight: 0.8, fillOpacity: 0, opacity: 0.5 }
const HOVER: PathOptions = { color: '#2563eb', weight: 1.5, fillColor: '#2563eb', fillOpacity: 0.05, opacity: 0.5 }

type RegionsLayerProps = {
    active: boolean
    lockedId?: number | null
    onSelect: (place: Place) => void
}

export function RegionsLayer({ active, lockedId = null, onSelect }: RegionsLayerProps) {
    const [data, setData] = useState<FeatureCollection | null>(null)
    const layerRef = useRef<LeafletGeoJSON | null>(null)
    const activeRef = useRef(active)
    const onSelectRef = useRef(onSelect)
    const lockedRef = useRef(lockedId)

    useEffect(() => {
        activeRef.current = active
        onSelectRef.current = onSelect
        lockedRef.current = lockedId
    }, [active, lockedId, onSelect])

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
                        if (activeRef.current && !isLocked()) (layer as Polygon).setStyle(HOVER)
                    },
                    mouseout: () => layerRef.current?.resetStyle(layer),
                    click: () => {
                        if (!activeRef.current) return
                        const b = (layer as Polygon).getBounds()
                        onSelectRef.current({
                            id: Number(feature.id ?? props.id ?? Math.random() * 1e9),
                            name,
                            bounds: { west: b.getWest(), south: b.getSouth(), east: b.getEast(), north: b.getNorth() },
                            geometry: feature.geometry,
                        })
                    },
                })
            }}
        />
    )
}