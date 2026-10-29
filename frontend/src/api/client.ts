import axios, { type AxiosInstance, type AxiosRequestConfig } from 'axios';

const BASE_URL = (import.meta.env.VITE_API_URL as string | undefined) || '/api/v1';

const apiClient: AxiosInstance = axios.create({
  baseURL: BASE_URL,
  headers: { 'Content-Type': 'application/json' },
});

// ── Kiosk link token (K-016) ─────────────────────────────────────────────────
// A kiosk device is unlocked by a link with ?token=…; the token is kept in
// this device's storage and sent with every kiosk request instead of a JWT.
const KIOSK_TOKEN_KEY = 'sm_kiosk_token';

export function getKioskToken(): string | null {
  try {
    return localStorage.getItem(KIOSK_TOKEN_KEY);
  } catch {
    return null;
  }
}

export function setKioskToken(token: string): void {
  try {
    localStorage.setItem(KIOSK_TOKEN_KEY, token);
  } catch {
    // Private mode etc.: the kiosk then needs the link again on reload.
  }
}

const isKioskRequest = (url?: string) => !!url && url.startsWith('/kiosk/');

// Attach JWT from localStorage on every request (kiosk requests carry the
// kiosk link token instead).
apiClient.interceptors.request.use((config) => {
  config.headers = config.headers ?? {};
  if (isKioskRequest(config.url)) {
    const kiosk = getKioskToken();
    if (kiosk) config.headers['X-Kiosk-Token'] = kiosk;
    return config;
  }
  const token = localStorage.getItem('sm_token');
  if (token) {
    config.headers['Authorization'] = `Bearer ${token}`;
  }
  return config;
});

// On 401, clear token and redirect to login — except for the kiosk, which
// shows its own "not unlocked" screen.
apiClient.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err?.response?.status === 401 && !isKioskRequest(err?.config?.url)) {
      localStorage.removeItem('sm_token');
      if (!window.location.pathname.startsWith('/auth')) {
        window.location.href = '/auth/login';
      }
    }
    return Promise.reject(err);
  },
);

export { apiClient };

export async function apiGet<T>(path: string, config?: AxiosRequestConfig): Promise<T> {
  const res = await apiClient.get<T>(path, config);
  return res.data;
}

export async function apiPost<T>(path: string, body?: unknown): Promise<T> {
  const res = await apiClient.post<T>(path, body);
  return res.data;
}

export async function apiPut<T>(path: string, body?: unknown): Promise<T> {
  const res = await apiClient.put<T>(path, body);
  return res.data;
}

export async function apiPatch<T>(path: string, body?: unknown): Promise<T> {
  const res = await apiClient.patch<T>(path, body);
  return res.data;
}

export async function apiDelete<T>(path: string): Promise<T> {
  const res = await apiClient.delete<T>(path);
  return res.data;
}
