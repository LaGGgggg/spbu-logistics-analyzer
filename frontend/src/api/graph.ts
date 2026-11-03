import { apiClient } from "./client";
import type { Bounds, GraphMetrics, Job, LogisticsGraph, TransportMode } from './types'


const POLL_INTERVAL = 2000
const METRICS_TIMEOUT = 60000
const GRAPH_TIMEOUT = 120000
const MAX_BUILD_ATTEMPTS = 3
const FAST_FAILURE_WINDOW = 15000
const RETRY_DELAY = 3000

function sleep(ms: number, signal: AbortSignal) {
    return new Promise<void>((resolve, reject) => {
        const t = setTimeout(resolve, ms)
        signal.addEventListener('abort', () => {
            clearTimeout(t)
            reject(new DOMException('Aborted', 'AbortError'))
        })
    })
}

async function waitForJob(jobId: string, signal: AbortSignal): Promise<Job> {
    while (true) {
        await sleep(POLL_INTERVAL, signal)
        const { data: job } = await apiClient.get<Job>(`/jobs/${jobId}`, { signal })
        if (job.status === 'completed' || job.status === 'failed' || job.status === 'cancelled') return job
    }
}

export async function buildGraph(bounds: Bounds, signal: AbortSignal, onJob?: (jobId: string) => void): Promise<string> {
    for (let attempt = 1; ; attempt++) {
        const startedAt = Date.now()
        const res = await apiClient.post<{ graph_id?: string; job_id?: string }>(
            '/graphs', { bounds }, { signal }
        )

        if (res.status === 200 && res.data.graph_id) return res.data.graph_id

        const jobId = res.data.job_id!
        onJob?.(jobId)

        const job = await waitForJob(jobId, signal)
        if (job.status === 'completed') return job.graph_id!
        if (job.status === 'cancelled') throw new Error('Построение отменено')

        const failedFast = Date.now() - startedAt < FAST_FAILURE_WINDOW
        if (!failedFast || attempt >= MAX_BUILD_ATTEMPTS) throw new Error(job.error ?? 'Не удалось построить граф')

        await sleep(RETRY_DELAY, signal)
    }
}

export async function cancelJob(id: string) {
    await apiClient.post(`/jobs/${id}/cancel`)
}

export async function getGraphMetrics(id: string, mode: TransportMode, signal?: AbortSignal) {
    const { data } = await apiClient.get<GraphMetrics>(`/graphs/${id}/metrics`, {
        params: { mode },
        signal,
        timeout: METRICS_TIMEOUT,
    })
    return data
}

export async function getGraph(id: string, signal?: AbortSignal) {
    const { data } = await apiClient.get<LogisticsGraph>(`/graphs/${id}`, { signal, timeout: GRAPH_TIMEOUT })
    return data
}