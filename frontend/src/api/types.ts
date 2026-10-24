export type TransportMode = 'rail' | 'sea' | 'river' 

export type Bounds = { west: number; south: number; east: number; north: number }

export type Coordinates = { longitude: number; latitude: number }

export type LogisticsNode = {
    id: string,
    name: string,
    region_id: string,
    coordinates: Coordinates
    modes: TransportMode[]
    external_ids: { namespace: string; value: string }[] | null
    kind?: string
    cargo?: string
}

export type LogisticsEdge = {
    id: string
    from_node_id: string
    to_node_id: string
    mode: TransportMode
    distance_km: number
    geometry: Coordinates[]
    directed: boolean
}

export type LogisticsGraph = {
    id: string
    region_id: string
    bounds?: Bounds
    nodes: LogisticsNode[]
    edges: LogisticsEdge[]
}

export type Job = {
    id: string
    status: 'pending' | 'processing' | 'completed' | 'failed'
    graph_id?: string
    error?: string
}