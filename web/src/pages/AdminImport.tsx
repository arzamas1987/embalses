import { useState } from 'react'
import { useAllReservoirs } from '../hooks/useQueries'
import { importReadingsCSV } from '../api/client'
import type { ReservoirSummary } from '../types'

const DownloadIcon = () => (
  <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
  </svg>
)

const UploadIcon = () => (
  <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
  </svg>
)

function downloadTemplate() {
  const csv = 'reservoir_slug,observed_at,volume_hm3,capacity_hm3,fill_pct\nembalse-de-el-pardo,2026-06-29,5.97,12.43,48\n'
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'importar-lecturas-template.csv'
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

export default function AdminImport() {
  const { data } = useAllReservoirs()
  const reservoirs = (data?.data as ReservoirSummary[] | undefined) ?? []
  const missing = reservoirs.filter((r) => r.latest_fill_pct == null)

  const [file, setFile] = useState<File | null>(null)
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<{ imported: number } | null>(null)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!file) return
    setLoading(true)
    setError(null)
    setResult(null)
    try {
      const res = await importReadingsCSV(file)
      setResult(res.data ?? null)
      setFile(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Error desconocido')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="animate-fade-in max-w-4xl mx-auto">
      <h1 className="text-2xl font-bold text-[#0f172a] mb-2">Importar lecturas</h1>
      <p className="text-[#475569] text-sm mb-6">
        Sube un CSV con lecturas que no estén disponibles en MITECO (por ejemplo, embalses pequeños o datos de otras fuentes).
      </p>

      <div className="grid md:grid-cols-2 gap-6 mb-8">
        <div className="gov-card p-5">
          <h2 className="font-bold text-[#0f172a] mb-3">Instrucciones</h2>
          <p className="text-sm text-[#475569] mb-4">
            El CSV debe tener cabecera y al menos estas columnas:
          </p>
          <code className="block bg-[#f1f5f9] p-3 rounded-lg text-xs text-[#0f172a] mb-4">
            reservoir_slug,observed_at,volume_hm3,capacity_hm3,fill_pct
          </code>
          <ul className="text-sm text-[#475569] space-y-1.5 list-disc list-inside">
            <li><strong>reservoir_slug</strong>: slug del embalse en la base de datos.</li>
            <li><strong>observed_at</strong>: fecha en formato YYYY-MM-DD.</li>
            <li><strong>volume_hm3</strong> + <strong>capacity_hm3</strong>: se calculará el %.</li>
            <li>O puedes usar directamente <strong>fill_pct</strong>.</li>
          </ul>
          <button
            onClick={downloadTemplate}
            className="gov-btn gov-btn-outline text-sm mt-5"
          >
            <DownloadIcon />
            Descargar plantilla
          </button>
        </div>

        <div className="gov-card p-5">
          <h2 className="font-bold text-[#0f172a] mb-3">Subir CSV</h2>
          <form onSubmit={handleSubmit} className="space-y-4">
            <label className="flex flex-col items-center justify-center w-full h-32 border-2 border-dashed border-[#e2e8f0] rounded-lg cursor-pointer hover:bg-[#f8fafc] transition-colors">
              <div className="flex flex-col items-center justify-center pt-5 pb-6">
                <UploadIcon />
                <p className="text-sm text-[#475569] mt-2">
                  {file ? file.name : 'Arrastra un CSV o haz clic para seleccionar'}
                </p>
              </div>
              <input
                type="file"
                accept=".csv,text/csv"
                className="hidden"
                onChange={(e) => setFile(e.target.files?.[0] ?? null)}
              />
            </label>
            <button
              type="submit"
              disabled={!file || loading}
              className="gov-btn w-full justify-center disabled:opacity-50"
            >
              {loading ? (
                <>
                  <svg className="animate-spin h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                  </svg>
                  Importando...
                </>
              ) : (
                'Importar lecturas'
              )}
            </button>
            {result && (
              <div className="p-3 bg-green-50 text-green-800 rounded-lg text-sm">
                Se importaron <strong>{result.imported}</strong> lecturas correctamente.
              </div>
            )}
            {error && (
              <div className="p-3 bg-red-50 text-red-800 rounded-lg text-sm">
                Error: {error}
              </div>
            )}
          </form>
        </div>
      </div>

      <div className="gov-card p-5">
        <h2 className="font-bold text-[#0f172a] mb-3">
          Embalses sin datos ({missing.length})
        </h2>
        <p className="text-sm text-[#475569] mb-3">
          Estos embalse no tienen lecturas tras importar MITECO. Puedes buscar sus datos en otras fuentes y subirlos mediante CSV.
        </p>
        <div className="max-h-64 overflow-auto border border-[#e2e8f0] rounded-lg">
          <table className="gov-table">
            <thead className="sticky top-0 bg-white">
              <tr>
                <th>Embalse</th>
                <th>Cuenca</th>
                <th>Slug</th>
              </tr>
            </thead>
            <tbody>
              {missing.map((r) => (
                <tr key={r.id}>
                  <td>{r.name}</td>
                  <td>{r.basin_name}</td>
                  <td><code className="text-xs">{r.slug}</code></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
