// // /m/frontend/src/api/user.js
import api from './index';

// Получить базовые данные пользователя по id
// export const getUser = async (userId) => {
//   const response = await api.get(`/users/${userId}`);
//   return response.data;
// };

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

// Получить сокращённые данные аутентифицированного пользователя (/me)
export const getMyData = async () => {
  const response = await api.get('/me');
  return response.data;
};

// Получить полный профиль аутентифицированного пользователя (/me/profile)
export const getMyProfile = async () => {
  const response = await api.get('/me/profile');
  return response.data;
};

// Получить биографию аутентифицированного пользователя (/me/bio)
export const getMyBio = async () => {
  const response = await api.get('/me/bio');
  return response.data;
};

// Получить профиль другого пользователя (/users/{id}/profile)
export const getUserProfile = async (userId) => {
  const response = await api.get(`/users/${userId}/profile`);
  return response.data;
};

// Получить биографию другого пользователя (/users/{id}/bio)
export const getUserBio = async (userId) => {
  const response = await api.get(`/users/${userId}/bio`);
  return response.data;
};

// ------------------------------------------------------------
// Обновление данных текущего пользователя
// ------------------------------------------------------------

// Обновить профиль аутентифицированного пользователя (/me/profile)
// Передаём: { firstName, lastName, about, city }
// export const updateMyProfile = async ({ firstName, lastName, about, city, latitude, longitude }) => {
//     const payload = { firstName, lastName, about, city };
//     // добавляем координаты, если они переданы
//     if (latitude != null && longitude != null) {
//       payload.latitude = latitude;
//       payload.longitude = longitude;
//     }
//     const response = await api.put('/me/profile', payload);
//     return response.data;
//   };

export const updateMyProfile = async ({ firstName, lastName, about, city, latitude, longitude }) => {
  const payload = { firstName, lastName, about, city: city.name || city};

  // Add coordinates if provided
  if (latitude != null && longitude != null) {
    payload.latitude = latitude;
    payload.longitude = longitude;

    // Adding earth_loc field to be updated on the backend
    payload.earth_loc = `ll_to_earth(${latitude}, ${longitude})`;
  }

  const response = await api.put('/me/profile', payload);
  return response.data;
};


// Обновить биографию аутентифицированного пользователя (/me/bio)
// Передаём: { interests, hobbies, music, food, travel, lookingFor,
//             priorityInterests, priorityHobbies, priorityMusic,
//             priorityFood, priorityTravel }
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

// Обновить предпочтения пользователя (/me/preferences)
// Передаём: { maxRadius, priorityInterests, priorityHobbies, ... }
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
