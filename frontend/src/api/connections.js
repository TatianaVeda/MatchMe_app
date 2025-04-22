// /m/frontend/src/api/connections.js

import api from './index';

export const getConnections = async () => {
  const response = await api.get('/connections');
  return response.data;
};

// Отправка запроса на подключение к пользователю с указанным id
export const sendConnectionRequest = async (targetUserId) => {
  const response = await api.post(`/connections/${targetUserId}`);
  return response.data;
};

// Принятие или отклонение запроса на подключение
// В теле запроса передается объект: { action: "accept" } или { action: "decline" }
export const updateConnectionRequest = async (senderUserId, action) => {
  const response = await api.put(`/connections/${senderUserId}`, { action });
  return response.data;
};

// Удаление (разрыв) существующего подключения
export const deleteConnection = async (targetUserId) => {
  const response = await api.delete(`/connections/${targetUserId}`);
  return response.data;
};
