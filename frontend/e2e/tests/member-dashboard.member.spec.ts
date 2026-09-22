import { test, expect } from '@playwright/test';
import { gotoApp, clickTab, expectNoError } from '../helpers/app';

// Runs with member storageState (chromium-member project).
// Verifies that a regular member sees the member UI, not the admin UI.

test.describe('Member Dashboard — initial view', () => {
  test.beforeEach(async ({ page }) => {
    await gotoApp(page);
  });

  test('stays authenticated — does not redirect to /auth/login', async ({ page }) => {
    await expect(page).not.toHaveURL(/\/auth\/login/);
  });

  test('shows member tabs, not admin tabs', async ({ page }) => {
    const nav = page.locator('nav.dt-nav, nav.mb-tabs');

    // Member tabs
    await expect(nav.getByRole('button', { name: 'Start',      exact: true })).toBeVisible();
    await expect(nav.getByRole('button', { name: 'Entdecken',  exact: true })).toBeVisible();
    await expect(nav.getByRole('button', { name: 'Schichten',  exact: true })).toBeVisible();
    await expect(nav.getByRole('button', { name: 'Profil',     exact: true })).toBeVisible();

    // Admin tabs must NOT be visible
    await expect(nav.getByRole('button', { name: 'Übersicht',    exact: true })).not.toBeVisible();
    await expect(nav.getByRole('button', { name: 'Mitglieder',   exact: true })).not.toBeVisible();
    await expect(nav.getByRole('button', { name: 'Einstellungen',exact: true })).not.toBeVisible();
  });

  test('Start tab renders without error boundary', async ({ page }) => {
    await expectNoError(page);
  });
});

test.describe('Member — Entdecken tab', () => {
  test.beforeEach(async ({ page }) => {
    await gotoApp(page);
    await clickTab(page, 'Entdecken');
  });

  test('Entdecken loads without crash', async ({ page }) => {
    await expectNoError(page);
  });
});

test.describe('Member — Schichten tab', () => {
  test.beforeEach(async ({ page }) => {
    await gotoApp(page);
    await clickTab(page, 'Schichten');
  });

  test('Schichten loads without crash', async ({ page }) => {
    await expectNoError(page);
  });

  test('shows upcoming or empty state section', async ({ page }) => {
    const hasSection = await page
      .getByText(/bevorstehend|keine schichten|meine schichten/i)
      .first()
      .isVisible({ timeout: 6_000 })
      .catch(() => false);
    expect(hasSection).toBeTruthy();
  });
});

test.describe('Member — Profil tab', () => {
  test.beforeEach(async ({ page }) => {
    await gotoApp(page);
    await clickTab(page, 'Profil');
  });

  test('Profil loads without crash', async ({ page }) => {
    await expectNoError(page);
  });

  test('Profil shows Abmelden option', async ({ page }) => {
    await expect(page.getByText('Abmelden')).toBeVisible({ timeout: 8_000 });
  });

  test('Abmelden from member profile redirects to login', async ({ page }) => {
    await page.getByText('Abmelden').waitFor({ timeout: 8_000 });
    await page.getByText('Abmelden').click();
    await expect(page).toHaveURL(/\/auth\/login/, { timeout: 5_000 });
  });
});
