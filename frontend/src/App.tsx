
import { MapView } from '@/components/map/MapView'
import { AppLayout } from './components/layout/AppLayout'
import { Header } from './header/Header'
import { BuildPanel } from './build/BuildPanel'



function App() {

  return (
    <AppLayout
      header={<Header />}
      left={<BuildPanel />}
      right={<div className="p-4">Статистика / Метрики / Узел</div>}
    >
      <MapView region={null} />
    </AppLayout>
  )
}

export default App
