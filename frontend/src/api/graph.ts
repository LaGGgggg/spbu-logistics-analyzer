import { apiClient } from "./client";
import type { Bounds, Job, LogisticsGraph } from './types'


const POLL_INTERVAL = 2000

function sleep(ms: number, signal: AbortSignal) {
    return new Promise<void>((resolve, reject) => {
        const t = setTimeout(resolve, ms)
        signal.addEventListener('abort', () => {
            clearTimeout(t)
            reject(new DOMException('Aborted', 'AbortError'))
        })
    })
}

export async function buildGraph(bounds: Bounds, signal: AbortSignal): Promise<string> {
    const res = await apiClient.post<{ graph_id?: string; job_id?: string }>(
        '/graphs', { bounds }, { signal }
    )

    if (res.status === 200 && res.data.graph_id) return res.data.graph_id

    const jobId = res.data.job_id!
    while (true) {
        await sleep(POLL_INTERVAL, signal)
        const { data: job } = await apiClient.get<Job>(`/jobs/${jobId}`, { signal })
        if (job.status === 'completed') return job.graph_id!
        if (job.status === 'failed') throw new Error(job.error ?? 'Не удалось построить граф')
    }
}


export async function getGraph(id: string, signal?: AbortSignal) {
    const { data } = await apiClient.get<LogisticsGraph>(`/graphs/${id}`, { signal })
    return data
}