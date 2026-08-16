// frontend/lib/config.js

export const getApiUrl = () => {
  if (process.env.NODE_ENV === 'development') {
    return 'http://localhost:8080';
  }
  return 'https://meta-clash-backend-production-7dc7.up.railway.app';
};

export const getWsUrl = () => {
  
  const apiUrl = getApiUrl().replace(/\/+$/, '');
  if (apiUrl.startsWith('https://')) {
    return apiUrl.replace('https://', 'wss://') + '/api/ws';
  } else if (apiUrl.startsWith('http://')) {
    return apiUrl.replace('http://', 'ws://') + '/api/ws';
  }
  
  return 'ws://localhost:8080/api/ws';
};
