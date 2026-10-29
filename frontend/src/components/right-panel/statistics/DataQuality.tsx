import type { Coverage } from "@/api/types";

type DataQualityProps = {
    coverage: Coverage | null;
};

export function DataQuality({ coverage }: DataQualityProps) {
    return (
        <div className="mt-6 border-t pt-4">
            <h3 className="mb-3 text-sm font-semibold">
                Качество данных
            </h3>

            <div className="grid grid-cols-2 gap-x-4 gap-y-2 text-xs">
                <div className="flex justify-between gap-2">
                    <span className="text-muted-foreground">Кандидатов</span>
                    <span className="font-mono">
                        {coverage
                            ? coverage.candidates.toLocaleString("ru-RU")
                            : "—"}
                    </span>
                </div>

                <div className="flex justify-between gap-2">
                    <span className="text-muted-foreground">Принято</span>
                    <span className="font-mono">
                        {coverage
                            ? coverage.accepted.toLocaleString("ru-RU")
                            : "—"}
                    </span>
                </div>

                <div className="flex justify-between gap-2">
                    <span className="text-muted-foreground">Отклонено</span>
                    <span className="font-mono">
                        {coverage
                            ? coverage.rejected.toLocaleString("ru-RU")
                            : "—"}
                    </span>
                </div>

                <div className="flex justify-between gap-2">
                    <span className="text-muted-foreground">
                        Предупреждений
                    </span>
                    <span className="font-mono">
                        {coverage
                            ? (coverage.warnings?.length ?? 0).toLocaleString("ru-RU")
                            : "—"}
                    </span>
                </div>
            </div>

            {(coverage?.warnings?.length ?? 0) > 0 && (
                <div className="mt-3 rounded-md border border-amber-200 bg-amber-50 p-3 text-amber-950">
                    <p className="mb-2 text-xs font-semibold">
                        Предупреждения источника
                    </p>

                    <ul className="list-disc space-y-1 pl-4 text-xs">
                        {(coverage?.warnings ?? []).map((warning, index) => (
                            <li key={`${index}-${warning}`}>
                                {warning}
                            </li>
                        ))}
                    </ul>
                </div>
            )}
        </div>
    );
}