import { lazy, Suspense } from 'react';
import { Routes, Route } from 'react-router-dom';
import App from './App';

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
    <Suspense fallback={<RouteFallback />}>
      <Routes>
        <Route path="/" element={<App />} />
        <Route path="/kiosk" element={<KioskPage />} />
        <Route path="/auth/login" element={<LoginPage />} />
        <Route path="/auth/callback" element={<CallbackPage />} />
        <Route path="/shifts/confirm/:token" element={<ConfirmPage />} />
      </Routes>
    </Suspense>
  );
}
