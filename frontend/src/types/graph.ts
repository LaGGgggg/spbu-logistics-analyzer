export type TransportMode = "road" | "rail" | "river" | "sea" | "air";
export type ModeFilter = TransportMode | "all";
export interface Coordinates {
  longitude: number;
  latitude: number;
}
export interface Bounds {
  west: number;
  south: number;
  east: number;
  north: number;
}
export interface Region {
  id: string;
  name: string;
  bounds: Bounds;
}
export interface LogisticsNode {
  id: string;
  name: string;
  region_id: string;
  coordinates: Coordinates;
  modes: TransportMode[];
  external_ids: { namespace: string; value: string }[];
  kind?: string;
  cargo?: string;
  sources?: string[];
}
export interface LogisticsEdge {
  id: string;
  from_node_id: string;
  to_node_id: string;
  mode: TransportMode;
  distance_km: number;
  directed: boolean;
  geometry: Coordinates[];
  source?: string;
}
export interface Coverage {
  candidates: number;
  accepted: number;
  rejected: number;
  reference_total: number;
  reference_matched: number;
  unmatched_reference_ids: string[];
  unrouted_node_modes: string[];
  sources: string[];
  warnings: string[];
}
export interface LogisticsGraph {
  id: string;
  region_id: string;
  bounds?: Bounds;
  nodes: LogisticsNode[];
  edges: LogisticsEdge[];
  coverage?: Coverage;
}
export interface Job {
  id: string;
  region_id: string;
  status: "pending" | "processing" | "completed" | "failed";
  graph_id?: string;
  error?: string;
  created_at: string;
  updated_at: string;
}
export interface BuildResult {
  job_id?: string;
  graph_id?: string;
  region?: Region;
}
export type BuildRequest = { region_id: string } | { bounds: Bounds };
export const modeNames: Record<TransportMode, string> = {
  rail: "Железная дорога",
  road: "Автодороги",
  sea: "Море",
  river: "Реки",
  air: "Авиация",
};
export const modeColors: Record<TransportMode, string> = {
  rail: "#0e896b",
  road: "#d8943c",
  sea: "#5277c7",
  river: "#289fb2",
  air: "#9a74c7",
};
export const modes: TransportMode[] = ["rail", "road", "sea", "river", "air"];
export const exportFormats = [
  "geojson",
  "graphml",
  "nodes.csv",
  "edges.csv",
  "adjacency.csv",
  "incidence.csv",
] as const;
export type ExportFormat = (typeof exportFormats)[number];
