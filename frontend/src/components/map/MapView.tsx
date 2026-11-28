import 'leaflet/dist/leaflet.css'
import type { GeoJsonObject } from 'geojson'
import { MapContainer, TileLayer, GeoJSON, ScaleControl, Rectangle, Pane } from 'react-leaflet'
import { FitToRegion } from './FitToRegion'
import { RegionsLayer } from './RegionsLayer'
import { BBoxDrawer, type BBox } from './BBoxDrawer'
import type { Coordinates, LogisticsEdge, LogisticsGraph, LogisticsNode } from '@/api/types'
import { GraphNodes } from './GraphNodes'
import { FitToNodes } from './FitToNodes'
import { GraphEdges } from './GraphEdges'
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
    building?: boolean
    onDrawn?: (bbox: BBox) => void
    onDrawCancel?: () => void
    graph?: LogisticsGraph | null
    nodes?: LogisticsNode[]
    edges?: LogisticsEdge[]
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
    nodes = [],
    edges = [],
    drawMode = false,
    building = false,
    onDrawn = () => { },
    onDrawCancel = () => { },
    onSelectPlace = () => { },
    selectedNodeId = null,
    onSelectNode,
    focusPoint = null,
}: MapViewProps) {


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
            <MapResizer />
            <PanToPoint target={focusPoint} />
            <RegionsLayer active={!drawMode && !building} hoverFill={!graph} lockedId={graph && region ? Number(region.id) : null} onSelect={onSelectPlace} />


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
                <GraphEdges edges={edges} />
                <GraphNodes nodes={nodes} selectedId={selectedNodeId} onSelect={onSelectNode} />
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