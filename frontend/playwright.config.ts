import { defineConfig, devices } from '@playwright/test';
import { resolve } from 'path';
import { fileURLToPath } from 'url';

const __dirname = fileURLToPath(new URL('.', import.meta.url));

// ── E2E-specific config ──────────────────────────────────────────────────────
// The E2E backend runs on a DEDICATED port so it never clashes with the
// developer's debug session (which lives on :8080).
// TEST_MODE=true → SQLite in-memory + pre-seeded fixture data + /dev/token active.
export const E2E_BACKEND_PORT = process.env.E2E_BACKEND_PORT ?? '8082';
const BACKEND_URL = `http://localhost:${E2E_BACKEND_PORT}`;

// The Vite dev server needs a separate port so it doesn't collide with the
// developer's own Vite instance that proxies to the debug backend.
const E2E_VITE_PORT = process.env.E2E_VITE_PORT ?? '5174';
const BASE_URL      = `http://localhost:${E2E_VITE_PORT}`;

export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,   // tests share backend state — run sequentially
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  reporter: [['html', { open: 'never' }], ['list']],

  use: {
    baseURL: BASE_URL,
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },

  projects: [
    // Generates storageState files (auth tokens) before any test runs
    {
      name: 'setup',
      testMatch: /global\.setup\.ts/,
    },

    // Authenticated tests — admin session
    {
      name: 'chromium',
      use: {
        ...devices['Desktop Chrome'],
        storageState: 'e2e/.auth/admin.json',
      },
      dependencies: ['setup'],
      testIgnore: [/.*\.public\.spec\.ts/, /.*\.member\.spec\.ts/],
    },

    // Authenticated tests — regular member session
    {
      name: 'chromium-member',
      use: {
        ...devices['Desktop Chrome'],
        storageState: 'e2e/.auth/member.json',
      },
      dependencies: ['setup'],
      testMatch: /.*\.member\.spec\.ts/,
    },

    // Unauthenticated tests — no storageState
    {
      name: 'chromium-public',
      use: devices['Desktop Chrome'],
      testMatch: /.*\.public\.spec\.ts/,
    },
  ],

  webServer: [
    // ── Go backend (TEST_MODE) ───────────────────────────────────────────────
    // SQLite in-memory, fixture data seeded, /dev/token active, no rate limiting.
    // Uses a dedicated port so it never clashes with the developer's debug backend.
    {
      command: `TEST_MODE=true PORT=${E2E_BACKEND_PORT} go run ./cmd/server`,
      cwd: resolve(__dirname, '..'),
      url: `${BACKEND_URL}/health`,
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
      stdout: 'pipe',
      stderr: 'pipe',
    },

    // ── Vite dev server ──────────────────────────────────────────────────────
    // VITE_API_URL points directly at the E2E backend (bypasses the :8080 proxy).
    // Runs on a separate port (5174) to avoid colliding with the developer's Vite.
    {
      command: `VITE_API_URL=${BACKEND_URL}/api/v1 npm run dev -- --port ${E2E_VITE_PORT} --strictPort`,
      url: BASE_URL,
      reuseExistingServer: !process.env.CI,
      timeout: 30_000,
      stdout: 'pipe',
      stderr: 'pipe',
    },
  ],
});
