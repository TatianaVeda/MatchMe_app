import api from './index';
/**
 * user.js
 *
 * API functions for user data: profile, bio, photo, and location.
 * Wraps HTTP requests to backend endpoints, handles errors.
 */
/**
 * Fetches user data by userId. Returns null if user not found or forbidden.
 * @param {string} userId
 * @returns {Promise<Object|null>}
 */
export const getUser = async (userId) => {
  try {
    const response = await api.get(`/users/${userId}`);

    return { id: userId, ...response.data };
  } catch (error) {
    if (error.response && (error.response.status === 404 || error.response.status === 403)) {
      return null;
    }
    throw new Error("Failed to fetch user");
  }
};
/**
 * Fetches current authenticated user's data.
 * @returns {Promise<Object>}
 */
export const getMyData = async () => {
  const response = await api.get('/me');
  return response.data;
};
/**
 * Fetches current user's profile.
 * @returns {Promise<Object>}
 */
export const getMyProfile = async () => {
  const response = await api.get('/me/profile');
  return response.data;
};
export const getMyBio = async () => {
  const response = await api.get('/me/bio');
  return response.data;
};
/**
 * Fetches profile of another user by userId.
 * @param {string} userId
 * @returns {Promise<Object>}
 */
export const getUserProfile = async (userId) => {
  const response = await api.get(`/users/${userId}/profile`);
  return response.data;
};
/**
 * Fetches bio of another user by userId.
 * @param {string} userId
 * @returns {Promise<Object>}
 */
export const getUserBio = async (userId) => {
  const response = await api.get(`/users/${userId}/bio`);
  return response.data;
};
/**
 * Updates current user's profile (name, about, city, location).
 * @param {Object} params
 * @returns {Promise<Object>}
 */
export const updateMyProfile = async ({ firstName, lastName, about, city, latitude, longitude }) => {
  const payload = { firstName, lastName, about, city };
  if (latitude != null && longitude != null) {
    payload.latitude = latitude;
    payload.longitude = longitude;
    payload.earth_loc = `ll_to_earth(${latitude}, ${longitude})`;
  }
  const response = await api.put('/me/profile', payload);
  return response.data;
};
/**
 * Updates current user's bio (interests, hobbies, etc).
 * @param {Object} params
 * @returns {Promise<Object>}
 */
export const updateMyBio = async ({
  interests,
  hobbies,
  music,
  food,
  travel,
  lookingFor,
  priorityInterests,
  priorityHobbies,
  priorityMusic,
  priorityFood,
  priorityTravel
}) => {
  const response = await api.put('/me/bio', {
    interests,
    hobbies,
    music,
    food,
    travel,
    lookingFor,
    priorityInterests,
    priorityHobbies,
    priorityMusic,
    priorityFood,
    priorityTravel
  });
  return response.data;
};
/**
 * Fetches current user's preferences.
 * @returns {Promise<Object>}
 */
export const getMyPreferences = async () => {
  const response = await api.get('/me/preferences');
  return response.data;
};
/**
 * Deletes current user's photo.
 * @returns {Promise<Object>}
 */
export const deleteMyPhoto = async () => {
  const response = await api.delete('/me/photo');
  return response.data;
};
/**
 * Updates current user's preferences (search radius, priorities).
 * @param {Object} params
 * @returns {Promise<Object>}
 */
export const updateMyPreferences = async ({
  maxRadius,
  priorityInterests,
  priorityHobbies,
  priorityMusic,
  priorityFood,
  priorityTravel
}) => {
  const response = await api.put('/me/preferences', {
    maxRadius,
    priorityInterests,
    priorityHobbies,
    priorityMusic,
    priorityFood,
    priorityTravel
  });
  return response.data;
};
/**
 * Batch fetches online status for a list of user IDs.
 * @param {string[]} ids
 * @returns {Promise<Object>} Map userId -> online status
 */
export const getBatchOnlineStatus = async (ids) => {
  if (!ids.length) return {};
  const { data } = await api.get('/api/user/online/batch', {
    params: { ids: ids.join(',') }
  });
  return data; 
};
