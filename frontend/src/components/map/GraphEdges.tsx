import { Polyline } from "react-leaflet"
import type { LogisticsEdge } from "@/api/types"
import { MODE_COLOR } from "@/lib/modes"

export function GraphEdges({ edges }: { edges: LogisticsEdge[] }) {
    return (
        <>
            {edges.map((edge) => (
                <Polyline
                    key={edge.id}
                    positions={edge.geometry.map((p) => [p.latitude, p.longitude])}
                    pathOptions={{
                        color: MODE_COLOR[edge.mode] ?? '#94a3b8',
                        weight: 2,
                        opacity: 0.7,
                    }}
                    interactive={false}
                >

                </Polyline>
            ))}
        </>
    )
}