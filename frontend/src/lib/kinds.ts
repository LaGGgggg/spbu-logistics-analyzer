export const KIND_LABEL: Record<string, string> = {
    port: 'Порт',
    freight_terminal: 'Грузовой терминал',
    freight_facility: 'Грузовой объект',
    cargo_airport: 'Грузовой аэропорт',
    warehouse: 'Склад',
}

export const UNKNOWN_KIND = 'other'

export function kindLabel(kind: string) {
    if (kind === UNKNOWN_KIND) return 'Прочее'
    return KIND_LABEL[kind] ?? kind
}