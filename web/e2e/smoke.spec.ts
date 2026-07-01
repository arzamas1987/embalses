import { test, expect } from '@playwright/test'

const API_URL = process.env.API_URL || 'http://localhost:8082'
const API_KEY = process.env.API_KEY || 'test-key-123'

async function apiGet(path: string) {
  const res = await fetch(`${API_URL}${path}`, {
    headers: { 'X-API-Key': API_KEY },
  })
  if (!res.ok) {
    throw new Error(`API ${path} returned ${res.status}`)
  }
  return res.json()
}

test.describe('MVP SQLite smoke tests', () => {
  test('API endpoints return reservoirs, basins and historical readings', async () => {
    const reservoirsResp = (await apiGet('/api/v1/reservoirs')) as { data: Array<{ name: string; slug?: string }> }
    expect(reservoirsResp.data.length).toBeGreaterThan(0)

    const basinsResp = (await apiGet('/api/v1/basins')) as { data: Array<{ name: string }> }
    expect(basinsResp.data.length).toBeGreaterThan(0)

    const first = reservoirsResp.data[0]
    const slug = first.slug ?? encodeURIComponent(first.name)
    const readingsResp = (await apiGet(`/api/v1/reservoirs/${slug}/readings`)) as {
      data: Array<{ observed_at: string; fill_pct: number; volume_hm3: number }>
    }
    expect(readingsResp.data.length).toBeGreaterThan(0)

    const firstReading = readingsResp.data[0]
    expect(firstReading.fill_pct).toBeGreaterThanOrEqual(0)
    expect(firstReading.volume_hm3).toBeGreaterThanOrEqual(0)
  })

  test('reservoirs page shows a list of reservoirs', async ({ page }) => {
    await page.goto('/embalses')
    await expect(page.getByTestId('reservoirs-table')).toBeVisible()
    const rows = page.getByTestId('reservoir-row')
    await expect(rows.first()).toBeVisible()
    const count = await rows.count()
    expect(count).toBeGreaterThan(0)
  })

  test('basins page shows a list of basins', async ({ page }) => {
    await page.goto('/cuencas')
    await expect(page.getByTestId('basins-grid')).toBeVisible()
    const cards = page.getByTestId('basin-card')
    await expect(cards.first()).toBeVisible()
    const count = await cards.count()
    expect(count).toBeGreaterThan(0)
  })

  test('reservoir detail shows name and historical readings', async ({ page }) => {
    const reservoirsResp = (await apiGet('/api/v1/reservoirs')) as { data: Array<{ name: string; slug?: string }> }
    expect(reservoirsResp.data.length).toBeGreaterThan(0)

    const first = reservoirsResp.data[0]
    const slug = first.slug ?? encodeURIComponent(first.name)
    await page.goto(`/embalses/${encodeURIComponent(slug)}`)

    await expect(page.getByTestId('reservoir-detail-name')).toHaveText(first.name)
    await expect(page.getByTestId('readings-table')).toBeVisible()
  })
})
