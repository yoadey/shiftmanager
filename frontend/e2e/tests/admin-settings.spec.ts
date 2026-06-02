import { test, expect } from '@playwright/test';
import { gotoApp, clickTab, expectNoError } from '../helpers/app';

// Runs with admin storageState (chromium project).

test.describe('Admin Einstellungen', () => {
  test.beforeEach(async ({ page }) => {
    await gotoApp(page);
    await clickTab(page, 'Einstellungen');
  });

  test('settings screen loads without error boundary', async ({ page }) => {
    await expectNoError(page);
  });

  test('settings page heading is visible', async ({ page }) => {
    await expect(page.getByText('Einstellungen').first()).toBeVisible();
  });

  test('Stundenziel section is rendered', async ({ page }) => {
    await expect(page.getByText('Stundenziel')).toBeVisible({ timeout: 8_000 });
    await expect(page.getByText('Globales Jahresziel')).toBeVisible();
  });

  test('Datenschutz section with Namensanzeige toggle is rendered', async ({ page }) => {
    await expect(page.getByText('Datenschutz')).toBeVisible({ timeout: 8_000 });
    await expect(page.getByText('Namensanzeige')).toBeVisible();
    // Both segment options present
    await expect(page.getByRole('button', { name: 'Abgekürzt', exact: true })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Vollständig', exact: true })).toBeVisible();
  });

  test('Namensanzeige toggle switches between modes', async ({ page }) => {
    await page.getByText('Datenschutz').waitFor({ timeout: 8_000 });

    const full = page.getByRole('button', { name: 'Vollständig', exact: true });
    const abbrev = page.getByRole('button', { name: 'Abgekürzt', exact: true });

    await full.click();
    await expect(page.getByText(/voller Name für alle/i)).toBeVisible();

    await abbrev.click();
    await expect(page.getByText(/Nachname gekürzt/i)).toBeVisible();
  });

  test('Reservierungen section is rendered', async ({ page }) => {
    await expect(page.getByText('Reservierungen')).toBeVisible({ timeout: 8_000 });
    await expect(page.getByText('Reservierung gültig für')).toBeVisible();
    await expect(page.getByText('Abmeldefrist vor Beginn')).toBeVisible();
  });

  test('Abrechnung section with Abrechnungsmodus is rendered', async ({ page }) => {
    // Section is below the fold — scroll to it first
    const section = page.getByText('Abrechnung').first();
    await section.waitFor({ timeout: 8_000 });
    await section.scrollIntoViewIfNeeded();
    await expect(page.getByText('Abrechnungsmodus')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Manuell', exact: true })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Automatisch', exact: true })).toBeVisible();
  });

  test('Branding section with colour picker is rendered', async ({ page }) => {
    await expect(page.getByText('Branding')).toBeVisible({ timeout: 8_000 });
    await expect(page.getByText('Primärfarbe')).toBeVisible();
    await expect(page.getByLabel('Primärfarbe')).toBeVisible();
    await expect(page.getByText('Vereinslogo')).toBeVisible();
  });

  test('E-Mail section links are visible for admins', async ({ page }) => {
    await expect(page.getByText('E-Mail').first()).toBeVisible({ timeout: 8_000 });
    await expect(page.getByText('E-Mail-Vorlagen')).toBeVisible();
    await expect(page.getByText('E-Mail-Protokoll')).toBeVisible();
  });

  test('System section with Audit-Log link is visible', async ({ page }) => {
    await expect(page.getByText('System')).toBeVisible({ timeout: 8_000 });
    await expect(page.getByText('Audit-Log')).toBeVisible();
  });

  test('Konto section with Abmelden link is visible', async ({ page }) => {
    await expect(page.getByText('Konto')).toBeVisible({ timeout: 8_000 });
    await expect(page.getByText('Abmelden')).toBeVisible();
  });

  test('Abmelden logs out and redirects to login', async ({ page }) => {
    await page.getByText('Konto').waitFor({ timeout: 8_000 });
    await page.getByText('Abmelden').click();
    await expect(page).toHaveURL(/\/auth\/login/, { timeout: 5_000 });
  });
});
