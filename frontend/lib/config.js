// frontend/lib/config.js

export const getApiUrl = () => {
  if (process.env.NEXT_PUBLIC_API_URL) {
    return process.env.NEXT_PUBLIC_API_URL.replace(/\/+$/, '');
  }
  return 'http://localhost:8080';
};

export const getWsUrl = () => {
  if (process.env.NEXT_PUBLIC_WS_URL) {
    let wsUrl = process.env.NEXT_PUBLIC_WS_URL.replace(/\/+$/, '');
    // If the URL is provided without an explicit socket path, append /api/ws
    if (!wsUrl.endsWith('/api/ws') && !wsUrl.endsWith('/ws')) {
      wsUrl += '/api/ws';
    }
    return wsUrl;
  }
  const apiUrl = getApiUrl().replace(/\/+$/, '');
  if (apiUrl.startsWith('https://')) {
    return apiUrl.replace('https://', 'wss://') + '/api/ws';
  } else if (apiUrl.startsWith('http://')) {
    return apiUrl.replace('http://', 'ws://') + '/api/ws';
  }
  return 'ws://localhost:8080/api/ws';
};
