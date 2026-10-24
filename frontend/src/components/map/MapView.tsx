import 'leaflet/dist/leaflet.css'
import type { GeoJsonObject } from 'geojson'
import { MapContainer, TileLayer, GeoJSON, ScaleControl, Rectangle } from 'react-leaflet'
import { FitToRegion } from './FitToRegion'
import { BBoxDrawer, type BBox } from './BBoxDrawer'
import { useMemo } from 'react'
import type { LogisticsGraph, TransportMode } from '@/api/types'
import { GraphNodes } from './GraphNodes'
import { FitToNodes } from './FitToNodes'
import { GraphEdges } from './GraphEdges'
import { UNKNOWN_KIND } from '@/lib/kinds'



export type MapRegion = {
    id: string,
    name: string,
    geometry: GeoJsonObject
}

type MapViewProps = {
    region?: MapRegion | null
    bbox?: BBox | null
    drawMode?: boolean
    onDrawn?: (bbox: BBox) => void
    onDrawCancel?: () => void
    graph?: LogisticsGraph | null
    visibleModes?: Set<TransportMode>
    hiddenKinds?: Set<string>
    showEdges?: boolean
}

const DEFAULT_CENTER: [number, number] = [61, 90]
const DEFAULT_ZOOM = 3

export function MapView({ region, bbox, graph, visibleModes, hiddenKinds, showEdges = true, drawMode = false, onDrawn = () => { }, onDrawCancel = () => { } }: MapViewProps) {

    
    const visibleNodes = useMemo(() => {
        if (!graph) return []
        return graph.nodes.filter((n) => {
            const modeOk = !visibleModes || n.modes.some((m) => visibleModes.has(m))
            const kindOk = !hiddenKinds || !hiddenKinds.has(n.kind ?? UNKNOWN_KIND)
            return modeOk && kindOk
        })
    }, [graph, visibleModes, hiddenKinds])


    const visibleEdges = useMemo(() => {
        if (!graph || !showEdges) return []

        const visibleIds = new Set(visibleNodes.map((n) => n.id))
        return graph.edges.filter((e) =>
            (!visibleModes || visibleModes.has(e.mode)) &&
            visibleIds.has(e.from_node_id) &&
            visibleIds.has(e.to_node_id),
        )
    }, [graph, showEdges, visibleModes, visibleNodes])


    return (
        <MapContainer
            center={DEFAULT_CENTER}
            zoom={DEFAULT_ZOOM}
            className='h-full w-full'
            worldCopyJump
            attributionControl={false}
            preferCanvas
        >
            <TileLayer
                attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
                url="https://tile.openstreetmap.org/{z}/{x}/{y}.png"
                maxZoom={19}
            />

            <ScaleControl position="bottomleft" imperial={false} />


            <BBoxDrawer active={drawMode} onDrawn={onDrawn} onCancel={onDrawCancel} />

            {bbox && (
                <Rectangle
                    bounds={[[bbox.south, bbox.west], [bbox.north, bbox.east]]}
                    pathOptions={{ color: '#2563eb', weight: 1.5, dashArray: '6 6', fillOpacity: 0.05 }}
                    interactive={false}
                />
            )}

            {graph && <FitToNodes nodes={graph.nodes} />}
            <GraphEdges edges={visibleEdges} />
            <GraphNodes nodes={visibleNodes} />



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


            {drawMode && (
                <div className="pointer-events-none absolute top-3 left-1/2 z-500 -translate-x-1/2 rounded-md bg-foreground px-3 py-1.5 text-xs text-background">
                    Зажмите кнопку мыши и протяните прямоугольник · Esc — отмена
                </div>
            )}
        </MapContainer>
    )
}