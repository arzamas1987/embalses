import { useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'

export default function Layout({ children }: { children: React.ReactNode }) {
  const { t } = useTranslation()
  const [menuOpen, setMenuOpen] = useState(false)
  const location = useLocation()

  const links = [
    { to: '/', label: t('nav.home') },
    { to: '/mapa', label: t('nav.map') },
    { to: '/embalses', label: t('nav.reservoirs') },
    { to: '/comparar', label: t('nav.compare') },
    { to: '/cuencas', label: t('nav.basins') },
    { to: '/fuentes', label: t('nav.sources') },
    { to: '/calidad-datos', label: t('nav.dataQuality') },
    { to: '/ajustes', label: t('nav.settings') },
  ]

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col">
      <header className="bg-blue-700 text-white shadow">
        <div className="max-w-6xl mx-auto px-4 py-3 flex items-center justify-between">
          <Link to="/" className="text-xl font-bold tracking-tight">
            {t('appName')}
          </Link>
          <button
            className="md:hidden p-2"
            onClick={() => setMenuOpen(!menuOpen)}
            aria-label="Toggle menu"
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
            </svg>
          </button>
          <nav className="hidden md:flex gap-6 text-sm">
            {links.map((l) => (
              <Link
                key={l.to}
                to={l.to}
                className={`hover:underline ${location.pathname === l.to ? 'font-semibold' : ''}`}
              >
                {l.label}
              </Link>
            ))}
          </nav>
        </div>
        {menuOpen && (
          <nav className="md:hidden border-t border-blue-600 px-4 pb-3">
            {links.map((l) => (
              <Link
                key={l.to}
                to={l.to}
                className={`block py-2 ${location.pathname === l.to ? 'font-semibold' : ''}`}
                onClick={() => setMenuOpen(false)}
              >
                {l.label}
              </Link>
            ))}
          </nav>
        )}
      </header>

      <main className="flex-1 max-w-6xl mx-auto w-full px-4 py-6">
        {children}
      </main>

      <footer className="bg-gray-100 text-gray-600 text-sm text-center py-4">
        <p>
          {t('appName')} &mdash; Fuente: MITECO
        </p>
      </footer>
    </div>
  )
}
