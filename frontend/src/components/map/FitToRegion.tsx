import { useEffect } from 'react'
import { useMap } from 'react-leaflet'
import { geoJSON } from 'leaflet'
import type { GeoJsonObject } from 'geojson'

type FitToRegionType = {
    geometry: GeoJsonObject
}

export function FitToRegion({geometry}: FitToRegionType) {
    const map = useMap()
    useEffect(()=>{
        const bounds = geoJSON(geometry).getBounds()
        if(bounds.isValid()) {
            map.flyToBounds(bounds, { padding: [24, 24], duration: 0.8 })
        }
    }, [map, geometry])
    return null
}