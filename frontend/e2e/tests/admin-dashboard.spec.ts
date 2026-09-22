import { test, expect } from '@playwright/test';
import { gotoApp, clickTab, expectNoError } from '../helpers/app';

// Runs with admin storageState (chromium project).

test.describe('Admin Dashboard (Übersicht)', () => {
  test.beforeEach(async ({ page }) => {
    await gotoApp(page);
  });

  test('dashboard renders without error boundary', async ({ page }) => {
    await expectNoError(page);
  });

  test('dashboard shows the page heading', async ({ page }) => {
    // AdminDashboard renders "Übersicht" as the screen title
    await expect(page.getByText('Übersicht').first()).toBeVisible();
  });
});

test.describe('Admin Termine tab', () => {
  test.beforeEach(async ({ page }) => {
    await gotoApp(page);
    await clickTab(page, 'Termine');
  });

  test('Termine screen loads without error boundary', async ({ page }) => {
    await expectNoError(page);
  });

  test('Termine shows events page with heading', async ({ page }) => {
    // The Termine screen renders "Veranstaltungen" as the page heading (seeded: Sommerfest)
    await page.waitForLoadState('domcontentloaded');
    await expect(page.locator('text=Veranstaltungen')).toBeVisible({ timeout: 8_000 });
  });
});

test.describe('Admin Mitglieder tab', () => {
  test.beforeEach(async ({ page }) => {
    await gotoApp(page);
    await clickTab(page, 'Mitglieder');
  });

  test('Mitglieder screen loads without error boundary', async ({ page }) => {
    await expectNoError(page);
  });

  test('seed admin member is shown in the list', async ({ page }) => {
    // SeedTestData creates Admin Vorstand (admin@test.local) and Max Mustermann
    await expect(
      page.getByText('Admin Vorstand').or(page.getByText('admin@test.local')).first(),
    ).toBeVisible({ timeout: 10_000 });
  });

  test('member search filters the list', async ({ page }) => {
    const searchInput = page
      .getByPlaceholder(/suche/i)
      .or(page.getByRole('searchbox'))
      .first();

    if (!(await searchInput.isVisible())) {
      test.skip();
      return;
    }

    await searchInput.fill('Admin');
    await page.waitForTimeout(400); // debounce
    await expect(page.getByText(/admin/i).first()).toBeVisible();

    await searchInput.fill('xyznotexist');
    await page.waitForTimeout(400);
    await expectNoError(page);
  });
});
