// Re-export the axios instance from client.ts so courseApi.ts can use it.
// This avoids circular imports while sharing the auth/tenant interceptor setup.

import axios from 'axios';

const api = axios.create({
  baseURL: '/api',
  headers: { 'Content-Type': 'application/json' },
});

// Auth header management (mirrors client.ts)
export function setAuthToken(token: string | null) {
  if (token) {
    api.defaults.headers.common['Authorization'] = `Bearer ${token}`;
  } else {
    delete api.defaults.headers.common['Authorization'];
  }
}

export function setTenantHeader(tenantId: string | null) {
  if (tenantId) {
    api.defaults.headers.common['X-Tenant-ID'] = tenantId;
  } else {
    delete api.defaults.headers.common['X-Tenant-ID'];
  }
}

// Sync tokens from localStorage on module load
const accessToken = localStorage.getItem('mycourses_access_token') || localStorage.getItem('lastsaas_access_token');
if (accessToken) setAuthToken(accessToken);

// Silent token refresh on 401
let isRefreshing = false;
let refreshSubscribers: ((token: string) => void)[] = [];

api.interceptors.response.use(
  (res) => res,
  async (error) => {
    const originalRequest = error.config;
    if (error.response?.status !== 401 || originalRequest?._retry || originalRequest?.url?.includes('/auth/')) {
      return Promise.reject(error);
    }

    const refreshToken = localStorage.getItem('mycourses_refresh_token') || localStorage.getItem('lastsaas_refresh_token');
    if (!refreshToken) {
      window.location.href = '/login';
      return Promise.reject(error);
    }

    if (isRefreshing) {
      return new Promise((resolve) => {
        refreshSubscribers.push((token: string) => {
          originalRequest.headers['Authorization'] = `Bearer ${token}`;
          resolve(api(originalRequest));
        });
      });
    }

    isRefreshing = true;
    originalRequest._retry = true;

    try {
      const { data } = await axios.post('/api/auth/refresh', { refreshToken });
      localStorage.setItem('mycourses_access_token', data.accessToken);
      localStorage.setItem('mycourses_refresh_token', data.refreshToken);
      setAuthToken(data.accessToken);
      refreshSubscribers.forEach(cb => cb(data.accessToken));
      refreshSubscribers = [];
      originalRequest.headers['Authorization'] = `Bearer ${data.accessToken}`;
      return api(originalRequest);
    } catch {
      localStorage.removeItem('mycourses_access_token');
      localStorage.removeItem('mycourses_refresh_token');
      window.location.href = '/login';
      return Promise.reject(error);
    } finally {
      isRefreshing = false;
    }
  }
);

export default api;
