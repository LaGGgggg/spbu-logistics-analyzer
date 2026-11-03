import { apiClient } from "./client";

const EXPORT_TIMEOUT = 120000;

export async function downloadGraph(graphId: string, format: string) {
    const response = await apiClient.get(
        `/graphs/${encodeURIComponent(graphId)}/export/${format}`,
        { responseType: "blob", timeout: EXPORT_TIMEOUT }
    );

    const url = URL.createObjectURL(response.data);
    const link = document.createElement("a");

    link.href = url;
    link.download = format.includes(".") ? `graph-${format}` : `graph.${format}`;
    document.body.appendChild(link);
    link.click();
    link.remove();

    window.setTimeout(() => URL.revokeObjectURL(url), 1000);
}