import { useQuotes } from './hooks/useQuotes'
import { useSymbolCatalog } from './hooks/useSymbolCatalog'
import { WorldMap } from './components/Map/WorldMap'
import { Legend } from './components/Legend/Legend'
import './App.css'

function App() {
  const { quotes, loading: quotesLoading, error: quotesError } = useQuotes()
  const { symbols, loading: symbolsLoading, error: symbolsError } = useSymbolCatalog()

  const loading = quotesLoading || symbolsLoading
  const error = quotesError ?? symbolsError

  return (
    <main>
      <h1>Mapa de Inversiones</h1>

      {error && <p className="status status--error">error: {error}</p>}
      {loading && !error && <p className="status">loading…</p>}

      {!loading && !error && (
        <>
          <Legend />
          <WorldMap quotes={quotes} symbols={symbols} />

          <table>
            <thead>
              <tr>
                <th>Symbol</th>
                <th>Price</th>
                <th>Change</th>
                <th>% Change</th>
                <th>Market</th>
                <th>Last Updated</th>
              </tr>
            </thead>
            <tbody>
              {quotes.map((q) => (
                <tr key={q.symbol}>
                  <td>{q.symbol}</td>
                  <td>
                    {q.price.toLocaleString(undefined, { maximumFractionDigits: 2 })} {q.currency}
                  </td>
                  <td className={q.change >= 0 ? 'positive' : 'negative'}>{q.change.toFixed(2)}</td>
                  <td className={q.changePercent >= 0 ? 'positive' : 'negative'}>
                    {q.changePercent.toFixed(2)}%
                  </td>
                  <td>{q.marketState}</td>
                  <td>{new Date(q.lastUpdated).toLocaleTimeString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </>
      )}
    </main>
  )
}

export default App
