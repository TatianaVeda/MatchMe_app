// /m/frontend/src/api/connections.js

import api from './index';

export const getConnections = async () => {
  const response = await api.get('/connections');
  return response.data;
};

export const getPendingConnections = async () => {
  const response = await api.get('/connections/pending');
  return response.data;
};

export const getSentConnections = async () => {
  const response = await api.get('/connections/sent');
  return response.data;
};

export const sendConnectionRequest = async (targetUserId) => {
  const response = await api.post(`/connections/${targetUserId}`);
  return response.data;
};

export const updateConnectionRequest = async (senderUserId, action) => {
  const response = await api.put(`/connections/${senderUserId}`, { action });
  return response.data;
};

export const deleteConnection = async (targetUserId) => {
  const response = await api.delete(`/connections/${targetUserId}`);
  return response.data;
};
