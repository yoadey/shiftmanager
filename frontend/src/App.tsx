import { useEffect, useState, useRef, lazy, Suspense, type ReactNode } from 'react';
import { Routes, Route, Navigate, useLocation, useParams } from 'react-router-dom';
import { DesktopShell } from '@/components/layout/DesktopShell';
import { MobileShell } from '@/components/layout/MobileShell';
import { Toast } from '@/components/ui/Toast';
import { ErrorBoundary } from '@/components/ui/ErrorBoundary';
import { TweaksPanel } from '@/components/tweaks/TweaksPanel';
import { useAppStore } from '@/store/app.store';
import { useAuthStore } from '@/store/auth.store';
import { useBranding, useSettings } from '@/api/settings';
import { useTokenRefresh } from '@/hooks/useTokenRefresh';
import { useIsVorstand } from '@/hooks/useIsVorstand';
import { useIsEventManager } from '@/hooks/useIsEventManager';
import { useSmartBack } from '@/hooks/useSmartBack';
import { routes, homePathForRole } from '@/routes';

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
const EmailTemplates = lazy(() => import('@/screens/admin/EmailTemplates').then(m => ({ default: m.EmailTemplates })));
const EmailLog = lazy(() => import('@/screens/admin/EmailLog').then(m => ({ default: m.EmailLog })));
const CreateEventFlow = lazy(() => import('@/screens/admin/CreateEventFlow').then(m => ({ default: m.CreateEventFlow })));
const AdminBilling = lazy(() => import('@/screens/admin/AdminBilling').then(m => ({ default: m.AdminBilling })));

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

// Shown additionally, in the same list, for veranstaltungsleiter and above —
// no separate "view" to switch into; `section` marks where the extra group
// starts so the shells can set it apart visually. Matches the backend's own
// RequireRole(RoleVeranstaltungsleiter) gate on event/shift management.
const EVENT_MANAGER_TABS = [
  { key: 'uebersicht', label: 'Übersicht', icon: 'chart', section: 'Verwaltung' },
  { key: 'events', label: 'Termine', icon: 'calendar' },
];

// Shown additionally for vorstand/admin only — member management, billing
// and settings are RequireRole(RoleVorstand) on the backend, a strictly
// higher tier than event management.
const BOARD_TABS = [
  { key: 'mitglieder', label: 'Mitglieder', icon: 'users' },
  { key: 'abrechnungen', label: 'Abrechnungen', icon: 'euro' },
  { key: 'settings', label: 'Einstellungen', icon: 'settings' },
];

function LoadingFallback() {
  return (
    <div style={{ padding: 40, textAlign: 'center', color: 'var(--muted)', fontWeight: 600 }}>
      Lädt…
    </div>
  );
}

// Guards a route on a role check, redirecting to that role's own home screen
// instead of rendering a screen the current role no longer has access to
// (e.g. a stale bookmark from before a role downgrade).
function RequireRole({ allowed, children }: { allowed: boolean; children: ReactNode }) {
  const role = useAuthStore((s) => s.user?.role);
  if (!allowed) return <Navigate to={homePathForRole(role)} replace />;
  return <>{children}</>;
}

// CreateEventFlow is a full-screen Sheet (variant="full"), so it needs no
// backdrop screen behind it — routed on its own at /events/neu regardless of
// which screen (AdminEvents or AdminDashboard) opened it.
function CreateEventFlowRoute() {
  const closeModal = useSmartBack(routes.events);
  return <CreateEventFlow onClose={closeModal} />;
}

// /stunden-buchen has no member preselected; /mitglieder/:id/stunden-buchen
// preselects the member whose detail page opened it.
function ManualBookingRoute() {
  const { id } = useParams();
  return <ManualBooking id={id} />;
}

