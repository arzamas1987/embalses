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
    expect(screen.getByText('Embalses')).toBeDefined()
  })

  it('renders navigation links', () => {
    render(<App />, { wrapper: Wrapper })
    expect(screen.getByText('Home')).toBeDefined()
    expect(screen.getByText('Map')).toBeDefined()
    expect(screen.getByText('Reservoirs')).toBeDefined()
    expect(screen.getByText('Compare')).toBeDefined()
    expect(screen.getByText('Settings')).toBeDefined()
  })

  it('shows home page title', async () => {
    render(<App />, { wrapper: Wrapper })
    await waitFor(() => {
      expect(screen.getByText('Spanish Reservoir Status')).toBeDefined()
    })
  })

  it('renders settings page with language switch', async () => {
    render(<App />, { wrapper: Wrapper })
    const settingsLink = screen.getByText('Settings')
    settingsLink.click()
    await waitFor(() => {
      expect(screen.getByText('Language')).toBeDefined()
    })
  })
})
