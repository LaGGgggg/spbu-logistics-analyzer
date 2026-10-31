import { useEffect } from 'react'
import { useMap } from 'react-leaflet'
import type { Coordinates } from '@/api/types'

export function PanToPoint({ target }: { target: Coordinates | null }) {
    const map = useMap()

    useEffect(() => {
        if (target) map.panTo([target.latitude, target.longitude], { animate: true, duration: 0.5 })
    }, [map, target])

    return null
}
