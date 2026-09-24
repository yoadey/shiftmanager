import React, { useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { Icon } from '@/components/ui/Icon';
import { Sheet } from '@/components/ui/Sheet';
import { isSectionStart, type NavTab } from '@/utils/navTabs';

// The bottom bar evenly spreads up to this many items; once the
// veranstaltungsleiter/vorstand/admin-only tabs are appended it no longer
// fits, so the rest are tucked behind a "Mehr" button that opens a menu
// instead of making the bar scrollable.
const MAX_VISIBLE_TABS = 4;

interface MobileShellProps {
  tabs: NavTab[];
  children: React.ReactNode;
}

export function MobileShell({ tabs, children }: MobileShellProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const [moreOpen, setMoreOpen] = useState(false);

  const overflow = tabs.length > MAX_VISIBLE_TABS;
  const visibleTabs = overflow ? tabs.slice(0, MAX_VISIBLE_TABS - 1) : tabs;
  const overflowTabs = overflow ? tabs.slice(MAX_VISIBLE_TABS - 1) : [];
  const isActive = (key: string) => location.pathname === '/' + key;
  const overflowActive = overflowTabs.some((t) => isActive(t.key));
  // The bottom nav only makes sense at a tab's own root — on a detail screen
  // (e.g. an event or member) it would just be in the way.
  const isTopLevel = tabs.some((t) => isActive(t.key));

  return (
    <div className="sm-app">
      <div className="sm-main" key={location.pathname}>
        {children}
      </div>

      {isTopLevel && (
        <nav className="sm-nav">
          {visibleTabs.map(({ key, label, icon }, i) => (
            <React.Fragment key={key}>
              {isSectionStart(visibleTabs, i) && <span className="nav-sep" />}
              <button
                className={'sm-navitem' + (isActive(key) ? ' active' : '')}
                onClick={() => navigate('/' + key)}
              >
                <span className="ni-ico">
                  <Icon name={icon} size={21} stroke={isActive(key) ? 2.4 : 2} />
                </span>
                {label}
              </button>
            </React.Fragment>
          ))}
          {overflow && (
            <button
              className={'sm-navitem' + (overflowActive ? ' active' : '')}
              onClick={() => setMoreOpen(true)}
            >
              <span className="ni-ico">
                <Icon name="more" size={21} stroke={overflowActive ? 2.4 : 2} />
              </span>
              Mehr
            </button>
          )}
        </nav>
      )}

      {moreOpen && (
        <Sheet onClose={() => setMoreOpen(false)} title="Mehr">
          <div className="sm-more-list">
            {overflowTabs.map(({ key, label, icon }, i) => (
              <React.Fragment key={key}>
                {isSectionStart(overflowTabs, i) && <div className="sm-more-sep" />}
                <button
                  className={'sm-more-item' + (isActive(key) ? ' active' : '')}
                  onClick={() => {
                    navigate('/' + key);
                    setMoreOpen(false);
                  }}
                >
                  <span className="ni-ico">
                    <Icon name={icon} size={20} stroke={isActive(key) ? 2.4 : 2} />
                  </span>
                  {label}
                </button>
              </React.Fragment>
            ))}
          </div>
        </Sheet>
      )}
    </div>
  );
}
