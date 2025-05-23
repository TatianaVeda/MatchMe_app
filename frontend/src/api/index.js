import axios from 'axios';
import {
  getAccessToken,
  setAccessToken,
  getRefreshToken,
  setRefreshToken,
  clearTokens,
} from '../services/tokenService';

const api = axios.create({
  baseURL: '/', 
});

api.interceptors.request.use(
  config => {
    const token = getAccessToken();
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`;
    }
    return config;
  },
  error => Promise.reject(error)
);

api.interceptors.response.use(
  response => response,
  async error => {
    const originalRequest = error.config;

    if (
      error.response?.status === 401 &&
      !originalRequest._retry &&
      !originalRequest.url.includes('/refresh')
    ) {
      originalRequest._retry = true;

      const refreshToken = getRefreshToken();
      if (refreshToken) {
        try {
          const { data } = await api.post('/refresh', { refreshToken });
          const { accessToken: newAccessToken, refreshToken: newRefreshToken } = data;

          setAccessToken(newAccessToken);
          setRefreshToken(newRefreshToken);

          originalRequest.headers['Authorization'] = `Bearer ${newAccessToken}`;
          return api(originalRequest);
        } catch (refreshError) {
          clearTokens();
          return Promise.reject(refreshError);
        }
      } else {
        clearTokens();
      }
    }

    return Promise.reject(error);
  }
);

export default api;
