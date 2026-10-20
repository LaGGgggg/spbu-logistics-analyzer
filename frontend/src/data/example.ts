import type {
  LogisticsGraph,
  LogisticsNode,
  TransportMode,
} from "../types/graph";

// Illustrative topology for UI presentation, not an OSM extract or a computed backend result.
const facilities: [string, string, number, number, TransportMode[], string][] =
  [
    [
      "port",
      "Большой порт Санкт-Петербург",
      30.211,
      59.884,
      ["sea", "rail", "road"],
      "port",
    ],
    [
      "avtovo",
      "Грузовой терминал Автово",
      30.272,
      59.866,
      ["rail", "road"],
      "freight_terminal",
    ],
    [
      "baltic",
      "Балтийский логистический узел",
      30.294,
      59.909,
      ["rail", "road"],
      "freight_terminal",
    ],
    [
      "shushary",
      "Распределительный центр Шушары",
      30.416,
      59.814,
      ["rail", "road"],
      "warehouse",
    ],
    [
      "predport",
      "Предпортовая",
      30.29,
      59.839,
      ["rail", "road"],
      "freight_terminal",
    ],
    [
      "neva",
      "Невский речной терминал",
      30.452,
      59.889,
      ["river", "road"],
      "port",
    ],
    [
      "volod",
      "Володарский грузовой узел",
      30.439,
      59.865,
      ["rail", "road"],
      "freight_terminal",
    ],
    [
      "sort",
      "Сортировочная-Московская",
      30.369,
      59.874,
      ["rail", "road"],
      "freight_terminal",
    ],
    [
      "pulkovo",
      "Грузовой терминал Пулково",
      30.274,
      59.799,
      ["air", "road"],
      "cargo_airport",
    ],
    [
      "obuh",
      "Терминал Обухово",
      30.466,
      59.837,
      ["rail", "road"],
      "freight_terminal",
    ],
    [
      "warehouse",
      "Южный распределительный центр",
      30.355,
      59.827,
      ["road"],
      "warehouse",
    ],
    [
      "island",
      "Морской грузовой терминал",
      30.178,
      59.924,
      ["sea", "road"],
      "port",
    ],
  ];
const nodes: LogisticsNode[] = facilities.map(
  ([id, name, longitude, latitude, modes, kind]) => ({
    id: `example:${id}`,
    name,
    coordinates: { longitude, latitude },
    modes,
    kind,
    region_id: "example-spb",
    external_ids: [],
    sources: ["Иллюстративный набор"],
  }),
);
const connections: [number, number, TransportMode, number][] = [
  [0, 1, "rail", 5.3],
  [1, 4, "rail", 4.2],
  [4, 3, "rail", 11.8],
  [3, 9, "rail", 5.6],
  [9, 6, "rail", 4.8],
  [6, 7, "rail", 6.1],
  [7, 2, "rail", 7.4],
  [2, 0, "rail", 8.2],
  [0, 11, "sea", 6.7],
  [5, 6, "river", 3.1],
  [8, 4, "road", 5.1],
  [4, 10, "road", 5.8],
  [10, 3, "road", 4.1],
  [7, 10, "road", 5.6],
  [5, 7, "road", 7.7],
  [0, 4, "road", 8.4],
];
export const exampleGraph: LogisticsGraph = {
  id: "example-spb",
  region_id: "example-spb",
  bounds: { west: 30.14, south: 59.77, east: 30.5, north: 59.95 },
  nodes,
  edges: connections.map(([a, b, mode, distance_km], i) => ({
    id: `example:edge:${i}`,
    from_node_id: nodes[a].id,
    to_node_id: nodes[b].id,
    mode,
    distance_km,
    directed: true,
    source: "Иллюстративная связь, не реальный маршрут",
    geometry: [
      nodes[a].coordinates,
      {
        longitude:
          (nodes[a].coordinates.longitude + nodes[b].coordinates.longitude) /
            2 +
          (i % 2 ? 0.008 : -0.008),
        latitude:
          (nodes[a].coordinates.latitude + nodes[b].coordinates.latitude) / 2 +
          0.005,
      },
      nodes[b].coordinates,
    ],
  })),
  coverage: {
    candidates: 15,
    accepted: 12,
    rejected: 3,
    reference_total: 0,
    reference_matched: 0,
    unmatched_reference_ids: [],
    unrouted_node_modes: ["example:pulkovo:air"],
    sources: ["Локальный пример для презентации"],
    warnings: [
      "Это иллюстративные узлы и связи. Для получения реальных данных запустите построение на backend.",
    ],
  },
};
