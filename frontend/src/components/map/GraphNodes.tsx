import { useEffect, useRef } from 'react'
import { CircleMarker, Tooltip } from 'react-leaflet'
import type { CircleMarker as LeafletCircleMarker } from 'leaflet'
import type { LogisticsNode } from '@/api/types'
import { MODE_COLOR, MODE_LABEL } from '@/lib/modes'

type GraphNodesProps = {
    nodes: LogisticsNode[]
    selectedId?: string | null
    onSelect?: (id: string) => void
}

type GraphNodeProps = {
    node: LogisticsNode
    selected: boolean
    onSelect?: (id: string) => void
}

const BASE_RADIUS = 7
const MAX_RADIUS = 22

function nodeRadius(weight: number): number {
    return Math.min(BASE_RADIUS * Math.sqrt(weight), MAX_RADIUS)
}

function GraphNode({ node, selected, onSelect }: GraphNodeProps) {
    const ref = useRef<LeafletCircleMarker>(null)
    const weight = Math.max(node.weight ?? 1, 1)

    useEffect(() => {
        ref.current?.getElement()?.classList.toggle('graph-node-active', selected)
    }, [selected])

    return (
        <CircleMarker
            ref={ref}
            center={[node.coordinates.latitude, node.coordinates.longitude]}
            radius={nodeRadius(weight)}
            pathOptions={{
                color: '#fff',
                weight: 1.5,
                fillColor: MODE_COLOR[node.modes[0]] ?? '#94a3b8',
                fillOpacity: 0.9,
                className: 'graph-node',
            }}
            eventHandlers={{ click: () => onSelect?.(node.id) }}
        >
            <Tooltip pane="tooltipPane">
                <div className='font-medium'>{node.name}</div>
                <div className='text-xs'>{node.modes.map((m) => MODE_LABEL[m] ?? m).join(', ')}</div>
                {weight > 1 && (
                    <div className='text-xs'>Объединяет узлов: {weight}</div>
                )}
            </Tooltip>
        </CircleMarker>
    )
}

export function GraphNodes({ nodes, selectedId = null, onSelect }: GraphNodesProps) {
    return (
        <>
            {nodes.map((node) => (
                <GraphNode key={node.id} node={node} selected={node.id === selectedId} onSelect={onSelect} />
            ))}
        </>
    )
}
