// m/frontend/src/services/tokenService.js

/**
 * tokenService.js
 *
 * Utility functions for managing JWT access and refresh tokens in localStorage.
 * Provides get/set/clear helpers for authentication flow.
 */

export const getAccessToken = () => {
    return localStorage.getItem('accessToken');
  };
  
  export const setAccessToken = (token) => {
    localStorage.setItem('accessToken', token);
  };
  
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