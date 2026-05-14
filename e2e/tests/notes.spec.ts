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
    await page.click('.gallery__new-btn')
    await expect(page).toHaveURL(/\/note/)

    const textarea = page.locator('.editor__textarea')
    await textarea.fill('My first napkin note')
    await page.click('.editor__save-btn')

    await expect(page).toHaveURL('/')
    await expect(page.locator('.napkin-card__content')).toContainText('My first napkin note')
  })

  test('edits an existing note', async ({ page }) => {
    // Create a note first
    await page.click('.gallery__new-btn')
    const textarea = page.locator('.editor__textarea')
    await textarea.fill('Original content')
    await page.click('.editor__save-btn')
    await expect(page).toHaveURL('/')

    // Open the note for editing
    await page.click('.napkin-card__content')
    await expect(page).toHaveURL(/\/note\//)

    // Clear and type new content
    const editorTextarea = page.locator('.editor__textarea')
    await editorTextarea.clear()
    await editorTextarea.fill('Updated content')
    await page.click('.editor__save-btn')

    await expect(page).toHaveURL('/')
    await expect(page.locator('.napkin-card__content')).toContainText('Updated content')
  })

  test('empty state shows no napkins message', async ({ page }) => {
    await expect(page.locator('.gallery__empty')).toBeVisible()
    await expect(page.locator('.gallery__empty')).toContainText('No napkins yet')
  })
})
