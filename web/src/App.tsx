import { useEffect, useState } from 'react'
import './App.css'

function App() {
  const [backendStatus, setBackendStatus] = useState<string>('checking...')

  useEffect(() => {
    fetch('/api/healthz')
      .then((res) => res.json())
      .then((data) => {
        if (data.status === 'ok') {
          setBackendStatus('connected')
        } else {
          setBackendStatus('unhealthy')
        }
      })
      .catch(() => setBackendStatus('unreachable'))
  }, [])

  return (
    <div className="app">
      <header>
        <h1>Embalses MVP</h1>
        <p className="subtitle">Spanish reservoir data platform</p>
      </header>
      <main>
        <div className="status-card">
          <h2>Backend Status</h2>
          <p className={`status status-${backendStatus}`}>
            {backendStatus === 'connected' ? '✅ Connected' : backendStatus === 'unhealthy' ? '⚠️ Unhealthy' : '❌ Unreachable'}
          </p>
        </div>
      </main>
      <footer>
        <p>Local-first MVP · Phase 0</p>
      </footer>
    </div>
  )
}

export default App
