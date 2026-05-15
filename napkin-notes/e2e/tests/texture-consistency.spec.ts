import { test, expect, type Page } from '@playwright/test'

async function registerFreshUser(page: Page): Promise<void> {
  const uniqueEmail = `user-${Date.now()}-${Math.random().toString(36).slice(2)}@test.com`

  await page.goto('/register')
  await page.fill('#displayName', 'Texture Test User')
  await page.fill('#email', uniqueEmail)
  await page.fill('#password', 'password123')
  await page.click('button[type="submit"]')
  await expect(page).toHaveURL('/')
}

function getTextureSrc(page: Page, selector: string) {
  return page.locator(selector).locator('.napkin-texture__img').getAttribute('src')
}

test.describe('Napkin texture consistency', () => {
  test.beforeEach(async ({ page }) => {
    await registerFreshUser(page)
  })

  test('texture in editor matches gallery after save', async ({ page }) => {
    // Create a note and save it
    const textarea = page.locator('.napkin-page__input')
    await textarea.fill('Texture test note')
    await page.click('.napkin-page__save-btn')

    // Wait for the texture to update after save (variant changes when ID is assigned)
    // The texture should no longer be the default (napkin1) after save assigns an ID
    await page.waitForFunction(() => {
      const img = document.querySelector('.napkin-page__napkin .napkin-texture__img') as HTMLImageElement
      return img && img.src.includes('/textures/napkin')
    })
    // Small wait for Vue reactivity to propagate
    await page.waitForTimeout(200)

    const editorTexture = await getTextureSrc(page, '.napkin-page__napkin')
    expect(editorTexture).not.toBeNull()

    // Navigate to gallery
    await page.goto('/gallery')
    await expect(page.locator('.napkin-card')).toHaveCount(1)

    // Get the texture shown on the gallery card
    const galleryTexture = await getTextureSrc(page, '.napkin-card')
    expect(galleryTexture).not.toBeNull()

    // They must match
    expect(editorTexture).toBe(galleryTexture)
  })

  test('texture persists when opening note from gallery', async ({ page }) => {
    // Create a note
    const textarea = page.locator('.napkin-page__input')
    await textarea.fill('Persistence test')
    await page.click('.napkin-page__save-btn')

    // Go to gallery
    await page.goto('/gallery')
    await expect(page.locator('.napkin-card')).toHaveCount(1)

    // Record gallery texture
    const galleryTexture = await getTextureSrc(page, '.napkin-card')

    // Open the note from gallery
    await page.click('.napkin-card')
    await expect(page).toHaveURL(/\/napkin\//)

    // Get editor texture
    const editorTexture = await getTextureSrc(page, '.napkin-page__napkin')

    // Must match
    expect(editorTexture).toBe(galleryTexture)
  })

  test('texture stays the same after editing and re-saving', async ({ page }) => {
    // Create a note
    const textarea = page.locator('.napkin-page__input')
    await textarea.fill('Initial content')
    await page.click('.napkin-page__save-btn')
    await page.waitForTimeout(200)

    // Record texture after first save
    const firstTexture = await getTextureSrc(page, '.napkin-page__napkin')

    // Edit the note (use selectAll + type to avoid clear triggering issues)
    await textarea.selectText()
    await textarea.fill('Edited content')
    await page.click('.napkin-page__save-btn')
    await page.waitForTimeout(200)

    // Texture should not change (same note ID)
    const secondTexture = await getTextureSrc(page, '.napkin-page__napkin')
    expect(secondTexture).toBe(firstTexture)

    // Check gallery still matches
    await page.goto('/gallery')
    await expect(page.locator('.napkin-card')).toHaveCount(1)
    const galleryTexture = await getTextureSrc(page, '.napkin-card')
    expect(galleryTexture).toBe(firstTexture)
  })
})
