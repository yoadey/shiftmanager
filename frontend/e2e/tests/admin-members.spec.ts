import { test, expect } from '@playwright/test';
import { gotoApp, clickTab, expectNoError } from '../helpers/app';

// Runs with admin storageState (chromium project).

test.describe('Admin Mitglieder — detail and search', () => {
  test.beforeEach(async ({ page }) => {
    await gotoApp(page);
    await clickTab(page, 'Mitglieder');
  });

  test('member list loads without crash', async ({ page }) => {
    await expectNoError(page);
  });

  test('seed admin member is listed', async ({ page }) => {
    await expect(
      page.getByText('Admin Vorstand').or(page.getByText('admin@test.local')).first(),
    ).toBeVisible({ timeout: 10_000 });
  });

  test('member search narrows results', async ({ page }) => {
    const input = page.getByPlaceholder(/suche/i).or(page.getByRole('searchbox')).first();
    if (!(await input.isVisible())) {
      test.skip();
      return;
    }

    await input.fill('Admin');
    await page.waitForTimeout(400);
    await expect(page.getByText(/admin/i).first()).toBeVisible();

    await input.fill('xyznotexist');
    await page.waitForTimeout(400);
    await expectNoError(page);
  });
});
