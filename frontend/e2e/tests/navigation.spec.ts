import { test, expect } from '@playwright/test';
import { gotoApp, expectNoError } from '../helpers/app';

// Runs with admin storageState (chromium project).

test.describe('Admin navigation', () => {
  test.beforeEach(async ({ page }) => {
    await gotoApp(page);
  });

  test('stays on / — does not redirect to login when authenticated', async ({ page }) => {
    await expect(page).not.toHaveURL(/\/auth\/login/);
  });

  test('all four admin tabs are visible in the sidebar', async ({ page }) => {
    const nav = page.locator('nav.dt-nav');
    await expect(nav.getByRole('button', { name: 'Übersicht', exact: true })).toBeVisible();
    await expect(nav.getByRole('button', { name: 'Termine',   exact: true })).toBeVisible();
    await expect(nav.getByRole('button', { name: 'Mitglieder', exact: true })).toBeVisible();
    await expect(nav.getByRole('button', { name: 'Einstellungen', exact: true })).toBeVisible();
  });

  test('no crash on initial load', async ({ page }) => {
    await expectNoError(page);
  });
});
