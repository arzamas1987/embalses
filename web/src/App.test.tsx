import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import App from './App'

describe('App', () => {
  it('renders Embalses MVP heading', () => {
    render(<App />)
    expect(screen.getByText('Embalses MVP')).toBeDefined()
  })

  it('shows backend status section', () => {
    render(<App />)
    expect(screen.getByText('Backend Status')).toBeDefined()
  })
})
