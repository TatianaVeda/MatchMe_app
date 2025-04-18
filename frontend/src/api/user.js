// /m/frontend/src/api/user.js

import api from './index';

// Получить базовые данные пользователя по id
export const getUser = async (userId) => {
  const response = await api.get(`/users/${userId}`);
  return response.data;
};

// Получить данные профиля аутентифицированного пользователя (короткий вариант /me)
export const getMyData = async () => {
  const response = await api.get('/me');
  return response.data;
};

// Получить полный профиль аутентифицированного пользователя (например, /me/profile)
export const getMyProfile = async () => {
  const response = await api.get('/me/profile');
  return response.data;
};

// Получить биографию аутентифицированного пользователя (/me/bio)
export const getMyBio = async () => {
  const response = await api.get('/me/bio');
  return response.data;
};

// Получить профиль другого пользователя (/users/{id}/profile)
export const getUserProfile = async (userId) => {
  const response = await api.get(`/users/${userId}/profile`);
  return response.data;
};

// Получить биографию другого пользователя (/users/{id}/bio)
export const getUserBio = async (userId) => {
  const response = await api.get(`/users/${userId}/bio`);
  return response.data;
};

// Обновить профиль аутентифицированного пользователя (/me/profile)
export const updateMyProfile = async (profileData) => {
  const response = await api.put('/me/profile', profileData);
  return response.data;
};

// Обновить биографию аутентифицированного пользователя (/me/bio)
export const updateMyBio = async (bioData) => {
  const response = await api.put('/me/bio', bioData);
  return response.data;
};
