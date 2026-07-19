import { useQuotes } from './hooks/useQuotes'
import './App.css'

// Temporary proof-of-pipeline view: a plain table of every tracked quote,
// polled from the Go backend. This gets replaced by the D3 choropleth map,
// but confirms end-to-end data flow (Yahoo -> Redis -> REST API -> React)
// before any map/projection code is written.
function App() {
  const { quotes, loading, error } = useQuotes()

  if (loading) return <p className="status">loading quotes…</p>
  if (error) return <p className="status status--error">error: {error}</p>

  return (
    <main>
      <h1>Mapa de Inversiones</h1>
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
    </main>
  )
}

export default App
