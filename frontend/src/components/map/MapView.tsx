import 'leaflet/dist/leaflet.css'
import type { GeoJsonObject } from 'geojson'
import { MapContainer, TileLayer, GeoJSON, ScaleControl, Rectangle, Pane } from 'react-leaflet'
import { FitToRegion } from './FitToRegion'
import { RegionsLayer } from './RegionsLayer'
import { BBoxDrawer, type BBox } from './BBoxDrawer'
import { useMemo } from 'react'
import type { Coordinates, LogisticsGraph, TransportMode } from '@/api/types'
import { GraphNodes } from './GraphNodes'
import { FitToNodes } from './FitToNodes'
import { GraphEdges } from './GraphEdges'
import { UNKNOWN_KIND } from '@/lib/kinds'
import type { Place } from '@/api/geocode'
import { MapResizer } from './MapResizer'
import { PanToPoint } from './PanToPoint'



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
    onSelectPlace?: (p: Place) => void
    selectedNodeId?: string | null
    onSelectNode?: (id: string) => void
    focusPoint?: Coordinates | null
}

const DEFAULT_CENTER: [number, number] = [61, 90]
const DEFAULT_ZOOM = 3

export function MapView({
    region,
    bbox,
    graph,
    visibleModes,
    hiddenKinds,
    showEdges = true,
    drawMode = false,
    onDrawn = () => { },
    onDrawCancel = () => { },
    onSelectPlace = () => { },
    selectedNodeId = null,
    onSelectNode,
    focusPoint = null,
}: MapViewProps) {


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
        >
            <TileLayer
                attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
                url="https://tile.openstreetmap.org/{z}/{x}/{y}.png"
                maxZoom={19}
            />

            <ScaleControl position="bottomleft" imperial={false} />
            <MapResizer />
            <PanToPoint target={focusPoint} />
            <RegionsLayer active={!drawMode} hoverFill={!graph} lockedId={graph && region ? Number(region.id) : null} onSelect={onSelectPlace} />


            <BBoxDrawer active={drawMode} onDrawn={onDrawn} onCancel={onDrawCancel} />

            {bbox && (
                <Rectangle
                    bounds={[[bbox.south, bbox.west], [bbox.north, bbox.east]]}
                    pathOptions={{ color: '#2563eb', weight: 1.5, dashArray: '6 6', fill: !graph, fillOpacity: 0.05 }}
                    interactive={false}
                />
            )}

            {graph && <FitToNodes nodes={graph.nodes} />}
            <Pane name="graph" style={{ zIndex: 450 }}>
                <GraphEdges edges={visibleEdges} />
                <GraphNodes nodes={visibleNodes} selectedId={selectedNodeId} onSelect={onSelectNode} />
            </Pane>


            {region && (
                <>
                    <GeoJSON
                        key={`${region.id}-${graph ? 'built' : 'draft'}`}
                        data={region.geometry}
                        interactive={false}
                        style={{ color: '#2563eb', weight: 2, fillOpacity: graph ? 0 : 0.08 }}
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