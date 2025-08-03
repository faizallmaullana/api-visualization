import axios from 'axios';

// Axios instance with base URL pointing to backend API
const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '',
  headers: {
    'Content-Type': 'application/json'
  }
});

// Composable for API calls
export function useApi() {
  return {
    getStats: () => apiClient.get('/api/stats'),
    getHistory: (params) => apiClient.get('/api/history', { params })
  };
}
