import { useState } from "react";
import {
  Activity,
  ArrowDownToLine,
  ArrowRight,
  ArrowUpRight,
  BookOpen,
  Check,
  ChevronDown,
  ChevronRight,
  CircleHelp,
  Code2,
  Copy,
  Database,
  Download,
  ExternalLink,
  FileJson,
  Globe2,
  Hexagon,
  Layers3,
  LoaderCircle,
  Map,
  MapPin,
  Network,
  Play,
  RefreshCw,
  Route,
  Settings2,
  Square,
  Terminal,
  TriangleAlert,
  Waypoints,
  X,
} from "lucide-react";
import GraphMap from "./components/map/GraphMap";
import DebugPanel from "./components/workbench/DebugPanel";
import { CoverageView, NodesTable } from "./components/workbench/DataViews";
import ApiReference from "./components/workbench/ApiReference";
import { getCoverage, getExport, saveFile } from "./api/graphs";
import { useLogistics } from "./hooks/useLogistics";
import type {
  Bounds,
  BuildRequest,
  Coverage,
  ExportFormat,
  LogisticsNode,
} from "./types/graph";
import { exportFormats, modeColors, modeNames, modes } from "./types/graph";
import "./App.css";

type Tab = "map" | "nodes" | "coverage" | "api";
const formatNumber = (n: number) =>
  new Intl.NumberFormat("ru-RU", { maximumFractionDigits: 1 }).format(n);
