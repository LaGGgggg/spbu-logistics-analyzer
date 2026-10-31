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

function GraphNode({ node, selected, onSelect }: GraphNodeProps) {
    const ref = useRef<LeafletCircleMarker>(null)

    useEffect(() => {
        ref.current?.getElement()?.classList.toggle('graph-node-active', selected)
    }, [selected])

    return (
        <CircleMarker
            ref={ref}
            center={[node.coordinates.latitude, node.coordinates.longitude]}
            radius={7}
            pathOptions={{
                color: '#fff',
                weight: 1.5,
                fillColor: MODE_COLOR[node.modes[0]] ?? '#94a3b8',
                fillOpacity: 0.9,
                className: 'graph-node',
            }}
            eventHandlers={{ click: () => onSelect?.(node.id) }}
        >
            <Tooltip>
                <div className='font-medium'>{node.name}</div>
                <div className='text-xs'>{node.modes.map((m) => MODE_LABEL[m] ?? m).join(', ')}</div>
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
