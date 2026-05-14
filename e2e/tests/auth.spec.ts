import { test, expect } from '@playwright/test'

test.describe('Authentication', () => {
  test('unauthenticated users see login page', async ({ page }) => {
    await page.goto('/')
    await expect(page).toHaveURL(/\/login/)
    await expect(page.locator('h1')).toHaveText('Login')
  })

  test('register a new user redirects to gallery', async ({ page }) => {
    const uniqueEmail = `user-${Date.now()}@test.com`

    await page.goto('/register')
    await page.fill('#displayName', 'Test User')
    await page.fill('#email', uniqueEmail)
    await page.fill('#password', 'password123')
    await page.click('button[type="submit"]')

    await expect(page).toHaveURL('/')
    await expect(page.locator('.gallery__title')).toHaveText('Your Napkins')
  })

  test('login with registered user arrives at gallery', async ({ page }) => {
    const uniqueEmail = `user-${Date.now()}@test.com`

    // First register
    await page.goto('/register')
    await page.fill('#displayName', 'Login Test User')
    await page.fill('#email', uniqueEmail)
    await page.fill('#password', 'password123')
    await page.click('button[type="submit"]')
    await expect(page).toHaveURL('/')

    // Clear local storage and go to login
    await page.evaluate(() => localStorage.clear())
    await page.goto('/login')

    // Login
    await page.fill('#email', uniqueEmail)
    await page.fill('#password', 'password123')
    await page.click('button[type="submit"]')

    await expect(page).toHaveURL('/')
    await expect(page.locator('.gallery__title')).toHaveText('Your Napkins')
  })

  test('wrong credentials shows error message', async ({ page }) => {
    await page.goto('/login')
    await page.fill('#email', 'nonexistent@test.com')
    await page.fill('#password', 'wrongpassword')
    await page.click('button[type="submit"]')

    await expect(page.locator('.error')).toBeVisible()
  })
})
