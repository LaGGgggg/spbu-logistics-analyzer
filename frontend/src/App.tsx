import { useEffect, useRef, useState } from "react";
import {
  Activity,
  ArrowDownToLine,
  ArrowLeft,
  ArrowRight,
  ArrowUpRight,
  Check,
  ChevronDown,
  ChevronRight,
  Code2,
  Copy,
  Download,
  FileJson,
  FolderOpen,
  Globe2,
  Hexagon,
  Layers3,
  LoaderCircle,
  Map,
  MapPin,
  Play,
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
  const [toolsOpen, setToolsOpen] = useState(false);
  const [showDebug, setShowDebug] = useState(false);
  const toolsRef = useRef<HTMLDivElement>(null);
  useEffect(() => {
    if (!toolsOpen) return;
    const outside = (event: PointerEvent) => {
      if (
        event.target instanceof Node &&
        !toolsRef.current?.contains(event.target)
      )
        setToolsOpen(false);
    };
    const escape = (event: KeyboardEvent) => {
      if (event.key === "Escape") setToolsOpen(false);
    };
    document.addEventListener("pointerdown", outside);
    document.addEventListener("keydown", escape);
    return () => {
      document.removeEventListener("pointerdown", outside);
      document.removeEventListener("keydown", escape);
    };
  }, [toolsOpen]);
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
  const connectionLabel =
    app.health === "online"
      ? "Сервис доступен"
      : app.health === "checking"
        ? "Проверка соединения"
        : "Сервис недоступен";
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
            <span className="brand-main">logistics</span>
            <span className="brand-light">atlas</span>
          </span>
        </a>
        <div className="header-right">
          <button
            className={`connection-indicator ${app.health}`}
            aria-label={connectionLabel}
            title={`${connectionLabel}. Проверить соединение`}
            onClick={() => void app.checkConnection()}
          >
            <Activity size={17} />
            <i />
          </button>
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

          <div className="tools-wrapper" ref={toolsRef}>
            <button
              className={`icon-button tools-trigger ${toolsOpen ? "active" : ""}`}
              aria-label="Инструменты"
              aria-expanded={toolsOpen}
              aria-controls="workspace-tools"
              title="Инструменты"
              onClick={() => setToolsOpen(!toolsOpen)}
            >
              <Settings2 size={19} />
            </button>
            {toolsOpen && (
              <div className="utility-menu" id="workspace-tools">
                <button
                  onClick={() => {
                    setTab("api");
                    setToolsOpen(false);
                  }}
                  aria-label="Открыть API"
                >
                  <Code2 size={17} />
                  <span>Справка API</span>
                  <ArrowUpRight size={14} />
                </button>
                <button
                  onClick={() => {
                    setShowDebug(!showDebug);
                    setToolsOpen(false);
                  }}
                  aria-pressed={showDebug}
                >
                  <Terminal size={17} />
                  <span>Консоль запросов</span>
                  {showDebug && <Check size={14} />}
                </button>
                <details className="open-result">
                  <summary>
                    <FolderOpen size={17} />
                    Открыть по ID
                    <ChevronDown size={14} />
                  </summary>
                  <form
                    className="open-form"
                    onSubmit={(e) => {
                      e.preventDefault();
                      setSelected(null);
                      setToolsOpen(false);
                      void app.open(existingId, existingKind);
                    }}
                  >
                    <label className="utility-field">
                      Тип результата
                      <select
                        aria-label="Тип результата"
                        value={existingKind}
                        onChange={(event) =>
                          setExistingKind(event.target.value as "graph" | "job")
                        }
                      >
                        <option value="graph">Результат расчёта</option>
                        <option value="job">Фоновое построение</option>
                      </select>
                    </label>
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
                </details>
              </div>
            )}
          </div>
        </div>
      </header>

      <div className="workspace-layout">
        <aside className={`sidebar ${settingsOpen ? "mobile-open" : ""}`}>
          <div className="sidebar-title">
            <h2>Территория</h2>
            <button
              className="mobile-settings-button"
              onClick={() => setSettingsOpen(!settingsOpen)}
              aria-expanded={settingsOpen}
            >
              <Settings2 size={15} /> Параметры территории{" "}
              <ChevronDown size={14} />
            </button>
          </div>
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
              <div className="field-label">Координаты</div>
              <div className="coordinate-grid">
                {(["west", "east", "south", "north"] as const).map((key) => (
                  <label key={key}>
                    {fieldNames[key]}
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
                Регион
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
            </div>
          )}
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
              <details className="job-details">
                <summary>Подробнее</summary>
                <code>{app.pendingJobId || app.job?.id}</code>
              </details>
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
            </div>
          )}

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
            </span>
            <ChevronRight size={16} />
          </button>
        </aside>

        <main className="main-workspace">
          <h1 className="sr-only">Карта логистической инфраструктуры</h1>
          {app.health === "offline" && (
            <div className="offline-notice" role="status">
              <TriangleAlert size={15} />
              Нет соединения с сервером
              <button onClick={() => void app.checkConnection()}>
                Повторить
              </button>
            </div>
          )}
          {tab === "api" && (
            <button className="back-to-map" onClick={() => setTab("map")}>
              <ArrowLeft size={15} />К карте
            </button>
          )}
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
                <span>Узлы</span>
                <strong>{formatNumber(app.viewGraph.nodes.length)}</strong>
              </div>
              <span className="stat-caption">объектов</span>
            </article>
            <article className="stat-card">
              <span className="stat-icon blue">
                <Route size={20} />
              </span>
              <div>
                <span>Связи</span>
                <strong>{formatNumber(app.viewGraph.edges.length)}</strong>
              </div>
              <span className="stat-caption">дуг</span>
            </article>
            <article className="stat-card">
              <span className="stat-icon amber">
                <Layers3 size={20} />
              </span>
              <div>
                <span>Транспорт</span>
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
                <span>Длина, км</span>
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
          </section>
          {showDebug && (
            <DebugPanel
              example={app.source === "example"}
              onClose={() => setShowDebug(false)}
            />
          )}
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
