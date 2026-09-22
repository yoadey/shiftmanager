import type { Page } from '@playwright/test';
import { expect } from '@playwright/test';

/** Navigate to / and wait until the shell (nav tabs) is ready. */
export async function gotoApp(page: Page) {
  await page.goto('/');
  // The sidebar nav is the earliest reliable signal that the shell rendered.
  await page.locator('nav.dt-nav, nav.mb-tabs').waitFor({ state: 'visible', timeout: 10_000 });
}

/** Click an admin tab by its label and wait for the content area to settle. */
export async function clickTab(page: Page, label: string) {
  await page.getByRole('button', { name: label, exact: true }).click();
  // Give React a tick to swap the screen.
  await page.waitForTimeout(200);
}

/** Assert no top-level error boundary is visible. */
export async function expectNoError(page: Page) {
  await expect(page.getByText('Ein Fehler ist aufgetreten.')).not.toBeVisible();
}
