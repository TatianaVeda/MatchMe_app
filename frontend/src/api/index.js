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

    // Если получили 401, ещё не пробовали обновлять, и это не сам /refresh
    if (
      error.response?.status === 401 &&
      !originalRequest._retry &&
      !originalRequest.url.includes('/refresh')
    ) {
      originalRequest._retry = true;

      const refreshToken = getRefreshToken();
      if (refreshToken) {
        try {
          // Делаем refresh
          const { data } = await api.post('/refresh', { refreshToken });
          const { accessToken: newAccessToken, refreshToken: newRefreshToken } = data;

          // Сохраняем новые токены
          setAccessToken(newAccessToken);
          setRefreshToken(newRefreshToken);

          // Повторяем исходный запрос с обновлённым accessToken
          originalRequest.headers['Authorization'] = `Bearer ${newAccessToken}`;
          return api(originalRequest);
        } catch (refreshError) {
          // Если обновление не удалось — чистим и пробрасываем ошибку
          clearTokens();
          return Promise.reject(refreshError);
        }
      } else {
        // Нет refreshToken — сразу очищаем и пробрасываем
        clearTokens();
      }
    }

    return Promise.reject(error);
  }
);

export default api;
