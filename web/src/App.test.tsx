import { render, screen, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import './i18n'

// Mock fetch for API calls
const mockFetch = vi.fn()
Object.defineProperty(globalThis, 'fetch', { value: mockFetch })

import App from './App'

function Wrapper({ children }: { children: React.ReactNode }) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return (
    <QueryClientProvider client={client}>
      {children}
    </QueryClientProvider>
  )
}

describe('App', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    // Default mock responses
    mockFetch.mockResolvedValue({
      ok: true,
      json: async () => ({ data: [], meta: { page: 1, per_page: 20, total: 0, total_pages: 1 } }),
    })
  })

  it('renders app name in header', () => {
    render(<App />, { wrapper: Wrapper })
    // Use getAllByText since "Embalses" appears in header logo and footer
    const elements = screen.getAllByText('Embalses')
    expect(elements.length).toBeGreaterThanOrEqual(1)
  })

  it('renders navigation links in header nav', () => {
    render(<App />, { wrapper: Wrapper })
    // Query by role to specifically target nav links
    const nav = screen.getByRole('navigation', { hidden: true })
    expect(nav).toBeDefined()
    // Check specific nav links exist
    const links = screen.getAllByRole('link', { hidden: true })
    const linkTexts = links.map((l) => l.textContent)
    expect(linkTexts.some((t) => t?.includes('Home'))).toBe(true)
    expect(linkTexts.some((t) => t?.includes('Map'))).toBe(true)
    expect(linkTexts.some((t) => t?.includes('Reservoirs'))).toBe(true)
  })

  it('shows home page title', async () => {
    render(<App />, { wrapper: Wrapper })
    await waitFor(() => {
      expect(screen.getByText('Spanish Reservoir Status')).toBeDefined()
    })
  })

  it('renders settings page with language switch', async () => {
    render(<App />, { wrapper: Wrapper })
    const settingsLinks = screen.getAllByText('Settings')
    settingsLinks[0].click()
    await waitFor(() => {
      expect(screen.getByText('Language')).toBeDefined()
    })
  })
})
