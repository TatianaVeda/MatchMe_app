// /m/frontend/src/api/recommendations.js

import api from './index';

/**
 * Запрос рекомендаций.
 * @param {Object} options
 * @param {'affinity'|'desire'} [options.mode='affinity']
 * @param {boolean} [options.withDistance=false]
 * @param {Object} [options.params={}] — дополнительные параметры фильтрации
 * @returns {Promise<Array<{id: string, distance?: number, score?: number}>>}
 */
export const getRecommendations = async ({
  mode = 'affinity',
  withDistance = false,
  params = {}
} = {}) => {
  const response = await api.get('/recommendations', {
    params: { mode, withDistance, ...params },
  });
  return response.data;
};

/**
 * Отклонить рекомендацию.
 * @param {string} id — UUID рекомендации
 */
export const declineRecommendation = async (id) => {
  return api.post(`/recommendations/${id}/decline`);
};
