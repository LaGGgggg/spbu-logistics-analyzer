import { useEffect, useMemo, useRef, useState } from 'react'
import { CircleMarker, Tooltip, useMap } from 'react-leaflet'
import type { CircleMarker as LeafletCircleMarker, Map as LeafletMap } from 'leaflet'
import type { LogisticsNode } from '@/api/types'
import { MODE_COLOR, MODE_LABEL } from '@/lib/modes'

const BASE_RADIUS = 7
const MAX_RADIUS = 22
const OPAQUE = 0.95
const FADED = 0.35

function nodeWeight(node: LogisticsNode): number {
    return Math.max(node.weight ?? 1, 1)
}

function nodeRadius(weight: number): number {
    return Math.min(BASE_RADIUS * Math.sqrt(weight), MAX_RADIUS)
}

// A node is faded only when a heavier node's marker overlaps it on screen;
// distant or equally weighted markers keep full opacity.
function coveredNodeIds(nodes: LogisticsNode[], map: LeafletMap, zoom: number): Set<string> {

    type projectedNode = { id: string; weight: number; x: number; y: number; radius: number }

    const projected: projectedNode[] = nodes.map((node) => {
        const point = map.project([node.coordinates.latitude, node.coordinates.longitude], zoom)
        const weight = nodeWeight(node)

        return { id: node.id, weight, x: point.x, y: point.y, radius: nodeRadius(weight) }
    })

    const covered = new Set<string>()

    for (let i = 0; i < projected.length; i++) {
        for (let j = i + 1; j < projected.length; j++) {
            const a = projected[i]
            const b = projected[j]

            if (a.weight === b.weight) {
                continue
            }

            const [heavy, light] = a.weight > b.weight ? [a, b] : [b, a]
            const distance = Math.hypot(a.x - b.x, a.y - b.y)

            if (distance < heavy.radius + light.radius) {
                covered.add(light.id)
            }
        }
    }

    return covered
}

type GraphNodesProps = {
    nodes: LogisticsNode[]
    selectedId?: string | null
    onSelect?: (id: string) => void
}

type GraphNodeProps = {
    node: LogisticsNode
    selected: boolean
    faded: boolean
    onSelect?: (id: string) => void
}

function GraphNode({ node, selected, faded, onSelect }: GraphNodeProps) {
    const ref = useRef<LeafletCircleMarker>(null)

    useEffect(() => {
        ref.current?.getElement()?.classList.toggle('graph-node-active', selected)
    }, [selected])

    return (
        <CircleMarker
            ref={ref}
            center={[node.coordinates.latitude, node.coordinates.longitude]}
            radius={nodeRadius(nodeWeight(node))}
            pathOptions={{
                color: '#fff',
                weight: 1.5,
                fillColor: MODE_COLOR[node.modes[0]] ?? '#94a3b8',
                fillOpacity: selected || !faded ? OPAQUE : FADED,
                className: 'graph-node',
            }}
            eventHandlers={{ click: () => onSelect?.(node.id) }}
        >
            <Tooltip pane="tooltipPane">
                <div className='font-medium'>{node.name}</div>
                <div className='text-xs'>{node.modes.map((m) => MODE_LABEL[m] ?? m).join(', ')}</div>
                {nodeWeight(node) > 1 && (
                    <div className='text-xs'>Объединяет узлов: {nodeWeight(node)}</div>
                )}
            </Tooltip>
        </CircleMarker>
    )
}

export function GraphNodes({ nodes, selectedId = null, onSelect }: GraphNodesProps) {
    const map = useMap()
    const [zoom, setZoom] = useState(map.getZoom())

    useEffect(() => {
        const onChange = () => setZoom(map.getZoom())

        map.on('zoomend', onChange)

        return () => {
            map.off('zoomend', onChange)
        }
    }, [map])

    const covered = useMemo(
        () => coveredNodeIds(nodes, map, zoom),
        [nodes, map, zoom],
    )

    return (
        <>
            {nodes.map((node) => (
                <GraphNode
                    key={node.id}
                    node={node}
                    selected={node.id === selectedId}
                    faded={covered.has(node.id)}
                    onSelect={onSelect}
                />
            ))}
        </>
    )
}
