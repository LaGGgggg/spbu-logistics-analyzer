import { useEffect, useState } from 'react'
import { Rectangle, useMap, useMapEvents } from 'react-leaflet'
import { DomUtil, type LeafletMouseEvent } from 'leaflet'

export type BBox = {
    north: number
    south: number
    west: number
    east: number
}

type BBoxDrawerProps = {
    active: boolean
    onDrawn: (bbox: BBox) => void
    onCancel: () => void
}



export function BBoxDrawer({ active, onDrawn, onCancel }: BBoxDrawerProps) {
    const map = useMap()
    const [start, setStart] = useState<[number, number] | null>(null)
    const [current, setCurrent] = useState<[number, number] | null>(null)

    useEffect(() => {
        const container = map.getContainer()
        if (active) {
            map.dragging.disable()
            container.style.cursor = 'crosshair'
            DomUtil.disableTextSelection()
        } else {
            map.dragging.enable()
            container.style.cursor = ''
            DomUtil.enableTextSelection()
        }
    }, [active, map])



    useEffect(() => {
        if (!active) return
        const onKey = (e: KeyboardEvent) => {
            if (e.key === 'Escape') onCancel()
        }
        window.addEventListener('keydown', onKey)
        return () => window.removeEventListener('keydown', onKey)
    }, [active, onCancel])



    useMapEvents({
        mousedown(e: LeafletMouseEvent) {
            if (!active) return
            setStart([e.latlng.lat, e.latlng.lng])
            setCurrent([e.latlng.lat, e.latlng.lng])
        },
        mousemove(e: LeafletMouseEvent) {
            if (active && start) setCurrent([e.latlng.lat, e.latlng.lng])
        },
        mouseup(e: LeafletMouseEvent) {
            if (!active || !start) return
            const [lat1, lng1] = start
            const { lat: lat2, lng: lng2 } = e.latlng
            setStart(null)
            setCurrent(null)

            if (Math.abs(lat1 - lat2) < 0.02 || Math.abs(lng1 - lng2) < 0.02) return
            onDrawn({
                north: +Math.max(lat1, lat2).toFixed(3),
                south: +Math.min(lat1, lat2).toFixed(3),
                west: +Math.min(lng1, lng2).toFixed(3),
                east: +Math.max(lng1, lng2).toFixed(3),
            })
        },
    })

    if (!start || !current) return null
    return (
        <Rectangle
            bounds={[start, current]}
            pathOptions={{ color: '#2563eb', weight: 1.5, dashArray: '4 4', fillOpacity: 0.08 }}
        />
    )
}
