import React from 'react';
import { Icon } from '@/components/ui/Icon';
import { Avatar } from '@/components/ui/Avatar';
import { useAppStore } from '@/store/app.store';
import { useAuthStore } from '@/store/auth.store';
import { useBranding } from '@/api/settings';

const DEFAULT_CLUB = 'TSC Schwarz-Gelb Aachen';

/** Builds short initials from a club name for the logo badge. */
function clubInitials(name: string): string {
  const words = name.split(/\s+/).filter(Boolean);
  if (words.length === 0) return 'SG';
  if (words.length === 1) return words[0].slice(0, 2).toUpperCase();
  return (words[0][0] + words[words.length - 1][0]).toUpperCase();
}

interface NavTab {
  key: string;
  label: string;
  icon: string;
}

interface DesktopShellProps {
  tabs: NavTab[];
  children: React.ReactNode;
  onOpenCreate?: () => void;
}

// Fallback user for demo/dev when no auth token exists
const DEMO_USER = {
  id: 'm-jonas',
  name: 'Jonas Berger',
  first: 'Jonas',
  last: 'Berger',
  email: 'jonas.berger@example.de',
  role: 'mitglied' as const,
};

export function DesktopShell({ tabs, children, onOpenCreate: _onOpenCreate }: DesktopShellProps) {
  const { role, setRole, tab, go, navStack } = useAppStore();
  const { user } = useAuthStore();
  const { data: branding } = useBranding();
  const me = user ?? DEMO_USER;
  const clubName = branding?.clubName ?? DEFAULT_CLUB;

  const memberMap: Record<string, { first: string; last: string }> = {
    [me.id]: { first: me.first ?? me.name.split(' ')[0], last: me.last ?? me.name.split(' ')[1] ?? '' },
  };

  return (
    <div className="stage desktop">
      <aside className="dt-sidebar">
        <div className="dt-brand">
          {branding?.logoUrl ? (
            <img src={branding.logoUrl} alt={clubName} style={{ width: 38, height: 38, borderRadius: '50%', objectFit: 'cover' }} />
          ) : (
            <span className="rb-logo" style={{ width: 38, height: 38, fontSize: 15 }}>
              {clubInitials(clubName)}
            </span>
          )}
          <div>
            <div className="dt-brand-name">ShiftManager</div>
            <div className="dt-brand-sub">{clubName}</div>
          </div>
        </div>

        <nav className="dt-nav">
          {tabs.map(({ key, label, icon }) => (
            <button
              key={key}
              className={'dt-navitem' + (tab === key && !navStack.length ? ' active' : '')}
              onClick={() => go(key)}
            >
              <Icon name={icon} size={20} stroke={2} />
              {label}
            </button>
          ))}
        </nav>

        <div className="dt-side-foot">
          <div className="dt-role-label">Ansicht wechseln</div>
          <div className="dt-roleseg">
            {([['mitglied', 'Mitglied'], ['vorstand', 'Vorstand']] as const).map(([v, l]) => (
              <button
                key={v}
                className={role === v ? 'active' : ''}
                onClick={() => setRole(v)}
              >
                {l}
              </button>
            ))}
          </div>
          <div className="dt-user">
            <Avatar
              memberId={me.id}
              members={memberMap}
              size={36}
            />
            <div style={{ flex: 1, minWidth: 0 }}>
              <div className="dt-user-name">
                {me.first ?? me.name.split(' ')[0]} {me.last ?? me.name.split(' ')[1] ?? ''}
              </div>
              <div className="dt-user-role">{role === 'mitglied' ? 'Mitglied' : 'Vorstand'}</div>
            </div>
          </div>
        </div>
      </aside>

      <main className="dt-content">
        <div className="sm-app desktop">
          {children}
        </div>
      </main>
    </div>
  );
}
