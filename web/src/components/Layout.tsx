import { useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'

const LogoIcon = () => (
  <svg width="32" height="32" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M12 2.69l5.66 5.66a8 8 0 1 1-11.31 0L12 2.69z" fill="currentColor" fillOpacity="0.2"/>
    <path d="M12 2.69l5.66 5.66a8 8 0 1 1-11.31 0L12 2.69z" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
    <path d="M12 8v8M8 12h8" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"/>
  </svg>
)

const MenuIcon = () => (
  <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
  </svg>
)

const CloseIcon = () => (
  <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
  </svg>
)

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
  ]

  const isActive = (path: string) => {
    if (path === '/') return location.pathname === '/'
    return location.pathname.startsWith(path)
  }

  return (
    <div className="min-h-screen bg-[#f8fafc] flex flex-col font-sans">
      {/* Top bar */}
      <div className="bg-[#002244] text-white/80 text-xs py-1.5">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex items-center justify-between">
          <span>Gobierno de España — Ministerio para la Transición Ecológica y el Reto Demográfico</span>
          <span className="hidden sm:inline">Datos oficiales MITECO</span>
        </div>
      </div>

      {/* Main header */}
      <header className="hero-gradient text-white shadow-lg sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <Link to="/" className="flex items-center gap-3 group">
              <div className="p-2 bg-white/10 rounded-lg group-hover:bg-white/15 transition-colors">
                <LogoIcon />
              </div>
              <div>
                <div className="text-lg font-bold tracking-tight leading-tight">{t('appName')}</div>
                <div className="text-xs text-white/70 font-medium">Sistema de Información Hídrica</div>
              </div>
            </Link>

            <button
              className="md:hidden p-2 rounded-lg hover:bg-white/10 transition-colors"
              onClick={() => setMenuOpen(!menuOpen)}
              aria-label="Toggle menu"
            >
              {menuOpen ? <CloseIcon /> : <MenuIcon />}
            </button>

            <nav className="hidden md:flex items-center gap-1">
              {links.map((l) => (
                <Link
                  key={l.to}
                  to={l.to}
                  className={`px-3 py-2 rounded-md text-sm font-medium transition-all duration-200 ${
                    isActive(l.to)
                      ? 'bg-white/15 text-white shadow-sm'
                      : 'text-white/80 hover:text-white hover:bg-white/10'
                  }`}
                >
                  {l.label}
                </Link>
              ))}
            </nav>
          </div>
        </div>

        {/* Mobile menu */}
        {menuOpen && (
          <nav className="md:hidden border-t border-white/10 bg-[#002a52]">
            <div className="px-4 py-2 space-y-1">
              {links.map((l) => (
                <Link
                  key={l.to}
                  to={l.to}
                  className={`block px-3 py-2.5 rounded-md text-sm font-medium transition-colors ${
                    isActive(l.to)
                      ? 'bg-white/15 text-white'
                      : 'text-white/80 hover:text-white hover:bg-white/10'
                  }`}
                  onClick={() => setMenuOpen(false)}
                >
                  {l.label}
                </Link>
              ))}
            </div>
          </nav>
        )}
      </header>

      {/* Sub-navigation / breadcrumb area */}
      <div className="bg-white border-b border-[#e2e8f0] shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-3">
          <div className="flex items-center gap-2 text-sm text-[#475569]">
            <Link to="/" className="hover:text-[#003366] font-medium">{t('appName')}</Link>
            <span className="text-[#cbd5e1]">/</span>
            <span className="text-[#0f172a] font-semibold">
              {links.find(l => isActive(l.to))?.label || t('nav.home')}
            </span>
          </div>
        </div>
      </div>

      <main className="flex-1 max-w-7xl mx-auto w-full px-4 sm:px-6 lg:px-8 py-8">
        {children}
      </main>

      {/* Footer */}
      <footer className="bg-[#0f172a] text-slate-300">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10">
          <div className="grid md:grid-cols-3 gap-8">
            <div>
              <div className="flex items-center gap-2 mb-3">
                <div className="p-1.5 bg-white/10 rounded-md text-white">
                  <LogoIcon />
                </div>
                <span className="text-white font-bold text-lg">{t('appName')}</span>
              </div>
              <p className="text-sm text-slate-400 leading-relaxed">
                Sistema de información sobre el estado de los embalses en España.
                Datos proporcionados por el Ministerio para la Transición Ecológica
                y el Reto Demográfico (MITECO).
              </p>
            </div>
            <div>
              <h3 className="text-white font-semibold mb-3 text-sm uppercase tracking-wider">Navegación</h3>
              <ul className="space-y-2 text-sm">
                {links.slice(0, 4).map((l) => (
                  <li key={l.to}>
                    <Link to={l.to} className="hover:text-white transition-colors">{l.label}</Link>
                  </li>
                ))}
              </ul>
            </div>
            <div>
              <h3 className="text-white font-semibold mb-3 text-sm uppercase tracking-wider">Información</h3>
              <ul className="space-y-2 text-sm">
                <li>
                  <Link to="/fuentes" className="hover:text-white transition-colors">{t('nav.sources')}</Link>
                </li>
                <li>
                  <Link to="/calidad-datos" className="hover:text-white transition-colors">{t('nav.dataQuality')}</Link>
                </li>
                <li>
                  <Link to="/ajustes" className="hover:text-white transition-colors">{t('nav.settings')}</Link>
                </li>
                <li>
                  <Link to="/admin/importar" className="hover:text-white transition-colors">Importar datos</Link>
                </li>
              </ul>
            </div>
          </div>
          <div className="border-t border-slate-700 mt-8 pt-6 flex flex-col sm:flex-row items-center justify-between gap-4 text-xs text-slate-500">
            <p>© {new Date().getFullYear()} Embalses — Datos oficiales MITECO</p>
            <p className="flex items-center gap-1.5">
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              Los datos se actualizan semanalmente conforme a la información publicada por MITECO
            </p>
          </div>
        </div>
      </footer>
    </div>
  )
}
