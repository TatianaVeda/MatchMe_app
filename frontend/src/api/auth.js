// /m/frontend/src/api/auth.js

import api from './index';

export const login = async (credentials) => {
  const response = await api.post('/login', credentials);
  return response.data;
};

export const signup = async (data) => {
  const response = await api.post('/signup', data);
  return response.data;
};

export const logout = async () => {
  const response = await api.post('/logout');
  return response.data;
};

export const refreshToken = async (currentRefreshToken) => {
  const response = await api.post('/refresh', { refreshToken: currentRefreshToken });
  return response.data;
};
