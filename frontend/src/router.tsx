import { lazy, Suspense, type ReactNode } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import App from './App';
import { useAuthStore } from '@/store/auth.store';
import { ErrorBoundary } from '@/components/ui/ErrorBoundary';

// RequireAuth gates the main application behind a session token. Public routes
// (kiosk, login, callback, shift confirmation) stay outside this guard.
function RequireAuth({ children }: { children: ReactNode }) {
  const token = useAuthStore((s) => s.token);
  if (!token) return <Navigate to="/auth/login" replace />;
  return <>{children}</>;
}

const KioskPage = lazy(() => import('@/screens/kiosk/KioskPage').then((m) => ({ default: m.KioskPage })));
const LoginPage = lazy(() => import('@/screens/auth/LoginPage').then((m) => ({ default: m.LoginPage })));
const CallbackPage = lazy(() => import('@/screens/auth/CallbackPage').then((m) => ({ default: m.CallbackPage })));
const ConfirmPage = lazy(() => import('@/screens/ConfirmPage').then((m) => ({ default: m.ConfirmPage })));

function RouteFallback() {
  return (
    <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--muted)', fontWeight: 600, background: 'var(--bg)' }}>
      Lädt…
    </div>
  );
}

export default function AppRouter() {
  return (
    <ErrorBoundary>
      <Suspense fallback={<RouteFallback />}>
        <Routes>
          <Route path="/" element={<RequireAuth><App /></RequireAuth>} />
          <Route path="/kiosk" element={<ErrorBoundary label="dem Kiosk"><KioskPage /></ErrorBoundary>} />
          <Route path="/auth/login" element={<ErrorBoundary><LoginPage /></ErrorBoundary>} />
          <Route path="/auth/callback" element={<ErrorBoundary><CallbackPage /></ErrorBoundary>} />
          <Route path="/shifts/confirm/:token" element={<ErrorBoundary><ConfirmPage /></ErrorBoundary>} />
        </Routes>
      </Suspense>
    </ErrorBoundary>
  );
}
