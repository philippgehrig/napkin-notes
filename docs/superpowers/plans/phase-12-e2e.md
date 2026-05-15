# Phase 12: E2E Tests

> **For agentic workers:** Use superpowers:subagent-driven-development or superpowers:executing-plans.

**Goal:** Add Playwright end-to-end tests covering registration, notes CRUD, and rip-to-delete.

**Branch:** `feat/phase-12-e2e`

---

## File Structure

```
e2e/
├── package.json
├── playwright.config.ts
├── tests/
│   ├── auth.spec.ts
│   ├── notes.spec.ts
│   └── gallery.spec.ts
└── tsconfig.json
```

---

### Task 1: Playwright setup

**Files:**
- Create: `e2e/package.json`
- Create: `e2e/playwright.config.ts`
- Create: `e2e/tsconfig.json`

- [ ] **Step 1: Create package.json**

Create `e2e/package.json`:
```json
{
  "name": "napkin-notes-e2e",
  "private": true,
  "scripts": {
    "test": "playwright test",
    "test:headed": "playwright test --headed",
    "test:ui": "playwright test --ui"
  },
  "devDependencies": {
    "@playwright/test": "^1.42.0",
    "typescript": "^5.3.0"
  }
}
```

- [ ] **Step 2: Create playwright.config.ts**

Create `e2e/playwright.config.ts`:
```ts
import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: './tests',
  timeout: 30000,
  retries: 1,
  use: {
    baseURL: process.env.BASE_URL || 'http://localhost',
    screenshot: 'only-on-failure',
    trace: 'retain-on-failure',
  },
  projects: [
    { name: 'chromium', use: { browserName: 'chromium' } },
  ],
})
```

- [ ] **Step 3: Create tsconfig.json**

Create `e2e/tsconfig.json`:
```json
{
  "compilerOptions": {
    "target": "ES2020",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "strict": true,
    "esModuleInterop": true
  }
}
```

- [ ] **Step 4: Install dependencies**

```bash
cd e2e && yarn install && npx playwright install chromium
```

- [ ] **Step 5: Commit**

```bash
git add e2e/
git commit -m "feat: add Playwright E2E test setup"
```

---

### Task 2: Auth E2E tests

**Files:**
- Create: `e2e/tests/auth.spec.ts`

- [ ] **Step 1: Write auth tests**

Create `e2e/tests/auth.spec.ts`:
```ts
import { test, expect } from '@playwright/test'

const testUser = {
  email: `test-${Date.now()}@example.com`,
  password: 'testpassword123',
  displayName: 'Test User',
}

test.describe('Authentication', () => {
  test('shows login page for unauthenticated users', async ({ page }) => {
    await page.goto('/')
    await expect(page).toHaveURL(/\/login/)
    await expect(page.getByRole('heading', { name: /napkin notes/i })).toBeVisible()
  })

  test('registers a new user', async ({ page }) => {
    await page.goto('/register')
    await page.getByPlaceholder('Display name').fill(testUser.displayName)
    await page.getByPlaceholder('Email').fill(testUser.email)
    await page.getByPlaceholder(/password/i).fill(testUser.password)
    await page.getByRole('button', { name: /create account/i }).click()

    await expect(page).toHaveURL('/')
    await expect(page.getByText(/your napkins/i)).toBeVisible()
  })

  test('logs in with existing user', async ({ page }) => {
    // Register first
    await page.goto('/register')
    const email = `login-${Date.now()}@example.com`
    await page.getByPlaceholder('Display name').fill('Login Test')
    await page.getByPlaceholder('Email').fill(email)
    await page.getByPlaceholder(/password/i).fill('password123')
    await page.getByRole('button', { name: /create account/i }).click()
    await expect(page).toHaveURL('/')

    // Logout
    await page.getByRole('button', { name: /logout/i }).click()
    await expect(page).toHaveURL(/\/login/)

    // Login
    await page.getByPlaceholder('Email').fill(email)
    await page.getByPlaceholder('Password').fill('password123')
    await page.getByRole('button', { name: /sign in/i }).click()
    await expect(page).toHaveURL('/')
  })

  test('shows error for wrong credentials', async ({ page }) => {
    await page.goto('/login')
    await page.getByPlaceholder('Email').fill('wrong@example.com')
    await page.getByPlaceholder('Password').fill('wrongpassword')
    await page.getByRole('button', { name: /sign in/i }).click()

    await expect(page.getByText(/invalid/i)).toBeVisible()
  })
})
```

