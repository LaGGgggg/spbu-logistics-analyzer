import { useState } from "react";
import {
  ArrowDownToLine,
  ArrowUpRight,
  CheckCircle2,
  ChevronLeft,
  ChevronRight,
  Database,
  Info,
  Search,
  ShieldCheck,
  TriangleAlert,
} from "lucide-react";
import { modeColors, modeNames } from "../../types/graph";
import type {
  Coverage,
  LogisticsGraph,
  LogisticsNode,
} from "../../types/graph";

export function NodesTable({
  graph,
  onSelect,
}: {
  graph: LogisticsGraph;
  onSelect: (node: LogisticsNode) => void;
}) {
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(0);
  const nodes = graph.nodes.filter((n) =>
    `${n.name} ${n.id}`
      .toLocaleLowerCase()
      .includes(search.toLocaleLowerCase()),
  );
  const pages = Math.max(1, Math.ceil(nodes.length / 30));
  const actualPage = Math.min(page, pages - 1);
  return (
    <div className="data-view">
      <div className="view-heading">
        <div>
          <h2>Узлы транспортной сети</h2>
          <p>Объекты, принятые в граф, и их внешние идентификаторы</p>
        </div>
        <label className="search-field">
          <Search size={15} />
          <input
            aria-label="Поиск узла"
            placeholder="Найти по названию или ID"
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(0);
            }}
          />
        </label>
      </div>
      <div className="table-scroll">
        <table>
          <thead>
            <tr>
              <th>ОБЪЕКТ</th>
              <th>ТИП</th>
              <th>ТРАНСПОРТ</th>
              <th>КООРДИНАТЫ</th>
              <th />
            </tr>
          </thead>
          <tbody>
            {nodes.slice(actualPage * 30, actualPage * 30 + 30).map((node) => (
              <tr key={node.id}>
                <td>
                  <b>{node.name}</b>
                  <code>{node.id}</code>
                </td>
                <td>{node.kind || "freight_facility"}</td>
                <td>
                  <div className="table-modes">
                    {node.modes.map((mode) => (
                      <span
                        key={mode}
                        title={modeNames[mode]}
                        style={{ background: modeColors[mode] }}
                      />
                    ))}
                    {node.modes.join(" / ")}
                  </div>
                </td>
                <td className="mono">
                  {node.coordinates.latitude.toFixed(4)},{" "}
                  {node.coordinates.longitude.toFixed(4)}
                </td>
                <td>
                  <button
                    className="icon-button"
                    title="Показать узел на карте"
                    aria-label={`Показать ${node.name} на карте`}
                    onClick={() => onSelect(node)}
                  >
                    <ArrowUpRight size={17} />
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {!nodes.length && (
        <div className="empty-view">
          Объекты по заданному запросу не найдены.
        </div>
      )}
      <div className="table-footer">
        <span>
          {nodes.length} объектов · страница {actualPage + 1} из {pages}
        </span>
        <div>
          <button
            className="icon-button"
            disabled={actualPage === 0}
            aria-label="Предыдущая страница"
            onClick={() => setPage(actualPage - 1)}
          >
            <ChevronLeft size={16} />
          </button>
          <button
            className="icon-button"
            disabled={actualPage + 1 >= pages}
            aria-label="Следующая страница"
            onClick={() => setPage(actualPage + 1)}
          >
            <ChevronRight size={16} />
          </button>
        </div>
      </div>
    </div>
  );
}
export function CoverageView({
  coverage,
  example,
  onRefresh,
  refreshing,
}: {
  coverage?: Coverage;
  example: boolean;
  onRefresh: () => void;
  refreshing: boolean;
}) {
  if (!coverage)
    return (
      <div className="empty-view">
        <Info size={24} />
        <p>Backend не вернул coverage для этого графа.</p>
        <button className="secondary-button" onClick={onRefresh}>
          Запросить coverage
        </button>
      </div>
    );
  const ratio = coverage.candidates
    ? Math.round((coverage.accepted / coverage.candidates) * 100)
    : 0;
  return (
    <div className="data-view coverage-view">
      <div className="view-heading">
        <div>
          <h2>Качество и происхождение данных</h2>
          <p>
            {example
              ? "Иллюстрация отчёта. Значения относятся к локальному примеру."
              : "Отчёт относится ко всему графу, независимо от транспортного фильтра."}
          </p>
        </div>
        <button
          className="secondary-button"
          disabled={example || refreshing}
          onClick={onRefresh}
        >
          <ArrowDownToLine size={14} />
          {refreshing ? "Загрузка…" : "Запросить coverage"}
        </button>
      </div>
      <div className="coverage-grid">
        <article className="quality-card">
          <ShieldCheck size={22} />
          <h3>Отбор объектов</h3>
          <strong>
            {coverage.accepted}
            <small> / {coverage.candidates}</small>
          </strong>
          <div className="quality-bar">
            <i style={{ width: `${Math.min(ratio, 100)}%` }} />
          </div>
          <p>
            Принято узлов: {coverage.accepted}. Отклонено кандидатов:{" "}
            {coverage.rejected}. После объединения возможна разница в
            количестве.
          </p>
        </article>
        <article className="quality-card">
          <CheckCircle2 size={22} />
          <h3>Сопоставление со справочниками</h3>
          <strong>
            {coverage.reference_matched}
            <small> / {coverage.reference_total}</small>
          </strong>
          <p>
            Найдено совпадений среди грузовых объектов справочников в выбранной
            области.
          </p>
        </article>
        <article className="quality-card">
          <Database size={22} />
          <h3>Источники</h3>
          {coverage.sources.length ? (
            <ul>
              {coverage.sources.map((s, i) => (
                <li key={i}>{s}</li>
              ))}
            </ul>
          ) : (
            <p>Источники не указаны.</p>
          )}
        </article>
      </div>
      <section className="warnings-card">
        <h3>
          <TriangleAlert size={17} /> Предупреждения{" "}
          <span className="light-count">{coverage.warnings.length}</span>
        </h3>
        {coverage.warnings.length ? (
          coverage.warnings.map((w, i) => <p key={i}>{w}</p>)
        ) : (
          <p>Предупреждений о качестве данных нет.</p>
        )}
      </section>
      <div className="coverage-details">
        {[
          ["Объекты без привязки к сети", coverage.unrouted_node_modes],
          ["Несопоставленные идентификаторы", coverage.unmatched_reference_ids],
        ].map(([label, values]) => (
          <details key={label as string}>
            <summary>
              {label as string}
              <span>{(values as string[]).length}</span>
            </summary>
            <pre>
              {(values as string[]).slice(0, 200).join("\n") || "Нет"}
              {(values as string[]).length > 200
                ? "\n… первые 200 записей"
                : ""}
            </pre>
          </details>
        ))}
      </div>
    </div>
  );
}
