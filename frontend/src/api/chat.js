// /m/frontend/src/api/chat.js

import api from './index';

// Получить список чатов для текущего пользователя
export const getChats = async () => {
  const response = await api.get('/chats');
  return response.data;
};

// Получить историю сообщений для чата по его id с пагинацией
export const getChatHistory = async (chatId, page = 1, limit = 20) => {
  const response = await api.get(`/chats/${chatId}`, {
    params: { page, limit },
  });
  return response.data;
};

// Отправить новое сообщение в чат
export const sendMessage = async (chatId, content) => {
  const response = await api.post(`/chats/${chatId}/messages`, { content });
  return response.data;
};
