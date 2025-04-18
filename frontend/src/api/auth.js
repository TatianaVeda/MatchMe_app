// /m/frontend/src/api/auth.js

import api from './index';

// Функция для входа (login)
// credentials: { email, password }
export const login = async (credentials) => {
  const response = await api.post('/login', credentials);
  return response.data;
};

// Функция для регистрации (signup)
// data: { email, password }
export const signup = async (data) => {
  const response = await api.post('/signup', data);
  return response.data;
};

// Функция для выхода (logout)
export const logout = async () => {
  const response = await api.post('/logout');
  return response.data;
};

// (Опционально) Функция для обновления токена вручную, 
// хотя interceptor уже обрабатывает обновление при 401 ошибке
export const refreshToken = async (currentRefreshToken) => {
  const response = await api.post('/refresh', { refreshToken: currentRefreshToken });
  return response.data;
};
