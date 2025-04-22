// /m/frontend/src/api/index.js

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
  //baseURL: process.env.REACT_APP_BACKEND_URL || 'http://localhost:8080',
});

// Перехватчик запросов – добавляем заголовок авторизации (если токен сохранен)
api.interceptors.request.use(
  (config) => {
    const token = getAccessToken();
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Перехватчик ответов — обрабатываем ошибки и обновляем токен при необходимости
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;
    // Проверяем, получили ли мы статус 401 и не пытались ли уже обновить токен
    if (error.response && error.response.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      try {
         // делаем refresh тем же экземпляром api — и с относительным URL
       const refreshToken = getRefreshToken();
       const response = await api.post('/refresh', { refreshToken });
        const { accessToken, refreshToken: newRefreshToken } = response.data;
        // Сохраняем новые токены
        setAccessToken(accessToken);
        setRefreshToken(newRefreshToken);
        // Обновляем заголовок авторизации для исходного запроса и повторяем его
        originalRequest.headers['Authorization'] = `Bearer ${accessToken}`;
        return api(originalRequest);
      } catch (err) {
        // Если обновление не удалось, очищаем сохранённые токены и перенаправляем пользователя на страницу логина
        clearTokens();
       // window.location.href = '/login';
        return Promise.reject(err);
      }
    }
    return Promise.reject(error);
  }
);

export default api;
