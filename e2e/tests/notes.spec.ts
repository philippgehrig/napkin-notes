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

test.describe('Notes CRUD', () => {
  test.beforeEach(async ({ page }) => {
    await registerFreshUser(page)
  })

  test('creates a new napkin note', async ({ page }) => {
    const textarea = page.locator('.napkin-page__input')
    await textarea.fill('My first napkin note')
    await page.click('.napkin-page__new-btn')

    // Navigate to gallery to verify note was saved
    await page.goto('/gallery')
    await expect(page.locator('.napkin-card__content')).toContainText('My first napkin note')
  })

  test('edits an existing note from gallery', async ({ page }) => {
    // Create a note first
    const textarea = page.locator('.napkin-page__input')
    await textarea.fill('Original content')
    await page.click('.napkin-page__new-btn')

    // Go to gallery and open the note
    await page.goto('/gallery')
    await page.click('.napkin-card')
    await expect(page).toHaveURL(/\/napkin\//)

    // Edit the note
    const napkinInput = page.locator('.napkin-page__input')
    await napkinInput.clear()
    await napkinInput.fill('Updated content')
    await page.click('.napkin-page__new-btn')

    // Verify in gallery
    await page.goto('/gallery')
    await expect(page.locator('.napkin-card__content')).toContainText('Updated content')
  })

  test('empty state shows napkin page with placeholder', async ({ page }) => {
    await expect(page.locator('.napkin-page__input')).toBeVisible()
    await expect(page.locator('.napkin-page__input')).toHaveAttribute('placeholder', 'Write on your napkin...')
  })
})
