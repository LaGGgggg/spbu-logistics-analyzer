import { useEffect, useRef, useState } from "react";
import {
  CircleMarker,
  MapContainer,
  Polyline,
  Popup,
  Rectangle,
  TileLayer,
  Tooltip,
  useMap,
  useMapEvents,
} from "react-leaflet";
import type { LatLngBoundsExpression, Map as LeafletMap } from "leaflet";
import {
  Crosshair,
  Expand,
  MapPinned,
  Minus,
  MousePointer2,
  Plus,
  X,
} from "lucide-react";
import type { Bounds, LogisticsGraph, LogisticsNode } from "../../types/graph";
import { modeColors, modeNames } from "../../types/graph";
import "leaflet/dist/leaflet.css";

function extent(graph: LogisticsGraph): LatLngBoundsExpression {
  if (graph.bounds)
    return [
      [graph.bounds.south, graph.bounds.west],
      [graph.bounds.north, graph.bounds.east],
    ];
  if (graph.nodes.length)
    return graph.nodes.map(
      (n) =>
        [n.coordinates.latitude, n.coordinates.longitude] as [number, number],
    );
  return [
    [59.77, 30.14],
    [59.96, 30.51],
  ];
}
function MapSync({
  graph,
  selected,
}: {
  graph: LogisticsGraph;
  selected: LogisticsNode | null;
}) {
  const map = useMap();
  useEffect(() => {
    map.fitBounds(extent(graph), {
      padding: [45, 45],
      maxZoom: 13,
      animate: false,
    });
  }, [map, graph.id, graph.bounds, graph]);
  useEffect(() => {
    if (selected)
      map.panTo(
        [selected.coordinates.latitude, selected.coordinates.longitude],
        { animate: false },
      );
  }, [map, selected]);
  useEffect(() => {
    const observer = new ResizeObserver(() => map.invalidateSize());
    observer.observe(map.getContainer());
    return () => observer.disconnect();
  }, [map]);
  return null;
}
function AreaPicker({
  onChosen,
  onCancel,
}: {
  onChosen: (bounds: Bounds) => void;
  onCancel: () => void;
}) {
  const [start, setStart] = useState<[number, number] | null>(null);
  const [cursor, setCursor] = useState<[number, number] | null>(null);
  useMapEvents({
    click(event) {
      const point: [number, number] = [event.latlng.lat, event.latlng.lng];
      if (!start) {
        setStart(point);
        setCursor(point);
        return;
      }
      const bounds = {
        west: Math.max(-180, Math.min(start[1], point[1])),
        east: Math.min(180, Math.max(start[1], point[1])),
        south: Math.max(-90, Math.min(start[0], point[0])),
        north: Math.min(90, Math.max(start[0], point[0])),
      };
      if (
        bounds.east - bounds.west > 0.0001 &&
        bounds.north - bounds.south > 0.0001
      )
        onChosen(bounds);
    },
    mousemove(event) {
      if (start) setCursor([event.latlng.lat, event.latlng.lng]);
    },
  });
  useEffect(() => {
    function escape(event: KeyboardEvent) {
      if (event.key === "Escape") onCancel();
    }
    window.addEventListener("keydown", escape);
    return () => window.removeEventListener("keydown", escape);
  }, [onCancel]);
  return start && cursor ? (
    <Rectangle
      bounds={[start, cursor]}
      pathOptions={{
        color: "#0e896b",
        fillOpacity: 0.12,
        dashArray: "6 5",
        weight: 2,
        interactive: false,
      }}
    />
  ) : null;
}
interface Props {
  graph: LogisticsGraph;
  selected: LogisticsNode | null;
  onSelect: (node: LogisticsNode | null) => void;
  drawing: boolean;
  onDrawing: (drawing: boolean) => void;
  onBounds: (bounds: Bounds) => void;
  previewBounds?: Bounds;
}
export default function GraphMap({
  graph,
  selected,
  onSelect,
  drawing,
  onDrawing,
  onBounds,
  previewBounds,
}: Props) {
  const map = useRef<LeafletMap | null>(null);
  const [tileError, setTileError] = useState(false);
  return (
    <div className={`map-stage ${drawing ? "is-drawing" : ""}`}>
      <MapContainer
        ref={map}
        className="graph-map"
        bounds={extent(graph)}
        zoomControl={false}
        zoomAnimation={false}
        fadeAnimation={false}
        markerZoomAnimation={false}
        zoomSnap={0.25}
        preferCanvas
        minZoom={2}
      >
        <TileLayer
          url="https://tile.openstreetmap.org/{z}/{x}/{y}.png"
          maxZoom={19}
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
          eventHandlers={{
            tileerror: () => setTileError(true),
            load: () => setTileError(false),
          }}
        />
        <MapSync graph={graph} selected={selected} />
        {previewBounds && (
          <Rectangle
            bounds={[
              [previewBounds.south, previewBounds.west],
              [previewBounds.north, previewBounds.east],
            ]}
            pathOptions={{
              color: "#0e896b",
              dashArray: "6 6",
              fillOpacity: 0.035,
              weight: 1.5,
              interactive: false,
            }}
          />
        )}
        {graph.edges.map((edge) => (
          <Polyline
            key={edge.id}
            positions={edge.geometry.map((p) => [p.latitude, p.longitude])}
            pathOptions={{
              color: modeColors[edge.mode] ?? "#75848b",
              weight: edge.mode === "rail" ? 3.5 : 2.5,
              opacity: 0.8,
              dashArray:
                edge.mode === "sea" || edge.mode === "river"
                  ? "7 7"
                  : undefined,
              interactive: !drawing,
            }}
          >
            <Popup>
              <div className="map-popup">
                <b>{modeNames[edge.mode]}</b>
                <span>
                  {edge.distance_km.toFixed(2)} км ·{" "}
                  {edge.directed ? "Направленная дуга" : "Двусторонняя связь"}
                </span>
                <small>{edge.source || "Источник не указан"}</small>
              </div>
            </Popup>
          </Polyline>
        ))}
        {graph.nodes.map((node) => (
          <CircleMarker
            key={node.id}
            center={[node.coordinates.latitude, node.coordinates.longitude]}
            radius={selected?.id === node.id ? 10 : 6.5}
            pathOptions={{
              color: "#fff",
              weight: 2.5,
              fillColor: modeColors[node.modes[0]] ?? "#0e896b",
              fillOpacity: 1,
              interactive: !drawing,
            }}
            eventHandlers={{ click: () => onSelect(node) }}
          >
            <Tooltip direction="top" offset={[0, -8]}>
              {node.name}
            </Tooltip>
          </CircleMarker>
        ))}
        {drawing && (
          <AreaPicker
            onChosen={(bounds) => {
              onBounds(bounds);
              onDrawing(false);
            }}
            onCancel={() => onDrawing(false)}
          />
        )}
      </MapContainer>
      <div className="map-top-controls">
        <button
          className={`map-button ${drawing ? "active" : ""}`}
          onClick={() => onDrawing(!drawing)}
        >
          <MousePointer2 size={14} />
          {drawing ? "Отменить выбор" : "Выбрать область"}
        </button>
      </div>
      {drawing && (
        <div className="drawing-notice">
          <Crosshair size={16} /> Укажите два противоположных угла области. Esc
          — отмена.
        </div>
      )}
      {tileError && (
        <div className="tile-notice">
          Подложка карты недоступна. Граф остаётся интерактивным.
        </div>
      )}
      <div className="map-zoom-controls">
        <button
          title="Приблизить карту"
          aria-label="Приблизить карту"
          onClick={() => map.current?.zoomIn()}
        >
          <Plus size={17} />
        </button>
        <button
          title="Отдалить карту"
          aria-label="Отдалить карту"
          onClick={() => map.current?.zoomOut()}
        >
          <Minus size={17} />
        </button>
        <button
          title="Показать весь граф"
          aria-label="Показать весь граф"
          onClick={() =>
            map.current?.fitBounds(extent(graph), {
              padding: [45, 45],
              maxZoom: 13,
              animate: false,
            })
          }
        >
          <Expand size={16} />
        </button>
      </div>
      <button
        className="use-viewport map-button"
        onClick={() => {
          const bounds = map.current?.getBounds();
          if (bounds)
            onBounds({
              west: Math.max(-180, bounds.getWest()),
              east: Math.min(180, bounds.getEast()),
              south: Math.max(-90, bounds.getSouth()),
              north: Math.min(90, bounds.getNorth()),
            });
        }}
      >
        <MapPinned size={14} /> Использовать видимую область
      </button>
      {selected && (
        <aside className="node-detail" aria-label="Карточка узла">
          <div className="detail-heading">
            <span className="eyebrow">УЗЕЛ СЕТИ</span>
            <button
              className="icon-button"
              aria-label="Закрыть карточку узла"
              onClick={() => onSelect(null)}
            >
              <X size={16} />
            </button>
          </div>
          <h3>{selected.name}</h3>
          <code>{selected.id}</code>
          <div className="node-mode-list">
            {selected.modes.map((mode) => (
              <span key={mode} style={{ color: modeColors[mode] }}>
                <i style={{ background: modeColors[mode] }} />
                {modeNames[mode]}
              </span>
            ))}
          </div>
          <dl>
            <div>
              <dt>Широта</dt>
              <dd>{selected.coordinates.latitude.toFixed(5)}</dd>
            </div>
            <div>
              <dt>Долгота</dt>
              <dd>{selected.coordinates.longitude.toFixed(5)}</dd>
            </div>
            <div>
              <dt>Тип</dt>
              <dd>{selected.kind || "—"}</dd>
            </div>
          </dl>
          <div className="detail-footer">
            {selected.external_ids.map((id) => (
              <span key={`${id.namespace}:${id.value}`}>
                {id.namespace}: {id.value}
              </span>
            ))}
            {selected.sources?.join(" · ") || "Источник не указан"}
          </div>
        </aside>
      )}
    </div>
  );
}
