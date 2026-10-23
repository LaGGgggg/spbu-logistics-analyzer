import { apiClient } from "./client";

export async function downloadGraph(graphId: string, format: string) {
    const response = await apiClient.get(
        `/graphs/${encodeURIComponent(graphId)}/export/${format}`,
        { responseType: "blob" }
    );

    const url = URL.createObjectURL(response.data);
    const link = document.createElement("a");

    link.href = url;
    link.download = `graph-${format}`;
    document.body.appendChild(link);
    link.click();
    link.remove();

    window.setTimeout(() => URL.revokeObjectURL(url), 1000);
}