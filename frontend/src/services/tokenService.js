// m/frontend/src/services/tokenService.js

export const getAccessToken = () => {
    return localStorage.getItem('accessToken');
  };
  
  export const setAccessToken = (token) => {
    localStorage.setItem('accessToken', token);
  };
  
  // Аналогично для refreshToken
  export const getRefreshToken = () => {
    return localStorage.getItem('refreshToken');
  };
  
  export const setRefreshToken = (token) => {
    localStorage.setItem('refreshToken', token);
  };
  
  export const clearTokens = () => {
    localStorage.removeItem('accessToken');
    localStorage.removeItem('refreshToken');
  };