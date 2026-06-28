import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { QueryClientProvider } from '@tanstack/react-query'
import { queryClient } from './api/client'
import './i18n'
import Layout from './components/Layout'
import Home from './pages/Home'
import MapPage from './pages/MapPage'
import Reservoirs from './pages/Reservoirs'
import ReservoirDetail from './pages/ReservoirDetail'
import Comparator from './pages/Comparator'
import Basins from './pages/Basins'
import Sources from './pages/Sources'
import DataQuality from './pages/DataQuality'
import Settings from './pages/Settings'
import NotFound from './pages/NotFound'

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Layout>
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/mapa" element={<MapPage />} />
            <Route path="/embalses" element={<Reservoirs />} />
            <Route path="/embalses/:slug" element={<ReservoirDetail />} />
            <Route path="/comparar" element={<Comparator />} />
            <Route path="/cuencas" element={<Basins />} />
            <Route path="/fuentes" element={<Sources />} />
            <Route path="/calidad-datos" element={<DataQuality />} />
            <Route path="/ajustes" element={<Settings />} />
            <Route path="*" element={<NotFound />} />
          </Routes>
        </Layout>
      </BrowserRouter>
    </QueryClientProvider>
  )
}
