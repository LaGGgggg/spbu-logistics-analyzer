import { useEffect, useState } from "react";
import { DownloadIcon, XIcon } from "lucide-react";
import { Button } from "../ui/button";
import { downloadGraph } from "@/api/exportGraph";

type ExportProps = {
    graphId?: string;
};

const exportFormats = [
    { value: "nodes.csv", label: "CSV узлов", extension: ".csv" },
    { value: "edges.csv", label: "CSV рёбер", extension: ".csv" },
    { value: "adjacency.csv", label: "Матрица смежности", extension: ".csv" },
    { value: "incidence.csv", label: "Матрица инцидентности", extension: ".csv" },
    { value: "geojson", label: "GeoJSON", extension: ".geojson" },
    { value: "graphml", label: "GraphML", extension: ".graphml" },
];

export function Export({ graphId }: ExportProps) {
    const [isOpen, setIsOpen] = useState(false);
    const [isMounted, setIsMounted] = useState(false);
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
                onClick={() => {
                    setIsMounted(true);
                    setIsOpen(true);
                }}
            >
                <DownloadIcon />
                Экспорт
            </Button>

            {isMounted && (
                <div
                    className={`fixed inset-0 z-[1000] flex items-center justify-center bg-black/50 p-4 duration-200 [animation-fill-mode:forwards] ${
                        isOpen ? "animate-in fade-in" : "animate-out fade-out"
                    }`}
                    onAnimationEnd={(event) => {
                        if (event.target === event.currentTarget && !isOpen) setIsMounted(false);
                    }}
                >
                    <div
                        className={`w-full max-w-2xl rounded-sm border bg-background p-6 text-foreground duration-200 [animation-fill-mode:forwards] ${
                            isOpen
                                ? "animate-in fade-in zoom-in-95 slide-in-from-bottom-2"
                                : "animate-out fade-out zoom-out-95 slide-out-to-bottom-2"
                        }`}
                    >
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
                            {exportFormats.map((item) => (
                                <button
                                    key={item.value}
                                    type="button"
                                    onClick={() => setFormat(item.value)}
                                    className={`rounded-sm border-2 p-4 text-left ${
                                        format === item.value
                                            ? "border-blue-700"
                                            : "border-border"
                                    }`}
                                >
                                    <div className="flex items-center justify-between gap-3">
                                        <span className="font-semibold">
                                            {item.label}
                                        </span>
                                        <span className="text-sm text-muted-foreground">
                                            {item.extension}
                                        </span>
                                    </div>
                                </button>
                            ))}
                        </div>

                        {!graphId && (
                            <p className="mt-4 text-sm text-muted-foreground">
                                Для скачивания сначала постройте граф.
                            </p>
                        )}

                        <div className="mt-6 flex justify-end gap-3 border-t pt-4">
                            <Button
                                className={"h-10"}
                                variant="outline"
                                onClick={() => setIsOpen(false)}
                            >
                                Отмена
                            </Button>

                            <Button
                                className={"h-10"}
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