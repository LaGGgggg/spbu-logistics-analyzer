import { useCallback, useEffect, useRef, useState } from "react";
import {
  getGraph,
  getHealth,
  getJob,
  getRegions,
  requestBuild,
} from "../api/graphs";
import { exampleGraph } from "../data/example";
import type {
  BuildRequest,
  Job,
  LogisticsGraph,
  ModeFilter,
  Region,
} from "../types/graph";

export function useLogistics() {
  const [graph, setGraph] = useState<LogisticsGraph>(exampleGraph);
  const [viewGraph, setViewGraph] = useState<LogisticsGraph>(exampleGraph);
  const [source, setSource] = useState<"example" | "backend">("example");
  const [regions, setRegions] = useState<Region[]>([]);
  const [health, setHealth] = useState<"checking" | "online" | "offline">(
    "checking",
  );
  const [busy, setBusy] = useState(false);
  const [filtering, setFiltering] = useState(false);
  const [job, setJob] = useState<Job | null>(null);
  const [error, setError] = useState("");
  const [mode, setMode] = useState<ModeFilter>("all");
  const [pendingJobId, setPendingJobId] = useState("");
  const work = useRef<AbortController | null>(null);
  const filter = useRef<AbortController | null>(null);

  const loadGraph = useCallback(async (id: string, signal: AbortSignal) => {
    const result = await getGraph(id, signal);
    if (signal.aborted) return;
    if (!Array.isArray(result.nodes) || !Array.isArray(result.edges))
      throw new Error("Backend вернул граф в неизвестном формате.");
    setGraph(result);
    setViewGraph(result);
    setSource("backend");
    setMode("all");
    setPendingJobId("");
    const url = new URL(location.href);
    url.searchParams.delete("job");
    url.searchParams.set("graph", id);
    history.replaceState(null, "", url);
  }, []);

  const watchJob = useCallback(
    async (id: string, signal: AbortSignal) => {
      setPendingJobId(id);
      const url = new URL(location.href);
      url.searchParams.delete("graph");
      url.searchParams.set("job", id);
      history.replaceState(null, "", url);
      let after: string | undefined;
      while (!signal.aborted) {
        const state = await getJob(id, signal, after);
        if (signal.aborted) return;
        setJob(state);
        if (state.status === "failed")
          throw new Error(
            state.error ||
              "Построение завершилось ошибкой. Подробности доступны в логах backend.",
          );
        if (state.status === "completed") {
          if (!state.graph_id)
            throw new Error("Задание завершено, но graph_id отсутствует.");
          await loadGraph(state.graph_id, signal);
          setPendingJobId("");
          return;
        }
        if (!state.updated_at || Number.isNaN(Date.parse(state.updated_at)))
          throw new Error("В ответе задания отсутствует корректный updated_at.");
        after = state.updated_at;
      }
    },
    [loadGraph],
  );

  const checkConnection = useCallback(async (signal?: AbortSignal) => {
    const [healthResult, regionsResult] = await Promise.allSettled([
      getHealth(signal),
      getRegions(signal),
    ]);
    if (signal?.aborted) return;
    setHealth(
      healthResult.status === "fulfilled" &&
        healthResult.value.status === "healthy"
        ? "online"
        : "offline",
    );
    if (
      regionsResult.status === "fulfilled" &&
      Array.isArray(regionsResult.value)
    )
      setRegions(regionsResult.value);
  }, []);

  useEffect(() => {
    const controller = new AbortController();
    const connection = new AbortController();
    work.current = controller;
    async function initialize() {
      await checkConnection(connection.signal);
      if (controller.signal.aborted) return;
      const params = new URLSearchParams(location.search);
      const jobId = params.get("job");
      const graphId = params.get("graph");
      if (!jobId && !graphId) return;
      setBusy(true);
      try {
        if (jobId) await watchJob(jobId, controller.signal);
        else if (graphId) await loadGraph(graphId, controller.signal);
      } catch (cause) {
        if (!controller.signal.aborted)
          setError(
            cause instanceof Error
              ? cause.message
              : "Не удалось открыть результат.",
          );
      } finally {
        if (!controller.signal.aborted) setBusy(false);
      }
    }
    void initialize();
    return () => {
      controller.abort();
      connection.abort();
      work.current?.abort();
      filter.current?.abort();
    };
  }, [checkConnection, loadGraph, watchJob]);

  async function perform(action: (signal: AbortSignal) => Promise<void>) {
    work.current?.abort();
    filter.current?.abort();
    const controller = new AbortController();
    work.current = controller;
    setBusy(true);
    setError("");
    setFiltering(false);
    try {
      await action(controller.signal);
    } catch (cause) {
      if (!controller.signal.aborted)
        setError(
          cause instanceof Error
            ? cause.message
            : "Не удалось выполнить запрос.",
        );
    } finally {
      if (!controller.signal.aborted) setBusy(false);
    }
  }

  async function build(body: BuildRequest) {
    setJob(null);
    setPendingJobId("");
    await perform(async (signal) => {
      const result = await requestBuild(body, signal);
      if (signal.aborted) return;
      if (result.graph_id) await loadGraph(result.graph_id, signal);
      else if (result.job_id) await watchJob(result.job_id, signal);
      else throw new Error("В ответе backend отсутствуют job_id и graph_id.");
    });
  }

  async function open(id: string, kind: "job" | "graph") {
    if (!id.trim() || id.trim().length > 128) {
      setError("Укажите ID длиной от 1 до 128 символов.");
      return;
    }
    setJob(null);
    await perform((signal) =>
      kind === "job"
        ? watchJob(id.trim(), signal)
        : loadGraph(id.trim(), signal),
    );
  }

  function stopWaiting() {
    work.current?.abort();
    setBusy(false);
  }
  function showExample() {
    work.current?.abort();
    filter.current?.abort();
    setGraph(exampleGraph);
    setViewGraph(exampleGraph);
    setSource("example");
    setBusy(false);
    setMode("all");
    setJob(null);
    setError("");
    setPendingJobId("");
    setFiltering(false);
    const url = new URL(location.href);
    url.searchParams.delete("job");
    url.searchParams.delete("graph");
    history.replaceState(null, "", url);
  }
  async function changeMode(next: ModeFilter) {
    filter.current?.abort();
    const controller = new AbortController();
    filter.current = controller;
    setMode(next);
    setError("");
    if (source === "example") {
      setViewGraph(
        next === "all"
          ? graph
          : {
              ...graph,
              nodes: graph.nodes.filter((n) => n.modes.includes(next)),
              edges: graph.edges.filter((e) => e.mode === next),
            },
      );
      return;
    }
    setFiltering(true);
    try {
      const result = await getGraph(graph.id, controller.signal, next);
      if (!controller.signal.aborted) setViewGraph(result);
    } catch (cause) {
      if (!controller.signal.aborted) {
        setMode("all");
        setViewGraph(graph);
        setError(cause instanceof Error ? cause.message : "Ошибка фильтрации");
      }
    } finally {
      if (!controller.signal.aborted) setFiltering(false);
    }
  }
  return {
    graph,
    viewGraph,
    source,
    regions,
    health,
    busy,
    filtering,
    job,
    error,
    mode,
    pendingJobId,
    build,
    open,
    stopWaiting,
    showExample,
    changeMode,
    checkConnection,
    setError,
  };
}
