import { CircleMarker, Tooltip } from 'react-leaflet'
import type { LogisticsNode } from '@/api/types'
import { MODE_COLOR, MODE_LABEL } from '@/lib/modes'

export function GraphNodes({ nodes }: { nodes: LogisticsNode[] }) {
    return (
        <>
            {nodes.map((node) => (
                <CircleMarker
                    key={node.id}
                    center={[node.coordinates.latitude, node.coordinates.longitude]}
                    radius={5}
                    pathOptions={{ color: '#fff', weight: 1, fillColor: MODE_COLOR[node.modes[0]] ?? '#94a3b8', fillOpacity: 0.9 }}
                >
                    <Tooltip>
                        <div className='font-medium'>{node.name}</div>
                        <div className='text-xs'>{node.modes.map((m) => MODE_LABEL[m] ?? m).join(', ')}</div>
                    </Tooltip>
                </CircleMarker>
            ))}
        </>
    )
}
