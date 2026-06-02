import { test, expect } from '@playwright/test';

// These tests run without any auth state (chromium-public project).

test.describe('Unauthenticated access', () => {
  test('redirects / to /auth/login when not logged in', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveURL(/\/auth\/login/);
  });

  test('login page shows OIDC button', async ({ page }) => {
    await page.goto('/auth/login');
    await expect(page.getByRole('button', { name: /login via oidc/i })).toBeVisible();
  });

  test('login page shows app title', async ({ page }) => {
    await page.goto('/auth/login');
    await expect(page.getByText('ShiftManager')).toBeVisible();
  });

  test('kiosk page is accessible without auth', async ({ page }) => {
    await page.goto('/kiosk');
    // Should NOT redirect to login
    await expect(page).not.toHaveURL(/\/auth\/login/);
  });
});
