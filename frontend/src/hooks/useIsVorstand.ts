import { useAuthStore } from '@/store/auth.store';

/**
 * True for vorstand/admin — the same boundary the backend uses for its
 * board-only routes (internal/domain/member.go: `HasRole(role, RoleVorstand)`,
 * RoleLevel vorstand=3/admin=4) and for NM-003's "board always sees full
 * names" rule. Deliberately excludes veranstaltungsleiter: that role can
 * manage events/shifts but not members, billing or settings, so it does not
 * get the admin-only nav items (a veranstaltungsleiter landing on e.g.
 * Abrechnungen would just get 403s from every request).
 */
export function useIsVorstand(): boolean {
  const role = useAuthStore((s) => s.user?.role);
  return role === 'vorstand' || role === 'admin';
}
