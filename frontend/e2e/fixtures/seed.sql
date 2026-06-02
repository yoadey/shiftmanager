-- E2E test fixtures — inserted before every test run, idempotent via ON CONFLICT.
-- UUIDs are fixed so the JWTs generated in global.setup.ts always match.

INSERT INTO members (id, first_name, last_name, email, joined_at, is_active, role)
VALUES
  ('00000000-0000-0000-0000-e2e000000001', 'E2E',    'Admin',  'e2e-admin@test.local',  NOW(), true,  'admin'),
  ('00000000-0000-0000-0000-e2e000000002', 'E2E',    'Member', 'e2e-member@test.local', NOW(), true,  'mitglied'),
  ('00000000-0000-0000-0000-e2e000000003', 'E2E',    'Inactive','e2e-inactive@test.local',NOW(),false,'mitglied')
ON CONFLICT (id) DO UPDATE
  SET first_name = EXCLUDED.first_name,
      last_name  = EXCLUDED.last_name,
      email      = EXCLUDED.email,
      is_active  = EXCLUDED.is_active,
      role       = EXCLUDED.role;
