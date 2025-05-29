// /m/frontend/src/api/recommendations.js
import api from './index';

/**
 * recommendations.js
 *
 * API functions for fetching and declining user recommendations.
 * Wraps HTTP requests to backend endpoints, handles errors.
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

export const declineRecommendation = async (id) => {
  return api.post(`/recommendations/${id}/decline`);
};
