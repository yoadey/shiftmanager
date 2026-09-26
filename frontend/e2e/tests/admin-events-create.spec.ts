import { test, expect } from '@playwright/test';
import { gotoApp, expectNoError } from '../helpers/app';

// Runs with admin storageState (chromium project).
// Regression coverage for the bedienkonzept-formulare-als-seite OpenSpec
// change (UX-001): event creation used to be a Sheet overlay on top of
// AdminEvents; it is now a real page with no overlay/backdrop.

test.describe('Admin — Veranstaltung erstellen (own page, not a popup)', () => {
  test.beforeEach(async ({ page }) => {
    await gotoApp(page);
    await page.goto('/events/neu');
  });

  test('renders as a page, not inside a Sheet overlay', async ({ page }) => {
    await expectNoError(page);
    await expect(page.getByText('Eckdaten')).toBeVisible();
    await expect(page.locator('.sm-overlay')).toHaveCount(0);
    await expect(page.locator('.sm-sheet')).toHaveCount(0);
    // The sidebar nav stays visible next to the form — unlike a Sheet, which
    // would cover or dim it.
    await expect(page.locator('nav.dt-nav, nav.mb-tabs')).toBeVisible();
  });

  test('can fill the first step and advance to Schichten', async ({ page }) => {
    await page.getByPlaceholder('z. B. Sommerturnier 2026').fill('E2E Testturnier');
    await page.getByPlaceholder('z. B. Sporthalle Aachen-Brand').fill('Vereinsheim');
    await page.locator('input[type="date"]').first().fill('2026-08-01');
    await page.getByRole('button', { name: 'Weiter' }).click();
    await expect(page.locator('.sm-title')).toHaveText('Schichten');
  });
});
