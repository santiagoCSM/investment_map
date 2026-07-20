import { useQuotes } from './hooks/useQuotes'
import { useSymbolCatalog } from './hooks/useSymbolCatalog'
import { WorldMap } from './components/Map/WorldMap'
import { DataPanel } from './components/DataPanel/DataPanel'
import { Legend } from './components/Legend/Legend'
import { ThemeToggle } from './components/ThemeToggle/ThemeToggle'

// The map is the product, so it gets top billing, but it's laid out
// normally now (a flex column: map area, then legend row) rather than
// pinned full-viewport with everything floating on top of it — that was
// letting the legend overlap map content. DataPanel stays a floating,
// draggable overlay on top of the map area; the legend is a plain block
// row with its own reserved space below it.
function App() {
  const { quotes, loading: quotesLoading, error: quotesError } = useQuotes()
  const { symbols, loading: symbolsLoading, error: symbolsError } = useSymbolCatalog()

  return (
    <div className="app-shell relative flex h-dvh flex-col overflow-hidden">
      <div className="map-area relative min-h-0 flex-1">
        <WorldMap quotes={quotes} symbols={symbols} />
        <ThemeToggle />
        <DataPanel
          quotes={quotes}
          symbols={symbols}
          loading={quotesLoading || symbolsLoading}
          error={quotesError ?? symbolsError}
        />
      </div>
      <Legend />
    </div>
  )
}

export default App
