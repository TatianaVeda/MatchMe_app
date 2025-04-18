// // m/frontend/src/api/axiosInstance.js
// import axios from 'axios';
// import { getAccessToken, setAccessToken, getRefreshToken, setRefreshToken, clearTokens } from '../services/tokenService';

// const instance = axios.create({
//   baseURL: process.env.REACT_APP_BACKEND_URL || 'http://localhost:8080', // замените на URL вашего бекэнда
// });

// // Перехватчик запросов – добавляем токен
// instance.interceptors.request.use(
//   (config) => {
//     const token = getAccessToken();
//     if (token) {
//       config.headers['Authorization'] = `Bearer ${token}`;
//     }
//     return config;
//   },
//   (error) => Promise.reject(error)
// );

// // Перехватчик ответов – обновление токена при получении ошибки 401 (например, Token expired)
// instance.interceptors.response.use(
//   (response) => response,
//   async (error) => {
//     const originalRequest = error.config;
//     // Проверяем, получили ли мы 401 и не пытались ли уже обновить токен
//     if (error.response && error.response.status === 401 && !originalRequest._retry) {
//       originalRequest._retry = true;
//       try {
//         // Вызываем эндпоинт refresh для обновления токена
//         const refreshToken = getRefreshToken();
//         const response = await axios.post(
//           `${instance.defaults.baseURL}/refresh`,
//           { refreshToken } // В зависимости от API, возможно, нужно передавать токен в другом виде
//         );
//         const { accessToken, refreshToken: newRefreshToken } = response.data;
//         // Сохраняем новые токены
//         setAccessToken(accessToken);
//         setRefreshToken(newRefreshToken);
//         // Обновляем заголовок авторизации и повторяем исходный запрос
//         originalRequest.headers['Authorization'] = `Bearer ${accessToken}`;
//         return instance(originalRequest);
//       } catch (err) {
//         // Если обновление не удалось, чистим сохранённые токены и перенаправляем на страницу логина
//         clearTokens();
//         window.location.href = '/login';
//         return Promise.reject(err);
//       }
//     }
//     return Promise.reject(error);
//   }
// );

// export default instance;