function ScreenRoutes() {
  const isVorstand = useIsVorstand();
  const isEventManager = useIsEventManager();
  const role = useAuthStore((s) => s.user?.role);
  const homePath = homePathForRole(role);

  return (
    <Routes>
      <Route path="/" element={<Navigate to={homePath} replace />} />

      <Route path="start" element={<MemberDashboard />} />
      <Route path="entdecken" element={<MemberDiscover />} />
      <Route path="schichten" element={<MyShifts />} />
      <Route path="profil" element={<MemberProfile />} />

      <Route path="uebersicht" element={<RequireRole allowed={isEventManager}><AdminDashboard /></RequireRole>} />
      <Route path="events" element={<RequireRole allowed={isEventManager}><AdminEvents /></RequireRole>} />
      <Route path="events/neu" element={<RequireRole allowed={isEventManager}><CreateEventFlowRoute /></RequireRole>} />
      <Route path="events/:id/*" element={<RequireRole allowed={isEventManager}><EventDetail /></RequireRole>} />

      <Route path="stunden-buchen" element={<RequireRole allowed={isEventManager}><ManualBookingRoute /></RequireRole>} />

      <Route path="mitglieder" element={<RequireRole allowed={isVorstand}><AdminMembers /></RequireRole>} />
      <Route path="mitglieder/neu" element={<RequireRole allowed={isVorstand}><AdminMembers /></RequireRole>} />
      <Route path="mitglieder/import" element={<RequireRole allowed={isVorstand}><AdminMembers /></RequireRole>} />
      <Route path="mitglieder/:id/stunden-buchen" element={<RequireRole allowed={isVorstand}><ManualBookingRoute /></RequireRole>} />
      <Route path="mitglieder/:id/*" element={<RequireRole allowed={isVorstand}><MemberDetail /></RequireRole>} />

      <Route path="abrechnungen/*" element={<RequireRole allowed={isVorstand}><AdminBilling /></RequireRole>} />

      <Route path="settings" element={<RequireRole allowed={isVorstand}><AdminSettings /></RequireRole>} />
      <Route path="settings/branding-verlauf" element={<RequireRole allowed={isVorstand}><AdminSettings /></RequireRole>} />
      <Route path="settings/email-vorlagen" element={<RequireRole allowed={isVorstand}><EmailTemplates /></RequireRole>} />
      <Route path="settings/email-vorlagen/:name" element={<RequireRole allowed={isVorstand}><EmailTemplates /></RequireRole>} />
      <Route path="settings/email-log" element={<RequireRole allowed={isVorstand}><EmailLog /></RequireRole>} />
      <Route path="settings/audit" element={<RequireRole allowed={isVorstand}><AuditLog /></RequireRole>} />

      <Route path="*" element={<Navigate to={homePath} replace />} />
    </Routes>
  );
}

export default function App() {
  const { tweaks, setTweak, setNameMode } = useAppStore();
  const location = useLocation();
  const isVorstand = useIsVorstand();
  const isEventManager = useIsEventManager();
  // Reset the screen error boundary whenever the user navigates, so a crash on
  // one screen never sticks after switching tabs or drilling into a detail.
  const navKey = location.pathname;
  const [isDesktop, setIsDesktop] = useState(() => window.innerWidth >= 900);

  // A-004: keep the session alive in the background while the app is open.
  useTokenRefresh();

  // Branding (B-002/B-005/B-006) — applied at app root, falls back to defaults while loading.
  const { data: branding } = useBranding();
  const { data: settings } = useSettings();
  const brandingAppliedRef = useRef(false);

  // Hydrate the primary colour from branding once (TweaksPanel stays the live source after that).
  useEffect(() => {
    if (branding?.primaryColor && !brandingAppliedRef.current) {
      brandingAppliedRef.current = true;
      setTweak('primaryColor', branding.primaryColor);
    }
  }, [branding?.primaryColor, setTweak]);

  // Accent colour derived from the brand primary.
  useEffect(() => {
    document.documentElement.style.setProperty('--accent', tweaks.primaryColor);
  }, [tweaks.primaryColor]);

  // NM-005: hydrate the global name-display mode from the settings API.
  useEffect(() => {
    if (settings?.nameMode) setNameMode(settings.nameMode);
  }, [settings?.nameMode, setNameMode]);

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

  const tabs = [
    ...MEMBER_TABS,
    ...(isEventManager ? EVENT_MANAGER_TABS : []),
    ...(isVorstand ? BOARD_TABS : []),
  ];

  const screenContent = (
    <Suspense fallback={<LoadingFallback />}>
      <ErrorBoundary key={navKey} label="diesem Bereich">
        <ScreenRoutes />
      </ErrorBoundary>
      <ErrorBoundary>
        <Toast />
      </ErrorBoundary>
    </Suspense>
  );

  if (isDesktop) {
    return (
      <>
        <DesktopShell tabs={tabs}>
          <div className="sm-main">
            {screenContent}
          </div>
        </DesktopShell>
        <ErrorBoundary>
          <TweaksPanel />
        </ErrorBoundary>
      </>
    );
  }

  return (
    <>
      <MobileShell tabs={tabs}>
        {screenContent}
      </MobileShell>
      <ErrorBoundary>
        <TweaksPanel />
      </ErrorBoundary>
    </>
  );
}
