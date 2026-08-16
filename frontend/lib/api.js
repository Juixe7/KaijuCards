// frontend/lib/api.js
import { getApiUrl } from './config';

async function request(path, options = {}) {
  const apiUrl = getApiUrl();
  const token = typeof window !== 'undefined' ? localStorage.getItem('meta_clash_token') : null;

  const headers = {
    'Content-Type': 'application/json',
    ...options.headers,
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const res = await fetch(`${apiUrl}${path}`, {
    ...options,
    headers,
  });

  if (res.status === 401 && typeof window !== 'undefined') {
    localStorage.removeItem('meta_clash_token');
    localStorage.removeItem('meta_clash_user_id');
    localStorage.removeItem('lastPlayerName');
    window.location.href = '/login';
    throw new Error('Session expired');
  }

  // Handle case where body might be empty or not JSON
  let data = {};
  const contentType = res.headers.get('content-type');
  if (contentType && contentType.includes('application/json')) {
    data = await res.json();
  }

  if (!res.ok) {
    throw new Error(data.error || 'API Request failed');
  }

  return data;
}

export const api = {
  get: (path, options) => request(path, { ...options, method: 'GET' }),
  post: (path, body, options) => request(path, { ...options, method: 'POST', body: JSON.stringify(body) }),
  put: (path, body, options) => request(path, { ...options, method: 'PUT', body: JSON.stringify(body) }),
  delete: (path, options) => request(path, { ...options, method: 'DELETE' }),
};
