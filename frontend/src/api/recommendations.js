// /m/frontend/src/api/recommendations.js

import api from './index';

// Получить рекомендации (до 10 рекомендованных идентификаторов)
export const getRecommendations = async () => {
  const response = await api.get('/recommendations');
  return response.data;
};
