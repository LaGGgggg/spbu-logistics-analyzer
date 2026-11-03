import type { TransportMode } from '@/api/types'

export const MODES: TransportMode[] = ['rail', 'sea', 'river']

export const MODE_COLOR: Record<TransportMode, string> = {
    rail: '#2563eb',
    sea: '#0891b2',
    river: '#16a34a',
}

export const MODE_LABEL: Record<TransportMode, string> = {
    rail: 'Ж/д',
    sea: 'Море',
    river: 'Река',
}
