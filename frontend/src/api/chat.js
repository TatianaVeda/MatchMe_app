// /m/frontend/src/api/chat.js

import api from './index';

/**
 * chat.js
 *
 * API functions for chat operations (fetching, sending messages).
 * Wraps HTTP requests to backend endpoints, handles errors.
 */

export const getChats = async () => {
  const response = await api.get('/chats');
  return response.data;
};

export const getChatHistory = async (chatId, page = 1, limit = 20) => {
  const response = await api.get(`/chats/${chatId}`, {
    params: { page, limit },
  });
  return response.data;
};

export const sendMessage = async (chatId, content) => {
  const response = await api.post(`/chats/${chatId}/messages`, { content });
  return response.data;
};
