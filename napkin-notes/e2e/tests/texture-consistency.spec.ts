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

async function waitForAutoSave(page: Page): Promise<void> {
  await expect(page).toHaveURL(/\/napkin\//, { timeout: 5000 })
}

test.describe('Napkin texture consistency', () => {
  test.beforeEach(async ({ page }) => {
    await registerFreshUser(page)
  })

  test('texture in editor matches gallery after save', async ({ page }) => {
    const textarea = page.locator('.napkin-page__input')
    await textarea.fill('Texture test note')
    await waitForAutoSave(page)

    const editorTexture = await getTextureSrc(page, '.napkin-page__napkin')
    expect(editorTexture).not.toBeNull()

    await page.goto('/gallery')
    await expect(page.locator('.napkin-card')).toHaveCount(1)

    const galleryTexture = await getTextureSrc(page, '.napkin-card')
    expect(galleryTexture).not.toBeNull()

    expect(editorTexture).toBe(galleryTexture)
  })

  test('texture persists when opening note from gallery', async ({ page }) => {
    const textarea = page.locator('.napkin-page__input')
    await textarea.fill('Persistence test')
    await waitForAutoSave(page)

    await page.goto('/gallery')
    await expect(page.locator('.napkin-card')).toHaveCount(1)

    const galleryTexture = await getTextureSrc(page, '.napkin-card')

    await page.click('.napkin-card')
    await expect(page).toHaveURL(/\/napkin\//)

    const editorTexture = await getTextureSrc(page, '.napkin-page__napkin')

    expect(editorTexture).toBe(galleryTexture)
  })

  test('texture stays the same after editing and re-saving', async ({ page }) => {
    const textarea = page.locator('.napkin-page__input')
    await textarea.fill('Initial content')
    await waitForAutoSave(page)

    const firstTexture = await getTextureSrc(page, '.napkin-page__napkin')

    await textarea.selectText()
    await textarea.fill('Edited content')
    await page.waitForTimeout(1500)

    const secondTexture = await getTextureSrc(page, '.napkin-page__napkin')
    expect(secondTexture).toBe(firstTexture)

    await page.goto('/gallery')
    await expect(page.locator('.napkin-card')).toHaveCount(1)
    const galleryTexture = await getTextureSrc(page, '.napkin-card')
    expect(galleryTexture).toBe(firstTexture)
  })
})
