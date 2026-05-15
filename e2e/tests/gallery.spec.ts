import { test, expect, type Page } from '@playwright/test'

async function registerAndCreateNotes(page: Page, count: number): Promise<void> {
  const uniqueEmail = `user-${Date.now()}-${Math.random().toString(36).slice(2)}@test.com`

  await page.goto('/register')
  await page.fill('#displayName', 'Gallery Test User')
  await page.fill('#email', uniqueEmail)
  await page.fill('#password', 'password123')
  await page.click('button[type="submit"]')
  await expect(page).toHaveURL('/')

  for (let i = 1; i <= count; i++) {
    const textarea = page.locator('.napkin-page__input')
    await textarea.fill(`Napkin note number ${i}`)
    await page.click('.napkin-page__new-btn')
  }
}

test.describe('Gallery', () => {
  test.beforeEach(async ({ page }) => {
    await registerAndCreateNotes(page, 3)
  })

  test('displays multiple napkin cards', async ({ page }) => {
    await page.goto('/gallery')
    const cards = page.locator('.napkin-card')
    await expect(cards).toHaveCount(3)
  })

  test('navigates to trash view', async ({ page }) => {
    await page.goto('/trash')
    await expect(page.locator('.trash__title')).toHaveText('Ripped Napkins')
  })

  test('rip-to-delete moves note to trash', async ({ page }) => {
    await page.goto('/gallery')
    const firstCard = page.locator('.napkin-card').first()
    const box = await firstCard.boundingBox()
    expect(box).not.toBeNull()

    // Simulate horizontal drag (pointer events) exceeding the 40% threshold
    const startX = box!.x + box!.width * 0.2
    const startY = box!.y + box!.height / 2
    const endX = box!.x + box!.width * 0.8 // drag 60% of width

    await page.mouse.move(startX, startY)
    await page.mouse.down()
    await page.mouse.move(endX, startY, { steps: 10 })
    await page.mouse.up()

    // Should now have 2 cards
    await expect(page.locator('.napkin-card')).toHaveCount(2)
  })

  test('restore from trash moves note back', async ({ page }) => {
    await page.goto('/gallery')
    const firstCard = page.locator('.napkin-card').first()
    const box = await firstCard.boundingBox()
    expect(box).not.toBeNull()

    // Rip the first card
    const startX = box!.x + box!.width * 0.2
    const startY = box!.y + box!.height / 2
    const endX = box!.x + box!.width * 0.8

    await page.mouse.move(startX, startY)
    await page.mouse.down()
    await page.mouse.move(endX, startY, { steps: 10 })
    await page.mouse.up()

    await expect(page.locator('.napkin-card')).toHaveCount(2)

    // Navigate to trash
    await page.goto('/trash')
    await expect(page.locator('.trash__card')).toHaveCount(1)

    // Click restore button
    await page.click('.trash__restore-btn')
    await expect(page.locator('.trash__card')).toHaveCount(0)

    // Go back to gallery and verify restored
    await page.goto('/gallery')
    await expect(page.locator('.napkin-card')).toHaveCount(3)
  })
})
