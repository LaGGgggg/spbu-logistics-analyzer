import type {
  BuildRequest,
  BuildResult,
  Coverage,
  ExportFormat,
  Job,
  LogisticsGraph,
  ModeFilter,
  Region,
} from "../types/graph";

export interface ApiTrace {
  id: number;
  method: string;
  path: string;
  startedAt: string;
  status: number | "pending" | "network" | "aborted";
  duration?: number;
  request?: unknown;
  response?: unknown;
  contentType?: string;
}
let traces: ApiTrace[] = [];
let nextId = 1;
const listeners = new Set<() => void>();
export const subscribeTraces = (listener: () => void) => {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
};
export const getTraces = () => traces;
export const clearTraces = () => {
  traces = [];
  listeners.forEach((listener) => listener());
};
function safePreview(value: unknown): unknown {
  try {
    const json = JSON.stringify(value, (key, entry: unknown) =>
      /password|token|authorization|cookie|secret/i.test(key)
        ? "[скрыто]"
        : entry,
    );
    if (!json) return value;
    return json.length > 24_000
      ? `${json.slice(0, 24_000)}\n… ответ сокращён в консоли`
      : (JSON.parse(json) as unknown);
  } catch {
    return "[не удалось отобразить ответ]";
  }
}
function emit(trace: ApiTrace) {
  traces = [trace, ...traces.filter((item) => item.id !== trace.id)]
    .sort((a, b) => b.id - a.id)
    .slice(0, 80);
  listeners.forEach((listener) => listener());
}
export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}
function errorMessage(data: unknown, status: number): string {
  if (data && typeof data === "object" && "error" in data) {
    const value = data.error;
    if (typeof value === "string") return value;
    if (
      value &&
      typeof value === "object" &&
      "message" in value &&
      typeof value.message === "string"
    )
      return value.message;
  }
  return status === 502 || status === 503
    ? "Backend недоступен. Проверьте Go-сервис и PostgreSQL."
    : `Ошибка HTTP ${status}`;
}
export async function request<T>(
  path: string,
  options: { signal?: AbortSignal; body?: unknown; download?: boolean } = {},
): Promise<T> {
  const started = performance.now();
  const trace: ApiTrace = {
    id: nextId++,
    method: options.body === undefined ? "GET" : "POST",
    path,
    startedAt: new Date().toISOString(),
    status: "pending",
    request: safePreview(options.body),
  };
  emit(trace);
  try {
    const response = await fetch(path, {
      method: trace.method,
      credentials: "same-origin",
      signal: options.signal
        ? AbortSignal.any([options.signal, AbortSignal.timeout(60_000)])
        : AbortSignal.timeout(60_000),
      headers:
        options.body === undefined
          ? undefined
          : { "Content-Type": "application/json" },
      body:
        options.body === undefined ? undefined : JSON.stringify(options.body),
    });
    const contentType = response.headers.get("content-type") ?? "";
    let data: unknown;
    if (options.download && response.ok) {
      data = await response.blob();
      emit({
        ...trace,
        status: response.status,
        duration: performance.now() - started,
        contentType,
        response: {
          type: contentType,
          bytes: (data as Blob).size,
          download: true,
        },
      });
    } else {
      const text = await response.text();
      try {
        data = text ? (JSON.parse(text) as unknown) : null;
      } catch {
        data = text.slice(0, 24_000);
      }
      emit({
        ...trace,
        status: response.status,
        duration: performance.now() - started,
        contentType,
        response: safePreview(data),
      });
    }
    if (!response.ok)
      throw new ApiError(errorMessage(data, response.status), response.status);
    return data as T;
  } catch (error) {
    if (error instanceof ApiError) throw error;
    const aborted = options.signal?.aborted === true;
    const message = aborted
      ? "Ожидание остановлено в браузере"
      : error instanceof DOMException && error.name === "TimeoutError"
        ? "Backend не ответил за 60 секунд. Можно повторить запрос."
        : "Не удалось подключиться к backend. Проверьте сервис и адрес proxy.";
    emit({
      ...trace,
      status: aborted ? "aborted" : "network",
      duration: performance.now() - started,
      response: { error: message },
    });
    if (aborted) throw new DOMException(message, "AbortError");
    throw new Error(message, { cause: error });
  }
}
const idPath = (id: string) => encodeURIComponent(id);
export const getHealth = (signal?: AbortSignal) =>
  request<{ status: string; service: string }>("/health", { signal });
export const getRegions = (signal?: AbortSignal) =>
  request<Region[]>("/api/regions", { signal });
export const requestBuild = (body: BuildRequest, signal?: AbortSignal) =>
  request<BuildResult>("/api/graphs", { body, signal });
export const getJob = (id: string, signal?: AbortSignal, after?: string) => {
  const query = after
    ? `?${new URLSearchParams({ wait: "25", after }).toString()}`
    : "";
  return request<Job>(`/api/jobs/${idPath(id)}${query}`, { signal });
};
export const getGraph = (
  id: string,
  signal?: AbortSignal,
  mode: ModeFilter = "all",
) =>
  request<LogisticsGraph>(
    `/api/graphs/${idPath(id)}${mode === "all" ? "" : `?mode=${mode}`}`,
    { signal },
  );
export const getCoverage = (id: string, signal?: AbortSignal) =>
  request<Coverage>(`/api/graphs/${idPath(id)}/coverage`, { signal });
export const getExport = (
  id: string,
  format: ExportFormat,
  signal?: AbortSignal,
) =>
  request<Blob>(`/api/graphs/${idPath(id)}/export/${format}`, {
    signal,
    download: true,
  });
export function saveFile(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  link.click();
  window.setTimeout(() => URL.revokeObjectURL(url), 30_000);
}
