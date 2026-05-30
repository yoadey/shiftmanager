import { useAppStore } from '@/store/app.store';
import { useAuthStore } from '@/store/auth.store';
import type { Member } from '@/types';

interface FormatOptions {
  /** Force full name regardless of nameMode (used by board/admin) */
  viewerFull?: boolean;
  /** Override global nameMode */
  mode?: 'abbrev' | 'full';
}

/**
 * NM-001..007 name formatting rules:
 * NM-002: abbreviate last name → "Maximilian M."
 * NM-003: board always sees full name
 * NM-006: own name always shown in full
 */
export function useNameFormat() {
  const { tweaks: _t, role } = useAppStore();
  const { user } = useAuthStore();

  const formatName = (member: Member | undefined | null, opts: FormatOptions = {}): string => {
    if (!member) return 'Unbekannt';
    const full = `${member.first} ${member.last}`;

    // NM-006: own name always full
    if (user && member.id === user.id) return full;

    // NM-003: board viewers always see full
    if (opts.viewerFull || role === 'vorstand') return full;

    // Explicit mode override
    if (opts.mode === 'full') return full;

    // Default abbreviation (NM-002)
    return `${member.first} ${member.last.charAt(0)}.`;
  };

  return formatName;
}
