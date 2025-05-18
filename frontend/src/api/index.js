import axios from 'axios';
import {
  getAccessToken,
  setAccessToken,
  getRefreshToken,
  setRefreshToken,
  clearTokens,
} from '../services/tokenService';

// Создаем инстанс axios с базовыми настройками
const api = axios.create({
  baseURL: '/', 
});

// Перехватчик запросов – добавляем заголовок авторизации (если токен сохранен)
api.interceptors.request.use(
  config => {
    const token = getAccessToken();
    console.log('🔑 Request interceptor - Access token:', token ? 'present' : 'missing');
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`;
    }
    return config;
  },
  error => Promise.reject(error)
);

// Перехватчик ответов — обрабатываем 401 и пытаемся обновить токен
api.interceptors.response.use(
  response => response,
  async error => {
    const originalRequest = error.config;
    console.log('🔑 Response interceptor - Status:', error.response?.status);
    console.log('🔑 Response interceptor - URL:', originalRequest.url);

    // Если получили 401, ещё не пробовали обновлять, и это не сам /refresh
    if (
      error.response?.status === 401 &&
      !originalRequest._retry &&
      !originalRequest.url.includes('/refresh')
    ) {
      console.log('🔑 Attempting token refresh...');
      originalRequest._retry = true;

      const refreshToken = getRefreshToken();
      console.log('🔑 Refresh token:', refreshToken ? 'present' : 'missing');
      
      if (refreshToken) {
        try {
          // Делаем refresh
          const { data } = await api.post('/refresh', { refreshToken });
          const { accessToken: newAccessToken, refreshToken: newRefreshToken } = data;
          console.log('🔑 New tokens received');

          // Сохраняем новые токены
          setAccessToken(newAccessToken);
          setRefreshToken(newRefreshToken);

          // Отправляем событие об обновлении токена
          window.dispatchEvent(new CustomEvent('tokenRefreshed', {
            detail: {
              accessToken: newAccessToken,
              refreshToken: newRefreshToken
            }
          }));

          // Повторяем исходный запрос с обновлённым accessToken
          originalRequest.headers['Authorization'] = `Bearer ${newAccessToken}`;
          return api(originalRequest);
        } catch (refreshError) {
          console.error('🔑 Token refresh failed:', refreshError);
          // Если обновление не удалось — чистим и пробрасываем ошибку
          clearTokens();
          return Promise.reject(refreshError);
        }
      } else {
        console.log('🔑 No refresh token available');
        // Нет refreshToken — сразу очищаем и пробрасываем
        clearTokens();
      }
    }

    return Promise.reject(error);
  }
);

export default api;
