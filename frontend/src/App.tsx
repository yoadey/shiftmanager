import React, { useEffect, useState, lazy, Suspense } from 'react';
import { DesktopShell } from '@/components/layout/DesktopShell';
import { MobileShell } from '@/components/layout/MobileShell';
import { Toast } from '@/components/ui/Toast';
import { TweaksPanel } from '@/components/tweaks/TweaksPanel';
import { useAppStore } from '@/store/app.store';

// Screens — lazy loaded
const MemberDashboard = lazy(() => import('@/screens/member/MemberDashboard').then(m => ({ default: m.MemberDashboard })));
const MemberDiscover = lazy(() => import('@/screens/member/MemberDiscover').then(m => ({ default: m.MemberDiscover })));
const MyShifts = lazy(() => import('@/screens/member/MyShifts').then(m => ({ default: m.MyShifts })));
const MemberProfile = lazy(() => import('@/screens/member/MemberProfile').then(m => ({ default: m.MemberProfile })));
const EventDetail = lazy(() => import('@/screens/events/EventDetail').then(m => ({ default: m.EventDetail })));
const AdminDashboard = lazy(() => import('@/screens/admin/AdminDashboard').then(m => ({ default: m.AdminDashboard })));
const AdminEvents = lazy(() => import('@/screens/admin/AdminEvents').then(m => ({ default: m.AdminEvents })));
const AdminMembers = lazy(() => import('@/screens/admin/AdminMembers').then(m => ({ default: m.AdminMembers })));
const MemberDetail = lazy(() => import('@/screens/admin/MemberDetail').then(m => ({ default: m.MemberDetail })));
const ManualBooking = lazy(() => import('@/screens/admin/ManualBooking').then(m => ({ default: m.ManualBooking })));
const AdminSettings = lazy(() => import('@/screens/admin/AdminSettings').then(m => ({ default: m.AdminSettings })));
const AuditLog = lazy(() => import('@/screens/admin/AuditLog').then(m => ({ default: m.AuditLog })));
const CreateEventFlow = lazy(() => import('@/screens/admin/CreateEventFlow').then(m => ({ default: m.CreateEventFlow })));

function pickOn(hex: string): string {
  const h = hex.replace('#', '');
  const r = parseInt(h.slice(0, 2), 16);
  const g = parseInt(h.slice(2, 4), 16);
  const b = parseInt(h.slice(4, 6), 16);
  const lum = (0.299 * r + 0.587 * g + 0.114 * b) / 255;
  return lum > 0.6 ? '#1A1813' : '#ffffff';
}

const MEMBER_TABS = [
  { key: 'start', label: 'Start', icon: 'home' },
  { key: 'entdecken', label: 'Entdecken', icon: 'compass' },
  { key: 'schichten', label: 'Schichten', icon: 'calendar' },
  { key: 'profil', label: 'Profil', icon: 'user' },
];

const ADMIN_TABS = [
  { key: 'start', label: 'Übersicht', icon: 'chart' },
  { key: 'events', label: 'Termine', icon: 'calendar' },
  { key: 'mitglieder', label: 'Mitglieder', icon: 'users' },
  { key: 'settings', label: 'Einstellungen', icon: 'settings' },
];

function LoadingFallback() {
  return (
    <div style={{ padding: 40, textAlign: 'center', color: 'var(--muted)', fontWeight: 600 }}>
      Lädt…
    </div>
  );
}

function ScreenRouter() {
  const { role, tab, navStack } = useAppStore();

  if (navStack.length > 0) {
    const top = navStack[navStack.length - 1];
    if (top.name === 'event') return <EventDetail id={top.params.id ?? ''} />;
    if (top.name === 'member') return <MemberDetail id={top.params.id ?? ''} />;
    if (top.name === 'manual') return <ManualBooking id={top.params.id} />;
    if (top.name === 'audit') return <AuditLog />;
  }

  if (role === 'mitglied') {
    if (tab === 'start') return <MemberDashboard />;
    if (tab === 'entdecken') return <MemberDiscover />;
    if (tab === 'schichten') return <MyShifts />;
    if (tab === 'profil') return <MemberProfile />;
  } else {
    if (tab === 'start') return <AdminDashboard />;
    if (tab === 'events') return <AdminEvents />;
    if (tab === 'mitglieder') return <AdminMembers />;
    if (tab === 'settings') return <AdminSettings />;
  }
  return null;
}

export default function App() {
  const { role, tweaks } = useAppStore();
  const [isDesktop, setIsDesktop] = useState(() => window.innerWidth >= 900);
  const [createOpen, setCreateOpen] = useState(false);

  // Apply CSS custom properties from tweaks
  useEffect(() => {
    const root = document.documentElement;
    root.style.setProperty('--primary', tweaks.primaryColor);
    root.style.setProperty('--on-primary', pickOn(tweaks.primaryColor));

    const radii: Record<string, [string, string]> = {
      klein: ['12px', '9px'],
      standard: ['20px', '13px'],
      weich: ['28px', '18px'],
    };
    const [rad, radSm] = radii[tweaks.radius] ?? radii.standard;
    root.style.setProperty('--radius', rad);
    root.style.setProperty('--radius-sm', radSm);

    if (tweaks.warmth === 'neutral') {
      root.style.setProperty('--bg', '#F5F5F3');
      root.style.setProperty('--surface-2', '#F3F3F1');
      root.style.setProperty('--line', '#EAEAE7');
      root.style.setProperty('--line-2', '#DEDEDA');
    } else {
      root.style.setProperty('--bg', '#FBF7EF');
      root.style.setProperty('--surface-2', '#FAF6EC');
      root.style.setProperty('--line', '#ECE6D8');
      root.style.setProperty('--line-2', '#E2DBCB');
    }
  }, [tweaks.primaryColor, tweaks.radius, tweaks.warmth]);

  useEffect(() => {
    const check = () => setIsDesktop(window.innerWidth >= 900);
    window.addEventListener('resize', check);
    return () => window.removeEventListener('resize', check);
  }, []);

  const tabs = role === 'mitglied' ? MEMBER_TABS : ADMIN_TABS;

  const screenContent = (
    <Suspense fallback={<LoadingFallback />}>
      <ScreenRouter />
      {createOpen && (
        <CreateEventFlow onClose={() => setCreateOpen(false)} />
      )}
      <Toast />
    </Suspense>
  );

  if (isDesktop) {
    return (
      <>
        <DesktopShell tabs={tabs} onOpenCreate={() => setCreateOpen(true)}>
          <div className="sm-main">
            {screenContent}
          </div>
        </DesktopShell>
        <TweaksPanel />
      </>
    );
  }

  return (
    <>
      <MobileShell tabs={tabs}>
        {screenContent}
      </MobileShell>
      <TweaksPanel />
    </>
  );
}
