import { useNavigate, useLocation } from 'react-router-dom';

// Closes a route-backed modal/screen: goes back one entry in the browser's
// history when this session navigated here itself (so the back button and
// an explicit close button behave the same way), or replaces the URL with
// `fallback` when there is nothing to go back to in-app — e.g. the user
// reloaded the page, or followed a direct link straight to this route.
// React Router gives the initial entry of a fresh page load the key
// "default"; any entry pushed by in-app navigation gets a real key.
export function useSmartBack(fallback: string): () => void {
  const navigate = useNavigate();
  const location = useLocation();
  return () => {
    if (location.key === 'default') navigate(fallback, { replace: true });
    else navigate(-1);
  };
}