- [ ] **Step 2: Commit**

```bash
git add e2e/tests/auth.spec.ts
git commit -m "feat: add auth E2E tests"
```

---

### Task 3: Notes CRUD E2E tests

**Files:**
- Create: `e2e/tests/notes.spec.ts`

- [ ] **Step 1: Write notes tests**

Create `e2e/tests/notes.spec.ts`:
```ts
import { test, expect } from '@playwright/test'

test.describe('Notes CRUD', () => {
  test.beforeEach(async ({ page }) => {
    // Register and login
    const email = `notes-${Date.now()}@example.com`
    await page.goto('/register')
    await page.getByPlaceholder('Display name').fill('Notes Tester')
    await page.getByPlaceholder('Email').fill(email)
    await page.getByPlaceholder(/password/i).fill('password123')
    await page.getByRole('button', { name: /create account/i }).click()
    await expect(page).toHaveURL('/')
  })

  test('creates a new napkin note', async ({ page }) => {
    await page.getByRole('button', { name: /new napkin/i }).click()
    await expect(page).toHaveURL(/\/note/)

    await page.getByPlaceholder(/write your note/i).fill('My first napkin note!')
    await page.getByRole('button', { name: /save/i }).click()

    // Go back to gallery
    await page.getByRole('button', { name: /back/i }).click()
    await expect(page.getByText('My first napkin note!')).toBeVisible()
  })

  test('edits an existing note', async ({ page }) => {
    // Create note
    await page.getByRole('button', { name: /new napkin/i }).click()
    await page.getByPlaceholder(/write your note/i).fill('Original content')
    await page.getByRole('button', { name: /save/i }).click()
    await page.getByRole('button', { name: /back/i }).click()

    // Open and edit
    await page.getByText('Original content').click()
    const textarea = page.getByPlaceholder(/write your note/i)
    await textarea.clear()
    await textarea.fill('Updated content')
    await page.getByRole('button', { name: /save/i }).click()
    await page.getByRole('button', { name: /back/i }).click()

    await expect(page.getByText('Updated content')).toBeVisible()
    await expect(page.getByText('Original content')).not.toBeVisible()
  })

  test('shows empty state when no notes exist', async ({ page }) => {
    await expect(page.getByText(/no napkins yet/i)).toBeVisible()
  })
})
```

- [ ] **Step 2: Commit**

```bash
git add e2e/tests/notes.spec.ts
git commit -m "feat: add notes CRUD E2E tests"
```

---

### Task 4: Gallery and rip-to-delete E2E tests

**Files:**
- Create: `e2e/tests/gallery.spec.ts`

- [ ] **Step 1: Write gallery tests**

