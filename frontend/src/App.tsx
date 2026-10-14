import { useState } from 'react'
import { MapView, type MapRegion } from '@/components/map/MapView'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'

const TEST_REGION: MapRegion = {
  id: 'test-len',
  name: 'Ленинградская область',
  geometry: {
    type: 'Polygon',
    coordinates: [[[27.7, 58.4], [35.7, 58.4], [35.7, 61.4], [27.7, 61.4], [27.7, 58.4]]],
  } as GeoJSON.Polygon,
}

function App() {
  const [query, setQuery] = useState('')
  const [region, setRegion] = useState<MapRegion | null>(null)

  function handleSearch() {
    const cleanQuery = query.trim()
    if (cleanQuery === '') {
      alert('Введите название региона')
      return
    }
    setRegion(TEST_REGION)
  }

  return (
    <div className="flex h-svh">

      <aside className="dark flex w-80 shrink-0 flex-col border-r bg-background text-foreground">
        <header className="p-5">
          <h1 className="text-2xl mb-4">Анализатор логистики</h1>
          <p className="text-sm text-muted-foreground">
            Система анализа транспортной инфраструктуры
          </p>
        </header>

        <div className="flex flex-1 flex-col gap-5 overflow-y-auto p-5">
          <form
            className="space-y-2"
            onSubmit={(event) => {
              event.preventDefault()
              handleSearch()
            }}
          >
            <div className="flex gap-2">
              <Input
                id="region-search"
                className="h-10 rounded-none outline-none"
                placeholder="Введите регион"
                value={query}
                onChange={(event) => setQuery(event.target.value)}
              />
              <Button type="submit" className="h-10 px-4 rounded-none">
                Найти
              </Button>
            </div>
          </form>

          {region && (
            <section className="space-y-3 border bg-card p-4">
              <div>
                <p className="text-xs uppercase tracking-wide text-muted-foreground">
                  Выбранный регион
                </p>
                <p className="mt-1 font-medium">{region.name}</p>
              </div>
              <Button
                variant="outline"
                className="h-9 w-full rounded-none"
                onClick={() => setRegion(null)}
              >
                Сбросить
              </Button>
            </section>
          )}
        </div>


      </aside>

      <main className="isolate flex-1">
        <MapView region={region} />
      </main>
    </div>
  )
}

export default App
