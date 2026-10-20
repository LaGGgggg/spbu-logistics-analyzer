import { useState, useSyncExternalStore } from "react";
import {
  Braces,
  Check,
  ChevronDown,
  ChevronUp,
  Copy,
  Terminal,
  Trash2,
} from "lucide-react";
import { clearTraces, getTraces, subscribeTraces } from "../../api/graphs";

function Json({ value }: { value: unknown }) {
  const text =
    typeof value === "string" ? value : JSON.stringify(value ?? null, null, 2);
  return (
    <pre>
      {text
        .split(
          /("(?:\\.|[^"\\])*"|\btrue\b|\bfalse\b|\bnull\b|\b\d+(?:\.\d+)?\b)/g,
        )
        .map((part, i) => (
          <span
            key={i}
            className={
              part.startsWith('"')
                ? "json-string"
                : /^(true|false|null|\d)/.test(part)
                  ? "json-number"
                  : undefined
            }
          >
            {part}
          </span>
        ))}
    </pre>
  );
}
export default function DebugPanel({ example }: { example: boolean }) {
  const traces = useSyncExternalStore(subscribeTraces, getTraces);
  const [open, setOpen] = useState(true);
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const [detailTab, setDetailTab] = useState<"response" | "request">(
    "response",
  );
  const [copied, setCopied] = useState(false);
  const selected = traces.find((item) => item.id === selectedId) ?? traces[0];
  const failed = traces.filter(
    (t) =>
      t.status === "network" ||
      (typeof t.status === "number" && t.status >= 400),
  ).length;
  async function copy() {
    if (!selected) return;
    try {
      await navigator.clipboard.writeText(JSON.stringify(selected, null, 2));
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1500);
    } catch {
      setCopied(false);
    }
  }
  return (
    <section
      className={`debug-panel ${open ? "expanded" : ""}`}
      aria-label="Консоль API"
    >
      <div className="debug-toolbar">
        <button className="debug-title" onClick={() => setOpen(!open)}>
          <Terminal size={16} /> Консоль API{" "}
          <span className="count-chip">{traces.length}</span>
          {open ? <ChevronDown size={14} /> : <ChevronUp size={14} />}
        </button>
        <div className="debug-toolbar-right">
          <span className={failed ? "console-error-label" : ""}>
            {failed ? `${failed} ошибок` : "HTTP / JSON"}
          </span>
          <button
            title="Очистить журнал"
            aria-label="Очистить журнал"
            onClick={() => {
              clearTraces();
              setSelectedId(null);
            }}
          >
            <Trash2 size={14} />
          </button>
        </div>
      </div>
      {open && (
        <div className="debug-body">
          <div className="request-log">
            <div className="request-log-head">
              <span>МЕТОД / РЕСУРС</span>
              <span>СТАТУС</span>
              <span>ВРЕМЯ</span>
            </div>
            {traces.map((trace) => (
              <button
                key={trace.id}
                className={`request-row ${selected?.id === trace.id ? "selected" : ""}`}
                onClick={() => {
                  setSelectedId(trace.id);
                  setCopied(false);
                }}
              >
                <span className="request-path">
                  <b
                    className={
                      trace.method === "POST" ? "method-post" : "method-get"
                    }
                  >
                    {trace.method}
                  </b>
                  <span title={trace.path}>{trace.path}</span>
                </span>
                <span
                  className={`http-status ${typeof trace.status === "number" && trace.status < 400 ? "success" : trace.status === "pending" ? "waiting" : trace.status === "aborted" ? "muted" : "failure"}`}
                >
                  {trace.status === "network"
                    ? "Сеть"
                    : trace.status === "pending"
                      ? "…"
                      : trace.status === "aborted"
                        ? "Стоп"
                        : trace.status}
                </span>
                <span className="duration">
                  {trace.duration === undefined
                    ? "—"
                    : `${Math.round(trace.duration)} ms`}
                </span>
              </button>
            ))}
            {!traces.length && (
              <div className="console-empty">
                Здесь появятся настоящие запросы к API.
                {example && " Пример графа не создаёт запросов."}
              </div>
            )}
          </div>
          <div className="response-panel">
            <div className="response-tabs">
              <button
                className={detailTab === "response" ? "active" : ""}
                onClick={() => setDetailTab("response")}
              >
                <Braces size={13} /> Ответ
              </button>
              <button
                className={detailTab === "request" ? "active" : ""}
                onClick={() => setDetailTab("request")}
              >
                Запрос
              </button>
              <span>
                {selected?.contentType?.split(";")[0] || "application/json"}
              </span>
              <button
                title="Скопировать запись"
                aria-label="Скопировать запись"
                disabled={!selected}
                onClick={() => void copy()}
              >
                {copied ? <Check size={13} /> : <Copy size={13} />}
              </button>
            </div>
            {selected ? (
              <Json
                value={
                  detailTab === "response"
                    ? (selected.response ?? { status: "Ожидание ответа…" })
                    : {
                        method: selected.method,
                        path: selected.path,
                        body: selected.request ?? null,
                        time: selected.startedAt,
                      }
                }
              />
            ) : (
              <div className="console-empty">
                Выберите запрос для просмотра JSON.
              </div>
            )}
          </div>
        </div>
      )}
    </section>
  );
}