Create `e2e/tests/gallery.spec.ts`:
```ts
import { test, expect } from '@playwright/test'

test.describe('Gallery', () => {
  test.beforeEach(async ({ page }) => {
    const email = `gallery-${Date.now()}@example.com`
    await page.goto('/register')
    await page.getByPlaceholder('Display name').fill('Gallery Tester')
    await page.getByPlaceholder('Email').fill(email)
    await page.getByPlaceholder(/password/i).fill('password123')
    await page.getByRole('button', { name: /create account/i }).click()
    await expect(page).toHaveURL('/')

    // Create a few notes
    for (const text of ['Note A', 'Note B', 'Note C']) {
      await page.getByRole('button', { name: /new napkin/i }).click()
      await page.getByPlaceholder(/write your note/i).fill(text)
      await page.getByRole('button', { name: /save/i }).click()
      await page.getByRole('button', { name: /back/i }).click()
    }
  })

  test('displays multiple napkin cards', async ({ page }) => {
    await expect(page.getByText('Note A')).toBeVisible()
    await expect(page.getByText('Note B')).toBeVisible()
    await expect(page.getByText('Note C')).toBeVisible()
  })

  test('navigates to trash view', async ({ page }) => {
    await page.getByRole('link', { name: /trash/i }).click()
    await expect(page).toHaveURL('/trash')
    await expect(page.getByText(/no ripped napkins/i)).toBeVisible()
  })

  test('rip-to-delete moves note to trash', async ({ page }) => {
    const card = page.getByText('Note B').first()
    const box = await card.boundingBox()
    if (!box) throw new Error('Card not visible')

    // Simulate drag gesture (pointer down, move right, release)
    await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2)
    await page.mouse.down()
    await page.mouse.move(box.x + box.width * 1.5, box.y + box.height / 2, { steps: 10 })
    await page.mouse.up()

    // Wait for deletion
    await page.waitForTimeout(500)

    // Note B should be gone from gallery
    await expect(page.getByText('Note B')).not.toBeVisible()

    // Check trash
    await page.getByRole('link', { name: /trash/i }).click()
    await expect(page.getByText('Note B')).toBeVisible()
  })

  test('restores a trashed note', async ({ page }) => {
    // Delete Note A via drag
    const card = page.getByText('Note A').first()
    const box = await card.boundingBox()
    if (!box) throw new Error('Card not visible')

    await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2)
    await page.mouse.down()
    await page.mouse.move(box.x + box.width * 1.5, box.y + box.height / 2, { steps: 10 })
    await page.mouse.up()
    await page.waitForTimeout(500)

    // Go to trash and restore
    await page.getByRole('link', { name: /trash/i }).click()
    await page.getByRole('button', { name: /restore/i }).click()

    // Back to gallery
    await page.getByRole('link', { name: /gallery/i }).click()
    await expect(page.getByText('Note A')).toBeVisible()
  })
})
```

- [ ] **Step 2: Commit**

```bash
git add e2e/tests/gallery.spec.ts
git commit -m "feat: add gallery and rip-to-delete E2E tests"
```

---

### Task 5: Add E2E to CI and Makefile

**Files:**
- Modify: `.github/workflows/ci.yml`
- Modify: `Makefile`

- [ ] **Step 1: Add E2E job to CI**

Add to `.github/workflows/ci.yml`:
```yaml
  e2e:
    runs-on: ubuntu-latest
    needs: [test-api, test-web]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - name: Start services
        run: |
          docker compose -f docker/docker-compose.yml -f docker/docker-compose.dev.yml up -d --build
          sleep 10
      - name: Install E2E deps
        run: cd e2e && yarn install && npx playwright install chromium --with-deps
      - name: Run E2E tests
        run: cd e2e && yarn test
        env:
          BASE_URL: http://localhost
      - name: Stop services
        if: always()
        run: docker compose -f docker/docker-compose.yml -f docker/docker-compose.dev.yml down
```

- [ ] **Step 2: Add to Makefile**

Add to `Makefile`:
```makefile
test-e2e:
	cd e2e && yarn test

test-all: test test-e2e
```

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/ci.yml Makefile
git commit -m "ci: add E2E tests to CI pipeline"
```

---

### Task 6: Create PR

- [ ] **Step 1: Push and create PR**

```bash
git push -u origin feat/phase-12-e2e
gh pr create --title "feat: Playwright E2E test suite" --body "## Summary
- Add Playwright E2E test infrastructure
- Add auth flow tests (register, login, error handling)
- Add notes CRUD tests (create, edit, empty state)
- Add gallery tests (display, rip-to-delete, restore)
- Add E2E tests to CI pipeline

## Test plan
- [ ] \`cd e2e && yarn test\` passes against running app
- [ ] CI E2E job runs after unit tests pass
- [ ] Rip gesture E2E test works via mouse simulation

🤖 Generated with [Claude Code](https://claude.com/claude-code)"
```
