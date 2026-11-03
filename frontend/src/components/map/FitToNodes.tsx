import { useEffect } from 'react'
import { useMap } from 'react-leaflet'
import { latLngBounds } from 'leaflet'
import type { LogisticsNode } from '@/api/types'

export function FitToNodes({ nodes }: { nodes: LogisticsNode[] }) {
    const map = useMap()

    useEffect(() => {
        if (nodes.length === 0) return
        const bounds = latLngBounds(nodes.map((n) => [n.coordinates.latitude, n.coordinates.longitude]))
        map.fitBounds(bounds, { padding: [32, 32] })
    }, [nodes, map])

    return null
}
