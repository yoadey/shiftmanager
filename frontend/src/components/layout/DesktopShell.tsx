import React from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { Icon } from '@/components/ui/Icon';
import { Avatar } from '@/components/ui/Avatar';
import { useAuthStore } from '@/store/auth.store';
import { useBranding } from '@/api/settings';
import { isSectionStart, type NavTab } from '@/utils/navTabs';

const ROLE_LABELS: Record<string, string> = {
  mitglied: 'Mitglied',
  veranstaltungsleiter: 'Veranstaltungsleiter',
  vorstand: 'Vorstand',
  admin: 'Admin',
};

const DEFAULT_CLUB = 'TSC Schwarz-Gelb Aachen';

/** Builds short initials from a club name for the logo badge. */
function clubInitials(name: string): string {
  const words = name.split(/\s+/).filter(Boolean);
  if (words.length === 0) return 'SG';
  if (words.length === 1) return words[0].slice(0, 2).toUpperCase();
  return (words[0][0] + words[words.length - 1][0]).toUpperCase();
}

interface DesktopShellProps {
  tabs: NavTab[];
  children: React.ReactNode;
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

export function DesktopShell({ tabs, children }: DesktopShellProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const { user } = useAuthStore();
  const { data: branding } = useBranding();
  const me = user ?? DEMO_USER;
  const clubName = branding?.clubName ?? DEFAULT_CLUB;
  const roleLabel = ROLE_LABELS[me.role] ?? me.role;

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
          {tabs.map(({ key, label, icon, section }, i) => (
            <React.Fragment key={key}>
              {isSectionStart(tabs, i) && (
                <div className="dt-role-label" style={{ marginTop: i > 0 ? 14 : 0 }}>{section}</div>
              )}
              <button
                className={'dt-navitem' + (location.pathname === '/' + key ? ' active' : '')}
                onClick={() => navigate('/' + key)}
              >
                <Icon name={icon} size={20} stroke={2} />
                {label}
              </button>
            </React.Fragment>
          ))}
        </nav>

        <div className="dt-side-foot">
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
              <div className="dt-user-role">{roleLabel}</div>
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
