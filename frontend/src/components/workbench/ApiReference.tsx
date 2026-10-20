import { useState } from "react";
import {
  ArrowUpRight,
  Check,
  Copy,
  Download,
  Play,
  ShieldOff,
  Waypoints,
} from "lucide-react";
import specification from "../../../../openapi.yaml?raw";
import {
  getCoverage,
  getGraph,
  getHealth,
  getJob,
  getRegions,
  saveFile,
} from "../../api/graphs";
import type { BuildRequest } from "../../types/graph";

function shellQuote(value: string) {
  return "'" + value.replaceAll("'", "'\\''") + "'";
}

const endpoints = [
  {
    method: "GET",
    path: "/health",
    description: "Доступность HTTP-сервиса",
    response: "200",
    action: "health",
  },
  {
    method: "GET",
    path: "/api/regions",
    description: "Пресеты регионов и их границы",
    response: "200",
    action: "regions",
  },
  {
    method: "POST",
    path: "/api/graphs",
    description: "Поставить задание или получить готовый граф",
    response: "200 / 202",
    action: "build",
  },
  {
    method: "GET",
    path: "/api/jobs/{id}",
    description: "Статус и graph_id. Long polling: ?wait=25&after=updated_at",
    response: "200 / 404",
    action: "job",
  },
  {
    method: "GET",
    path: "/api/graphs/{id}",
    description: "Узлы, дуги и геометрия. Фильтр ?mode=rail",
    response: "200 / 404",
    action: "graph",
  },
  {
    method: "GET",
    path: "/api/graphs/{id}/coverage",
    description: "Полнота, источники и предупреждения",
    response: "200 / 404",
    action: "coverage",
  },
  {
    method: "GET",
    path: "/api/graphs/{id}/export/{format}",
    description: "CSV, матрицы, GeoJSON и GraphML",
    response: "200 / 422",
    action: "export",
  },
];
export default function ApiReference({
  graphId,
  jobId,
  example,
  body,
  onError,
}: {
  graphId: string;
  jobId?: string;
  example: boolean;
  body: BuildRequest;
  onError: (error: string) => void;
}) {
  const [running, setRunning] = useState("");
  const [copied, setCopied] = useState(false);
  const curl = `curl -X POST ${shellQuote(`${__BACKEND_TARGET__}/api/graphs`)} \\\n  -H 'Content-Type: application/json' \\\n  -d ${shellQuote(JSON.stringify(body))}`;
  async function run(action: string) {
    setRunning(action);
    try {
      switch (action) {
        case "health":
          await getHealth();
          break;
        case "regions":
          await getRegions();
          break;
        case "graph":
          await getGraph(graphId);
          break;
        case "job":
          if (jobId) await getJob(jobId);
          break;
        case "coverage":
          await getCoverage(graphId);
          break;
      }
    } catch (cause) {
      onError(cause instanceof Error ? cause.message : "Ошибка запроса");
    } finally {
      setRunning("");
    }
  }
  return (
    <div className="data-view api-reference">
      <div className="view-heading">
        <div>
          <span className="eyebrow">BACKEND CONTRACT</span>
          <h2>API, которое стоит за картой</h2>
          <p>
            Запускайте безопасные GET-запросы и смотрите ответы в консоли ниже.
          </p>
        </div>
        <button
          className="secondary-button"
          onClick={() =>
            saveFile(
              new Blob([specification], { type: "application/yaml" }),
              "openapi.yaml",
            )
          }
        >
          <Download size={14} /> OpenAPI 3.1
        </button>
      </div>
      <div className="api-info">
        <span>
          <i className="live-dot" />
          {__BACKEND_TARGET__}
        </span>
        <span>
          <ShieldOff size={14} /> Без авторизации
        </span>
        <span>
          <Waypoints size={14} /> REST · JSON · Polling
        </span>
      </div>
      <div className="endpoint-list">
        {endpoints.map((endpoint) => {
          const runnable =
            ["health", "regions"].includes(endpoint.action) ||
            (endpoint.action === "job" && !!jobId) ||
            (!example && ["graph", "coverage"].includes(endpoint.action));
          return (
            <article className="endpoint" key={endpoint.path}>
              <b className={`verb ${endpoint.method === "POST" ? "post" : ""}`}>
                {endpoint.method}
              </b>
              <div>
                <code>{endpoint.path}</code>
                <p>{endpoint.description}</p>
              </div>
              <span className="endpoint-response">{endpoint.response}</span>
              {runnable ? (
                <button
                  className="icon-button"
                  aria-label={`Выполнить ${endpoint.path}`}
                  title="Выполнить запрос"
                  disabled={!!running}
                  onClick={() => void run(endpoint.action)}
                >
                  <Play
                    size={15}
                    className={running === endpoint.action ? "pulse" : ""}
                  />
                </button>
              ) : (
                <span
                  className="endpoint-hint"
                  title={
                    endpoint.action === "build"
                      ? "Запуск через форму слева"
                      : endpoint.action === "export"
                        ? "Скачивание через кнопку «Экспорт»"
                        : "Сначала получите результат backend"
                  }
                >
                  <ArrowUpRight size={15} />
                </span>
              )}
            </article>
          );
        })}
      </div>
      <div className="curl-card">
        <div>
          <span>POST /api/graphs</span>
          <button
            className="text-button"
            onClick={async () => {
              try {
                await navigator.clipboard.writeText(curl);
                setCopied(true);
                window.setTimeout(() => setCopied(false), 1500);
              } catch {
                onError(
                  "Браузер запретил копирование. Выделите команду вручную.",
                );
              }
            }}
          >
            {copied ? <Check size={13} /> : <Copy size={13} />}
            {copied ? "Скопировано" : "Копировать cURL"}
          </button>
        </div>
        <pre>{curl}</pre>
      </div>
      <p className="api-footnote">
        Доступные возможности отражают текущий backend. Поиск по названию,
        отмена задания, PageRank и betweenness пока не реализованы. Локальный
        пример не отправляется на сервер.
      </p>
    </div>
  );
}
