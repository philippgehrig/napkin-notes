import { test, expect, type Page } from '@playwright/test'

async function registerFreshUser(page: Page): Promise<void> {
  const uniqueEmail = `user-${Date.now()}-${Math.random().toString(36).slice(2)}@test.com`

  await page.goto('/register')
  await page.fill('#displayName', 'E2E Test User')
  await page.fill('#email', uniqueEmail)
  await page.fill('#password', 'password123')
  await page.click('button[type="submit"]')
  await expect(page).toHaveURL('/')
}

async function waitForAutoSave(page: Page): Promise<void> {
  await expect(page).toHaveURL(/\/napkin\//, { timeout: 5000 })
}

test.describe('Notes CRUD', () => {
  test.beforeEach(async ({ page }) => {
    await registerFreshUser(page)
  })

  test('creates a new napkin note', async ({ page }) => {
    const textarea = page.locator('.napkin-page__input')
    await textarea.fill('My first napkin note')
    await waitForAutoSave(page)

    await page.goto('/gallery')
    await expect(page.locator('.napkin-card__content')).toContainText('My first napkin note')
  })

  test('edits an existing note from gallery', async ({ page }) => {
    const textarea = page.locator('.napkin-page__input')
    await textarea.fill('Original content')
    await waitForAutoSave(page)

    await page.goto('/gallery')
    await page.click('.napkin-card')
    await expect(page).toHaveURL(/\/napkin\//)

    const napkinInput = page.locator('.napkin-page__input')
    await napkinInput.clear()
    await napkinInput.fill('Updated content')
    await page.waitForTimeout(1500)

    await page.goto('/gallery')
    await expect(page.locator('.napkin-card__content')).toContainText('Updated content')
  })

  test('empty state shows napkin page with placeholder', async ({ page }) => {
    await expect(page.locator('.napkin-page__input')).toBeVisible()
    await expect(page.locator('.napkin-page__input')).toHaveAttribute('placeholder', 'Write on your napkin...')
  })
})
