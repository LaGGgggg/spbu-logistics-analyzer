import { MapView } from '@/components/map/MapView'
import { AppLayout } from './components/layout/AppLayout'
import { Header } from './header/Header'
import { BuildPanel } from './build/BuildPanel'
import { useCallback, useState } from 'react';
import type { BBox } from './components/map/BBoxDrawer';

function App() {
  const graphId: string | undefined = undefined;
  const [bbox, setBBox] = useState<BBox | null>(null)
  const [drawMode, setDrawMode] = useState(false)

  const handleDrawn = useCallback((b: BBox) => {
    setBBox(b)
    setDrawMode(false)
  }, [])
  const cancelDraw = useCallback(() => setDrawMode(false), [])


  return (
    <AppLayout
      header={<Header graphId={graphId} />}
      left={
        <BuildPanel
          bbox={bbox}
          drawMode={drawMode}
          onToggleDraw={() => setDrawMode((v) => !v)}
          onClearBBox={() => setBBox(null)}
        />
      }
      right={<div className="p-4">Статистика / Метрики / Узел</div>}
    >
      <MapView bbox={bbox} drawMode={drawMode} onDrawn={handleDrawn} onDrawCancel={cancelDraw} />
    </AppLayout>
  )
}

export default App
