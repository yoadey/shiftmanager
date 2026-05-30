import React, { useState, useEffect } from 'react';
import { Icon } from '@/components/ui/Icon';
import { IOSFrame } from '@/components/device/IOSFrame';
import { useAppStore } from '@/store/app.store';

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
        <div className="rb-brand">
          <span className="rb-logo">SG</span>
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
