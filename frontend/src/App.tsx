import { MapView } from '@/components/map/MapView'
import { AppLayout } from './components/layout/AppLayout'
import { Header } from './header/Header'
import { BuildPanel } from './build/BuildPanel'
import { useCallback, useMemo, useRef, useState } from 'react';
import { LayersPanel } from './build/LayersPanel'
import { MODES } from './lib/modes'
import type { BBox } from './components/map/BBoxDrawer';
import { buildGraph, getGraph } from './api/graph';
import axios from 'axios';
import { getApiErrorMessage } from './api/errors';
import type { LogisticsGraph, TransportMode } from './api/types';
import { UNKNOWN_KIND } from './lib/kinds'
import { RightPanel } from "@/components/right-panel/RightPanel";


function App() {
  const [bbox, setBBox] = useState<BBox | null>(null)
  const [drawMode, setDrawMode] = useState(false)
  const [graph, setGraph] = useState<LogisticsGraph | null>(null)
  const [building, setBuilding] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const abortRef = useRef<AbortController | null>(null)
  const [visibleModes, setVisibleModes] = useState<Set<TransportMode>>(() => new Set(MODES))
  const [hiddenKinds, setHiddenKinds] = useState<Set<string>>(() => new Set())
  const [showEdges, setShowEdges] = useState(true)


  const toggleMode = useCallback((mode: TransportMode) => {
    setVisibleModes((prev) => {
      const next = new Set(prev)
      if (next.has(mode)) next.delete(mode)
      else next.add(mode)
      return next
    })
  }, [])

  const toggleKind = useCallback((kind: string) => {
    setHiddenKinds((prev) => {
      const next = new Set(prev)
      if (next.has(kind)) next.delete(kind)
      else next.add(kind)
      return next
    })
  }, [])

  const kindCounts = useMemo(() => {
    const counts = new Map<string, number>()
    for (const node of graph?.nodes ?? []) {
      const kind = node.kind ?? UNKNOWN_KIND
      counts.set(kind, (counts.get(kind) ?? 0) + 1)
    }
    return [...counts.entries()].sort((a, b) => b[1] - a[1])
  }, [graph])

  const modeCounts = useMemo(() => {
    const counts = { rail: 0, sea: 0, river: 0 } as Record<TransportMode, number>
    for (const node of graph?.nodes ?? []) {
      for (const m of node.modes) if (m in counts) counts[m]++
    }
    return counts
  }, [graph])



  async function handleBuild() {
    if (!bbox) return
    abortRef.current?.abort()
    const controller = new AbortController()
    abortRef.current = controller


    setBuilding(true)
    setError(null)

    try {
      const graphId = await buildGraph(bbox, controller.signal)
      setGraph(await getGraph(graphId, controller.signal))
    }
    catch (e) {
      if (axios.isCancel(e) || (e instanceof DOMException && e.name === 'AbortError')) return
      setError(e instanceof Error && !axios.isAxiosError(e) ? e.message : getApiErrorMessage(e))
    } finally {
      if (abortRef.current === controller) setBuilding(false)
    }
  }

  const handleDrawn = useCallback((b: BBox) => {
    setBBox(b)
    setDrawMode(false)
  }, [])
  const cancelDraw = useCallback(() => setDrawMode(false), [])


  return (
    <AppLayout
      header={<Header graphId={graph?.id} />}
      left={
        <>
          <BuildPanel
            bbox={bbox}
            drawMode={drawMode}
            building={building}
            error={error}
            onToggleDraw={() => setDrawMode((v) => !v)}
            onClearBBox={() => setBBox(null)}
            onBuild={handleBuild}
          />
          {graph && (
            <LayersPanel
              visibleModes={visibleModes}
              counts={modeCounts}
              onToggleMode={toggleMode}
              kindCounts={kindCounts}
              hiddenKinds={hiddenKinds}
              onToggleKind={toggleKind}
              showEdges={showEdges}
              onShowEdgesChange={setShowEdges}
            />
          )}
        </>
      }
      right={<RightPanel />}
    >
      <MapView
        bbox={bbox}
        graph={graph}
        visibleModes={visibleModes}
        drawMode={drawMode}
        onDrawn={handleDrawn}
        onDrawCancel={cancelDraw}
        hiddenKinds={hiddenKinds}
        showEdges={showEdges}

      />


    </AppLayout>
  )
}

export default App
