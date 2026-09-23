import React from 'react';
import ReactDOM from 'react-dom/client';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter } from 'react-router-dom';
import AppRouter from './router';
import { setupApiClient } from './api/setup';
import './styles/global.css';

// Wire up the generated API client (base URL + JWT auth) before rendering.
setupApiClient();

// Recover from stale-chunk errors after a redeploy: a tab left open (or a
// cached page) can still reference a hashed asset file that a newer deploy
// no longer serves, so a lazy-loaded screen's dynamic import() 404s. Vite
// dispatches this event for every such failure; reload once to pick up the
// current index.html and its (now-matching) chunk hashes. Guarded with a
// per-session flag so a genuinely broken deploy shows the error instead of
// reload-looping forever.
window.addEventListener('vite:preloadError', () => {
  const key = 'sm_reloaded_after_preload_error';
  if (sessionStorage.getItem(key)) return;
  sessionStorage.setItem(key, '1');
  window.location.reload();
});

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 1,
    },
  },
});

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <AppRouter />
      </BrowserRouter>
    </QueryClientProvider>
  </React.StrictMode>,
);
