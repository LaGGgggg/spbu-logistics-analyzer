import { useEffect, useState } from "react";
import { DownloadIcon, XIcon } from "lucide-react";
import { Button } from "../ui/button";
import { downloadGraph } from "@/api/exportGraph";

type ExportProps = {
    graphId?: string;
};

export function Export({ graphId }: ExportProps) {
    const [isOpen, setIsOpen] = useState(false);
    const [format, setFormat] = useState("nodes.csv");
    const [isDownloading, setIsDownloading] = useState(false);

    useEffect(() => {
    if (!isOpen) return;

    function handleKeyDown(event: KeyboardEvent) {
            if (event.key === "Escape") {
                setIsOpen(false);
            }
        }

        document.addEventListener("keydown", handleKeyDown);

        return () => {
            document.removeEventListener("keydown", handleKeyDown);
        };
    }, [isOpen]);

    async function handleDownload() {
        if (!graphId || isDownloading) return;

        setIsDownloading(true);

        try {
            await downloadGraph(graphId, format);
        } catch {
            alert("Не удалось скачать файл. Попробуйте ещё раз.");
        } finally {
            setIsDownloading(false);
        }
    }  

    return (
        <>
            <Button
                className="h-10"
                onClick={() => setIsOpen(true)}
            >
                <DownloadIcon />
                Экспорт
            </Button>

            {isOpen && (
                <div className="fixed inset-0 z-[1000] flex items-center justify-center bg-black/50 p-4">
                    <div className="w-full max-w-3xl rounded-lg border bg-background p-6 text-foreground">
                        <div className="mb-6 flex items-start justify-between gap-4">
                            <div>
                                <h2 className="text-xl font-semibold">
                                    Экспорт графа
                                </h2>

                                <p className="mt-1 text-sm text-muted-foreground">
                                    Выберите формат файла.
                                </p>
                            </div>

                            <button
                                type="button"
                                onClick={() => setIsOpen(false)}
                                aria-label="Закрыть окно"
                                className="rounded p-1 hover:bg-muted"
                            >
                                <XIcon className="h-5 w-5" />
                            </button>
                        </div>

                        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                            <button
                                type="button"
                                onClick={() => setFormat("nodes.csv")}
                                className={`rounded-xl border-2 p-4 text-left ${
                                    format === "nodes.csv"
                                        ? "border-blue-700"
                                        : "border-border"
                                }`}
                            >
                                <div className="flex items-center justify-between gap-3">
                                    <span className="font-semibold">CSV узлов</span>
                                    <span className="text-sm text-muted-foreground">
                                        .csv
                                    </span>
                                </div>
                            </button>

                            <button
                                type="button"
                                onClick={() => setFormat("edges.csv")}
                                className={`rounded-xl border-2 p-4 text-left ${
                                    format === "edges.csv"
                                        ? "border-blue-700"
                                        : "border-border"
                                }`}
                            >
                                <div className="flex items-center justify-between gap-3">
                                    <span className="font-semibold">CSV рёбер</span>
                                    <span className="text-sm text-muted-foreground">
                                        .csv
                                    </span>
                                </div>
                            </button>

                            <button
                                type="button"
                                onClick={() => setFormat("adjacency.csv")}
                                className={`rounded-xl border-2 p-4 text-left ${
                                    format === "adjacency.csv"
                                        ? "border-blue-700"
                                        : "border-border"
                                }`}
                            >
                                <div className="flex items-center justify-between gap-3">
                                    <span className="font-semibold">Матрица смежности</span>
                                    <span className="text-sm text-muted-foreground">.csv</span>
                                </div>
                            </button>

                            <button
                                type="button"
                                onClick={() => setFormat("incidence.csv")}
                                className={`rounded-xl border-2 p-4 text-left ${
                                    format === "incidence.csv"
                                        ? "border-blue-700"
                                        : "border-border"
                                }`}
                            >
                                <div className="flex items-center justify-between gap-3">
                                    <span className="font-semibold">Матрица инцидентности</span>
                                    <span className="text-sm text-muted-foreground">.csv</span>
                                </div>
                            </button>

                            <button
                                type="button"
                                onClick={() => setFormat("geojson")}
                                className={`rounded-xl border-2 p-4 text-left ${
                                    format === "geojson"
                                        ? "border-blue-700"
                                        : "border-border"
                                }`}
                            >
                                <div className="flex items-center justify-between gap-3">
                                    <span className="font-semibold">GeoJSON</span>
                                    <span className="text-sm text-muted-foreground">.geojson</span>
                                </div>
                            </button>

                            <button
                                type="button"
                                onClick={() => setFormat("graphml")}
                                className={`rounded-xl border-2 p-4 text-left ${
                                    format === "graphml"
                                        ? "border-blue-700"
                                        : "border-border"
                                }`}
                            >
                                <div className="flex items-center justify-between gap-3">
                                    <span className="font-semibold">GraphML</span>
                                    <span className="text-sm text-muted-foreground">.graphml</span>
                                </div>
                            </button>
                        </div>

                        {!graphId && (
                            <p className="mt-4 text-sm text-muted-foreground">
                                Для скачивания сначала постройте граф.
                            </p>
                        )}

                        <div className="mt-6 flex justify-end gap-3 border-t pt-4">
                            <Button
                                variant="outline"
                                onClick={() => setIsOpen(false)}
                            >
                                Отмена
                            </Button>

                            <Button
                                disabled={!graphId || isDownloading}
                                onClick={handleDownload}
                            >
                                <DownloadIcon className="h-4 w-4" />
                                {isDownloading ? "Подготавливаем…" : "Скачать"}
                            </Button>
                        </div>
                    </div>
                </div>
            )}
        </>
    );
}