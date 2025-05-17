import api from './index';
export const getUser = async (userId) => {
  try {
    const response = await api.get(`/users/${userId}`);
    return response.data;
  } catch (error) {
    if (error.response && (error.response.status === 404 || error.response.status === 403)) {
      return null;
    }
    throw new Error("Failed to fetch user");
  }
};
export const getMyData = async () => {
  const response = await api.get('/me');
  return response.data;
};
export const getMyProfile = async () => {
  const response = await api.get('/me/profile');
  return response.data;
};
export const getMyBio = async () => {
  const response = await api.get('/me/bio');
  return response.data;
};
export const getUserProfile = async (userId) => {
  const response = await api.get(`/users/${userId}/profile`);
  return response.data;
};
export const getUserBio = async (userId) => {
  const response = await api.get(`/users/${userId}/bio`);
  return response.data;
};
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
    export const getMyPreferences = async () => {
    const response = await api.get('/me/preferences');
    return response.data;
  };
      export const deleteMyPhoto = async () => {
      const response = await api.delete('/me/photo');
      return response.data;
    };
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