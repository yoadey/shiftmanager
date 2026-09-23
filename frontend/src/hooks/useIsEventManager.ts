import { useAuthStore } from '@/store/auth.store';

/**
 * True for veranstaltungsleiter and above — the tier the backend grants
 * event/shift management to (internal/adapter/http/router.go:
 * `RequireRole(domain.RoleVeranstaltungsleiter)` on event/shift/registration
 * writes and the /stats read). Every vorstand/admin user is also an event
 * manager (RoleLevel is monotonic), so this is the broader of the two
 * "board" gates — see useIsVorstand for the stricter, vorstand-only one
 * (members, billing, settings).
 */
export function useIsEventManager(): boolean {
  const role = useAuthStore((s) => s.user?.role);
  return role !== undefined && role !== 'mitglied';
}
