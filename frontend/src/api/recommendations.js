// /m/frontend/src/api/recommendations.js
import api from './index';

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
