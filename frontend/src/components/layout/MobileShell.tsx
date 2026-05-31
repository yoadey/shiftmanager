import React, { useState, useEffect } from 'react';
import { Icon } from '@/components/ui/Icon';
import { IOSFrame } from '@/components/device/IOSFrame';
import { useAppStore } from '@/store/app.store';
import { useBranding } from '@/api/settings';

const DEFAULT_CLUB = 'TSC Schwarz-Gelb Aachen';

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

interface MobileShellProps {
  tabs: NavTab[];
  children: React.ReactNode;
}

export function MobileShell({ tabs, children }: MobileShellProps) {
  const { role, setRole, tab, go, navStack } = useAppStore();
  const { data: branding } = useBranding();
  const clubName = branding?.clubName ?? DEFAULT_CLUB;
  const [scale, setScale] = useState(1);

  useEffect(() => {
    const fit = () => {
      const pad = window.innerWidth < 520 ? 0 : 24;
      const s = Math.min(
        1,
        (window.innerHeight - 86 - pad) / 874,
        (window.innerWidth - pad) / 402,
      );
      setScale(Math.max(0.4, s));
    };
    fit();
    window.addEventListener('resize', fit);
    return () => window.removeEventListener('resize', fit);
  }, []);

  return (
    <div className="stage">
      {/* Role switcher chrome */}
      <div className="rolebar">
        <div className="rb-brand" title={clubName}>
          {branding?.logoUrl ? (
            <img src={branding.logoUrl} alt={clubName} className="rb-logo" style={{ objectFit: 'cover' }} />
          ) : (
            <span className="rb-logo">{clubInitials(clubName)}</span>
          )}
          <span>ShiftManager</span>
        </div>
        <div className="rb-seg">
          {([['mitglied', 'Mitglied'], ['vorstand', 'Vorstand']] as const).map(([v, l]) => (
            <button
              key={v}
              className={'rb-btn' + (role === v ? ' active' : '')}
              onClick={() => setRole(v)}
            >
              {l}
            </button>
          ))}
        </div>
      </div>

      <div className="scaler" style={{ transform: `scale(${scale})` }}>
        <IOSFrame>
          <div className="sm-app">
            <div className="sm-main" key={role + tab + navStack.length}>
              {children}
            </div>

            {navStack.length === 0 && (
              <nav className="sm-nav">
                {tabs.map(({ key, label, icon }) => (
                  <button
                    key={key}
                    className={'sm-navitem' + (tab === key ? ' active' : '')}
                    onClick={() => go(key)}
                  >
                    <span className="ni-ico">
                      <Icon name={icon} size={21} stroke={tab === key ? 2.4 : 2} />
                    </span>
                    {label}
                  </button>
                ))}
              </nav>
            )}
          </div>
        </IOSFrame>
      </div>
    </div>
  );
}
