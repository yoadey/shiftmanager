import { test as setup } from '@playwright/test';
import { mkdirSync } from 'fs';
import { resolve } from 'path';
import { fileURLToPath } from 'url';
import { E2E_BACKEND_PORT } from '../playwright.config';

const __dirname = fileURLToPath(new URL('.', import.meta.url));

// Fixed UUIDs from internal/adapter/db/seed.go (SeedTestData).
export const SEED_ADMIN_UUID  = '00000000-0000-0000-0000-000000000001';
export const SEED_MEMBER_UUID = '00000000-0000-0000-0000-000000000002';

const BACKEND = `http://localhost:${E2E_BACKEND_PORT}`;

async function devToken(role: string, id: string, email: string): Promise<string> {
  const url = `${BACKEND}/api/v1/dev/token?role=${role}&id=${id}&email=${encodeURIComponent(email)}`;
  const res = await fetch(url);
  if (!res.ok) throw new Error(`/dev/token failed: ${res.status} — is TEST_MODE=true running on port ${E2E_BACKEND_PORT}?`);
  const data = await res.json() as { token: string };
  return data.token;
}

setup('generate auth tokens via /dev/token', async ({ page }) => {
  mkdirSync(resolve(__dirname, '.auth'), { recursive: true });

  // ── Admin session ────────────────────────────────────────────────────────
  const adminToken = await devToken('vorstand', SEED_ADMIN_UUID, 'admin@test.local');

  await page.goto('/');
  await page.evaluate(
    ({ token, user }) => {
      localStorage.setItem('sm_token', token);
      localStorage.setItem('sm_user', user);
      // sm_role controls which tabs the shell renders
      localStorage.setItem('sm_role', 'vorstand');
    },
    {
      token: adminToken,
      user: JSON.stringify({
        id:    SEED_ADMIN_UUID,
        name:  'Admin Vorstand',
        email: 'admin@test.local',
        role:  'vorstand',
        first: 'Admin',
        last:  'Vorstand',
      }),
    },
  );
  await page.context().storageState({ path: resolve(__dirname, '.auth/admin.json') });

  // ── Member session ───────────────────────────────────────────────────────
  const memberToken = await devToken('mitglied', SEED_MEMBER_UUID, 'max@test.local');

  await page.evaluate(
    ({ token, user }) => {
      localStorage.setItem('sm_token', token);
      localStorage.setItem('sm_user', user);
      localStorage.setItem('sm_role', 'mitglied');
    },
    {
      token: memberToken,
      user: JSON.stringify({
        id:    SEED_MEMBER_UUID,
        name:  'Max Mustermann',
        email: 'max@test.local',
        role:  'mitglied',
        first: 'Max',
        last:  'Mustermann',
      }),
    },
  );
  await page.context().storageState({ path: resolve(__dirname, '.auth/member.json') });
});
