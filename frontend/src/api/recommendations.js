// /m/frontend/src/api/recommendations.js

import api from './index';

/**
 * Запрос рекомендаций.
 * @param {{mode: string, withDistance: boolean}} options
 * @returns {Promise<uuid.UUID[]>}
 */
export const getRecommendations = async (
  { mode = 'affinity', withDistance = false } = {}
) => {
  const response = await api.get('/recommendations', {
    params: { mode, withDistance },
  });
  return response.data;
};

export const declineRecommendation = async (id) => {
  // если у вас в бэке — POST /recommendations/:id/decline
  return api.post(`/recommendations/${id}/decline`);
};