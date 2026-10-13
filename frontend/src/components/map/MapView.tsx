import 'leaflet/dist/leaflet.css'
import type { GeoJsonObject } from 'geojson'
import { MapContainer, TileLayer, GeoJSON, ScaleControl } from 'react-leaflet'
import { FitToRegion } from './FitToRegion'

export type MapRegion = {
    id: string,
    name: string,
    geometry: GeoJsonObject
}

type MapViewProps = {
    region?: MapRegion | null
}

const DEFAULT_CENTER: [number, number] = [61, 90]
const DEFAULT_ZOOM = 3

export function MapView({ region }: MapViewProps) {
    return (
        <MapContainer
            center={DEFAULT_CENTER}
            zoom={DEFAULT_ZOOM}
            className='h-full w-full'
            worldCopyJump
        >
            <TileLayer
                attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
                url="https://tile.openstreetmap.org/{z}/{x}/{y}.png"
                maxZoom={19}
            />

            <ScaleControl position="bottomleft" imperial={false} />

            {region && (
                <>
                    <GeoJSON
                        key={region.id}
                        data={region.geometry}
                        style={{ color: '#2563eb', weight: 2, fillOpacity: 0.08 }}
                    />
                    <FitToRegion geometry={region.geometry} />
                </>
            )}
        </MapContainer>
    )
}