const fieldNames: Record<keyof Bounds, string> = {
  west: "Запад",
  south: "Юг",
  east: "Восток",
  north: "Север",
};
const initialBounds = {
  west: "30.16",
  south: "59.85",
  east: "30.34",
  north: "59.97",
};
function App() {
  const app = useLogistics();
  const [tab, setTab] = useState<Tab>("map");
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [areaMode, setAreaMode] = useState(false);
  const [regionId, setRegionId] = useState("");
  const [boundsFields, setBoundsFields] = useState(initialBounds);
  const [drawing, setDrawing] = useState(false);
  const [selected, setSelected] = useState<LogisticsNode | null>(null);
  const [existingId, setExistingId] = useState("");
  const [existingKind, setExistingKind] = useState<"job" | "graph">("graph");
  const [coverage, setCoverage] = useState<{
    id: string;
    value: Coverage;
  } | null>(null);
  const [refreshingCoverage, setRefreshingCoverage] = useState(false);
  const [exporting, setExporting] = useState("");
  const [exportOpen, setExportOpen] = useState(false);
  const [toast, setToast] = useState("");
  const [copied, setCopied] = useState(false);
  const selectedRegionId = regionId || app.regions[0]?.id || "";
  const bounds = Object.fromEntries(
    Object.entries(boundsFields).map(([key, value]) => [key, Number(value)]),
  ) as unknown as Bounds;
  const boundsValid =
    Object.values(boundsFields).every((value) => value.trim() !== "") &&
    Object.values(bounds).every(Number.isFinite) &&
    bounds.west >= -180 &&
    bounds.east <= 180 &&
    bounds.south >= -90 &&
    bounds.north <= 90 &&
    bounds.west < bounds.east &&
    bounds.south < bounds.north;
  const requestBody: BuildRequest = areaMode
    ? { bounds }
    : { region_id: selectedRegionId };
  const currentCoverage =
    coverage?.id === app.graph.id ? coverage.value : app.graph.coverage;
  const title =
    app.source === "example"
      ? "Санкт-Петербург"
      : app.regions.find((r) => r.id === app.graph.region_id)?.name ||
        "Пользовательская область";
  const totalKM = app.viewGraph.edges.reduce(
    (sum, edge) => sum + edge.distance_km,
    0,
  );
  const availableModes = modes.filter((mode) =>
    app.graph.nodes.some((node) => node.modes.includes(mode)),
  );
  function notify(message: string) {
    setToast(message);
    window.setTimeout(() => setToast(""), 3500);
  }
  function selectBounds(next: Bounds) {
    setBoundsFields(
      Object.fromEntries(
        Object.entries(next).map(([key, value]) => [key, value.toFixed(5)]),
      ) as typeof initialBounds,
    );
    setAreaMode(true);
    setTab("map");
    notify("Границы перенесены в форму. Нажмите «Построить граф».");
  }
  async function startBuild() {
    setSelected(null);
    setTab("map");
    await app.build(requestBody);
  }
  async function refreshCoverage() {
    setRefreshingCoverage(true);
    try {
      const value = await getCoverage(app.graph.id);
      setCoverage({ id: app.graph.id, value });
    } catch (cause) {
      app.setError(
        cause instanceof Error ? cause.message : "Не удалось получить coverage",
      );
    } finally {
      setRefreshingCoverage(false);
    }
  }
  async function download(format: ExportFormat) {
    setExporting(format);
    setExportOpen(false);
    try {
      const blob = await getExport(app.graph.id, format);
      saveFile(
        blob,
        `graph-${app.graph.id}.${format === "geojson" ? "geojson" : format === "graphml" ? "graphml" : format}`,
      );
      notify("Файл получен от backend.");
    } catch (cause) {
      app.setError(cause instanceof Error ? cause.message : "Ошибка экспорта");
    } finally {
      setExporting("");
    }
  }
  return (
    <div className="app-shell">
      <header className="app-header">
        <a
          className="brand"
          href={location.pathname}
          aria-label="Logistics Atlas — главная"
        >
          <span className="brand-symbol">
            <Waypoints size={23} strokeWidth={2} />
          </span>
          <span>
            logistics<span className="brand-light">atlas</span>
            <small>NETWORK EXPLORER</small>
          </span>
        </a>
        <nav className="top-nav" aria-label="Главная навигация">
          <button
            className={tab !== "api" ? "active" : ""}
            onClick={() => setTab("map")}
          >
            <Network size={15} /> Рабочая область
          </button>
          <button
            className={tab === "api" ? "active" : ""}
            onClick={() => setTab("api")}
          >
            <Code2 size={15} /> API и отладка
          </button>
        </nav>
        <div className="header-right">
          <button
            className={`connection-pill ${app.health}`}
            onClick={() => void app.checkConnection()}
            title="Повторно проверить /health и /api/regions"
          >
            <i />
            {app.health === "online"
              ? "Backend подключён"
              : app.health === "checking"
                ? "Подключение…"
                : "Backend недоступен"}
            <RefreshCw size={11} />
          </button>
          <span className="header-divider" />
          <span className="workspace-avatar">SP</span>
        </div>
      </header>

      <div className="workspace-layout">
        <aside className={`sidebar ${settingsOpen ? "mobile-open" : ""}`}>
          <div className="sidebar-title">
            <span className="eyebrow">НОВОЕ ИССЛЕДОВАНИЕ</span>
            <Settings2 className="desktop-settings-icon" size={15} />
            <button
              className="mobile-settings-button"
              onClick={() => setSettingsOpen(!settingsOpen)}
              aria-expanded={settingsOpen}
            >
              <Settings2 size={15} /> Параметры территории{" "}
              <ChevronDown size={14} />
            </button>
          </div>
          <h2>Территория анализа</h2>
          <p className="sidebar-description">
            От исходных данных — к связям транспортной инфраструктуры.
          </p>
          <div className="segmented" aria-label="Способ выбора территории">
            <button
              className={!areaMode ? "active" : ""}
              onClick={() => setAreaMode(false)}
            >
              <MapPin size={14} /> Регион
            </button>
            <button
              className={areaMode ? "active" : ""}
              onClick={() => setAreaMode(true)}
            >
              <Square size={13} /> Своя область
            </button>
          </div>
          {areaMode ? (
            <div className="area-form">
              <div className="field-label">
                Границы области <span>WGS84</span>
              </div>
              <div className="coordinate-grid">
                {(["west", "east", "south", "north"] as const).map((key) => (
                  <label key={key}>
                    {fieldNames[key]}
                    <span>{key}</span>
                    <input
                      type="number"
                      step="any"
                      value={boundsFields[key]}
                      onChange={(e) =>
                        setBoundsFields({
                          ...boundsFields,
                          [key]: e.target.value,
                        })
                      }
                      aria-label={fieldNames[key]}
                    />
                  </label>
                ))}
              </div>
              <button
                className={`outline-map-button ${drawing ? "active" : ""}`}
                onClick={() => {
                  setTab("map");
                  setDrawing(!drawing);
                }}
              >
                <Square size={14} />
                {drawing ? "Отменить выбор" : "Указать на карте"}
                <ArrowUpRight size={14} />
              </button>
              {!boundsValid && (
                <p className="field-error">
                  Проверьте координаты: west &lt; east, south &lt; north;
                  долгота ±180°, широта ±90°.
                </p>
              )}
            </div>
          ) : (
            <div className="region-form">
              <label className="field-label" htmlFor="region-select">
                Регион <span>{app.regions.length} в каталоге</span>
              </label>
              <div className="select-wrap">
                <Globe2 size={16} />
                <select
                  id="region-select"
                  value={selectedRegionId}
                  onChange={(e) => setRegionId(e.target.value)}
                  disabled={!app.regions.length}
                >
                  <option value="" disabled>
                    {app.health === "checking"
                      ? "Получение регионов…"
                      : "Подключите backend"}
                  </option>
                  {app.regions.map((r) => (
                    <option key={r.id} value={r.id}>
                      {r.name}
                    </option>
                  ))}
                </select>
                <ChevronDown size={13} />
              </div>
              <div className="region-note">
                <MapPin size={13} />
                <span>
                  {app.regions.find((r) => r.id === selectedRegionId)?.id ||
                    "GET /api/regions"}
                </span>
                <span className="tiny-tag">OSM</span>
              </div>
            </div>
          )}
          <div className="source-card">
            <div>
              <Database size={15} />
              <b>OpenStreetMap</b>
              <span className="source-indicator" />
            </div>
            <p>
              Региональная выгрузка или Overpass. Источник подбирает backend.
            </p>
            <span>
              XML / PBF <i /> Кэширование
            </span>
          </div>
          <button
            className="primary-button build-button"
            onClick={() => void startBuild()}
            disabled={app.busy || (areaMode ? !boundsValid : !selectedRegionId)}
          >
            {app.busy ? (
              <LoaderCircle size={17} className="spin" />
            ) : (
              <Play size={15} fill="currentColor" />
            )}
            {app.busy ? "Построение графа…" : "Построить граф"}
            {!app.busy && <ArrowRight size={16} />}
          </button>
          <p className="build-help">
            Первое построение может потребовать загрузки большой выгрузки.
            Задание выполняется в фоне.
          </p>
          {(app.busy || app.pendingJobId || app.job) && (
            <div className="job-card" aria-live="polite">
              <div>
                <span className={`job-dot ${app.busy ? "pulse" : ""}`} />
                <b>
                  {app.busy
                    ? app.job?.status === "pending"
                      ? "В очереди на backend"
                      : app.job?.status === "processing"
                        ? "Строится на backend"
                        : "Получение результата…"
                    : app.job?.status === "completed"
                      ? "Построение завершено"
                      : app.job?.status === "failed"
                        ? "Ошибка построения"
                        : "Ожидание приостановлено"}
                </b>
              </div>
              <code>
                {app.pendingJobId || app.job?.id || "POST /api/graphs"}
              </code>
              {app.busy ? (
                <button className="text-button" onClick={app.stopWaiting}>
                  Прекратить ожидание
                </button>
              ) : app.pendingJobId && app.job?.status !== "failed" ? (
                <button
                  className="text-button"
                  onClick={() => void app.open(app.pendingJobId, "job")}
                >
                  Продолжить ожидание <ArrowRight size={13} />
                </button>
              ) : null}
              <p>Остановка ожидания не отменяет задание на сервере.</p>
            </div>
          )}
          <div className="sidebar-divider" />
          <div className="sidebar-subheading">УЖЕ ЕСТЬ РЕЗУЛЬТАТ?</div>
          <form
            className="open-form"
            onSubmit={(e) => {
              e.preventDefault();
              setSelected(null);
              void app.open(existingId, existingKind);
            }}
          >
            <div className="open-kind">
              <button
                type="button"
                className={existingKind === "graph" ? "active" : ""}
                onClick={() => setExistingKind("graph")}
              >
                Граф
              </button>
              <button
                type="button"
                className={existingKind === "job" ? "active" : ""}
                onClick={() => setExistingKind("job")}
              >
                Задание
              </button>
            </div>
            <div className="id-input">
              <input
                aria-label="Идентификатор результата"
                placeholder={
                  existingKind === "graph"
                    ? "Введите graph_id"
                    : "Введите job_id"
                }
                value={existingId}
                onChange={(e) => setExistingId(e.target.value)}
                maxLength={128}
              />
              <button
                aria-label="Открыть по ID"
                title="Открыть по ID"
                disabled={!existingId.trim() || app.busy}
              >
                <ArrowRight size={16} />
              </button>
            </div>
          </form>
          <button
            className="example-card"
            onClick={() => {
              app.showExample();
              setSelected(null);
              setTab("map");
            }}
          >
            <span className="example-icon">
              <Hexagon size={21} />
            </span>
            <span>
              <b>Посмотреть пример</b>
              <small>Без загрузки данных</small>
            </span>
            <ChevronRight size={16} />
          </button>
          <div className="sidebar-bottom">
            <BookOpen size={14} />
            <span>Go + Gin · PostgreSQL</span>
            <button
              className="icon-button"
              title="Открыть справку API"
              aria-label="Открыть справку API"
              onClick={() => setTab("api")}
            >
              <CircleHelp size={15} />
            </button>
          </div>
        </aside>

        <main className="main-workspace">
          <div className="workspace-breadcrumb">
            <span>Исследования</span>
            <ChevronRight size={12} />
            <span>Транспортная сеть</span>
            <span className="workspace-version">WORKSPACE / 01</span>
          </div>
          <div className="workspace-title">
            <div>
              <div className="title-label">ЛОГИСТИЧЕСКИЙ ГРАФ</div>
              <h1>
                {title}
                <span className="title-dot">.</span>
              </h1>
              <p>
                Исследуйте инфраструктуру, связи и качество исходных данных.
              </p>
            </div>
            <div className="title-actions">
              <button
                className="secondary-button"
                title="Скопировать ссылку на текущий результат"
                onClick={async () => {
                  try {
                    await navigator.clipboard.writeText(location.href);
                    setCopied(true);
                    window.setTimeout(() => setCopied(false), 1800);
                  } catch {
                    notify("Скопируйте адрес из строки браузера.");
                  }
                }}
              >
                {copied ? <Check size={14} /> : <Copy size={14} />}
                {copied ? "Скопировано" : "Ссылка"}
              </button>
              <div className="export-wrapper">
                <button
                  className="dark-button"
                  onClick={() => setExportOpen(!exportOpen)}
                  disabled={!!exporting}
                >
                  <Download size={15} />
                  {exporting ? "Загрузка…" : "Экспорт"}
                  <ChevronDown size={13} />
                </button>
                {exportOpen && (
                  <>
                    <button
                      className="dropdown-backdrop"
                      aria-label="Закрыть меню экспорта"
                      onClick={() => setExportOpen(false)}
                    />
                    <div className="export-menu">
                      <div className="menu-label">
                        {app.source === "example"
                          ? "ЛОКАЛЬНЫЙ ПРИМЕР"
                          : "ПОЛНЫЙ ГРАФ BACKEND"}
                      </div>
                      {app.source === "example" ? (
                        <>
                          <p>
                            Серверные форматы станут доступны после построения.
                          </p>
                          <button
                            onClick={() => {
                              saveFile(
                                new Blob([JSON.stringify(app.graph, null, 2)], {
                                  type: "application/json",
                                }),
                                "example-graph.json",
                              );
                              setExportOpen(false);
                            }}
                          >
                            <FileJson size={15} /> JSON примера
                            <ArrowDownToLine size={14} />
                          </button>
                        </>
                      ) : (
                        exportFormats.map((format) => (
                          <button
                            key={format}
                            onClick={() => void download(format)}
                          >
                            <FileJson size={15} />
                            {format}
                            <ArrowDownToLine size={14} />
                          </button>
                        ))
                      )}
                      {app.mode !== "all" && (
                        <p>Экспортируется весь граф, без фильтра.</p>
                      )}
                    </div>
                  </>
                )}
              </div>
            </div>
          </div>
          {app.error && (
            <div className="error-banner" role="alert">
              <TriangleAlert size={17} />
              <span>{app.error}</span>
              <button
                className="icon-button"
                aria-label="Закрыть ошибку"
                onClick={() => app.setError("")}
              >
                <X size={15} />
              </button>
            </div>
          )}
          <div className="stats-grid">
            <article className="stat-card">
              <span className="stat-icon green">
                <Waypoints size={20} />
              </span>
              <div>
                <span>Узлы сети</span>
                <strong>{formatNumber(app.viewGraph.nodes.length)}</strong>
              </div>
              <span className="stat-caption">объектов</span>
            </article>
            <article className="stat-card">
              <span className="stat-icon blue">
                <Route size={20} />
              </span>
              <div>
                <span>Транспортные связи</span>
                <strong>{formatNumber(app.viewGraph.edges.length)}</strong>
              </div>
              <span className="stat-caption">дуг</span>
            </article>
            <article className="stat-card">
              <span className="stat-icon amber">
                <Layers3 size={20} />
              </span>
              <div>
                <span>Виды транспорта</span>
                <strong>
                  {app.mode === "all" ? availableModes.length : 1}
                  <small> / 5</small>
                </strong>
              </div>
              <span className="stat-caption">в графе</span>
            </article>
            <article className="stat-card">
              <span className="stat-icon violet">
                <Activity size={20} />
              </span>
              <div>
                <span>Сумма длин дуг</span>
                <strong>{formatNumber(totalKM)}</strong>
              </div>
              <span className="stat-caption">км</span>
            </article>
          </div>
          <section className="graph-workspace">
            <div className="workspace-tabs">
              <div role="tablist" aria-label="Представление результата">
                {(
                  [
                    { id: "map", label: "Карта сети", icon: Map },
                    { id: "nodes", label: "Узлы", icon: Waypoints },
                    {
                      id: "coverage",
                      label: "Качество данных",
                      icon: ShieldIcon,
                    },
                    { id: "api", label: "API", icon: Code2 },
                  ] as const
                ).map((item) => (
                  <button
                    key={item.id}
                    role="tab"
                    aria-selected={tab === item.id}
                    className={tab === item.id ? "active" : ""}
                    onClick={() => setTab(item.id)}
                  >
                    <item.icon size={15} />
                    {item.label}
                    {item.id === "nodes" && (
                      <span className="light-count">
                        {app.viewGraph.nodes.length}
                      </span>
                    )}
                  </button>
                ))}
              </div>
              <span
                className={`result-badge ${app.source === "example" ? "example" : "real"}`}
              >
                <i />
                {app.source === "example"
                  ? "Пример · локальные данные"
                  : "Результат backend"}
              </span>
            </div>
            {(tab === "map" || tab === "nodes") && (
              <div className="transport-toolbar">
                <span className="toolbar-label">ТРАНСПОРТ</span>
                <button
                  className={app.mode === "all" ? "active" : ""}
                  disabled={app.busy}
                  onClick={() => {
                    setSelected(null);
                    void app.changeMode("all");
                  }}
                >
                  Все виды
                </button>
                {modes.map((mode) => (
                  <button
                    key={mode}
                    className={app.mode === mode ? "active" : ""}
                    disabled={app.busy}
                    onClick={() => {
                      setSelected(null);
                      void app.changeMode(mode);
                    }}
                  >
                    <i style={{ background: modeColors[mode] }} />
                    {modeNames[mode]}
                  </button>
                ))}
                {app.filtering && <LoaderCircle size={14} className="spin" />}
              </div>
            )}
            {tab === "map" && (
              <GraphMap
                graph={app.viewGraph}
                selected={selected}
                onSelect={setSelected}
                drawing={drawing}
                onDrawing={setDrawing}
                onBounds={selectBounds}
                previewBounds={areaMode && boundsValid ? bounds : undefined}
              />
            )}
            {tab === "nodes" && (
              <NodesTable
                graph={app.viewGraph}
                onSelect={(node) => {
                  setSelected(node);
                  setTab("map");
                }}
              />
            )}
            {tab === "coverage" && (
              <CoverageView
                coverage={currentCoverage}
                example={app.source === "example"}
                onRefresh={() => void refreshCoverage()}
                refreshing={refreshingCoverage}
              />
            )}
            {tab === "api" && (
              <ApiReference
                graphId={app.graph.id}
                jobId={app.job?.id || app.pendingJobId}
                example={app.source === "example"}
                body={requestBody}
                onError={app.setError}
              />
            )}
            <footer className="graph-footer">
              <span>
                <i className="live-dot" />
                {app.source === "example" ? (
                  "Иллюстративная топология · не результат расчёта"
                ) : (
                  <>
                    <span className="graph-id-label">graph_id</span>{" "}
                    <code>{app.graph.id}</code>
                  </>
                )}
              </span>
              <span>
                {app.source === "example"
                  ? "12 объектов · Санкт-Петербург"
                  : "Геометрия и расстояния из backend"}
                <ExternalLink size={11} />
              </span>
            </footer>
          </section>
          <DebugPanel example={app.source === "example"} />
          <footer className="workspace-footer">
            <span>
              <Network size={13} /> LOGISTICS ATLAS <i /> Инфраструктура в
              деталях
            </span>
            <button onClick={() => setTab("api")}>
              <Terminal size={12} /> Документация API
              <ArrowUpRight size={12} />
            </button>
          </footer>
        </main>
      </div>
      {toast && (
        <div className="toast" role="status">
          <Check size={16} />
          {toast}
          <button
            className="icon-button"
            aria-label="Закрыть уведомление"
            onClick={() => setToast("")}
          >
            <X size={14} />
          </button>
        </div>
      )}
    </div>
  );
}
function ShieldIcon({ size = 15 }: { size?: number }) {
  return <Check size={size} />;
}
export default App;
