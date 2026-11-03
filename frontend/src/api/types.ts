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

export type GraphMetrics = {
    page_rank: Record<string, number>
    betweenness: Record<string, number>
    closeness: Record<string, number>
    power_load: {node_id: string; name: string;  degree: number}[]
}

export type LogisticsGraph = {
    id: string
    region_id: string
    bounds?: Bounds
    nodes: LogisticsNode[]
    edges: LogisticsEdge[]
    coverage: Coverage
    metrics?: GraphMetrics
}

export type Coverage = {
    candidates: number
    accepted: number
    rejected: number
    reference_total: number
    reference_matched: number
    unmatched_reference_ids: string[] | null
    unrouted_node_modes: string[] | null
    sources: string[] | null
    warnings: string[] | null
}

export type Job = {
    id: string
    status: 'pending' | 'processing' | 'completed' | 'failed' | 'cancelled'
    graph_id?: string
    error?: string
}