// m/frontend/src/contexts/AuthContext.jsx
import React, { createContext, useReducer, useContext, useEffect } from 'react';
import { setAccessToken, setRefreshToken, clearTokens } from '../services/tokenService';
import api from '../api/index';

const AuthStateContext = createContext();
const AuthDispatchContext = createContext();

const initialState = {
  user: null, // Объект пользователя (например, id, имя, email, фото)
  accessToken: localStorage.getItem('accessToken') || null,
  refreshToken: localStorage.getItem('refreshToken') || null,
};

function authReducer(state, action) {
  switch (action.type) {
    case 'LOGIN_SUCCESS':
      setAccessToken(action.payload.accessToken);
      setRefreshToken(action.payload.refreshToken);
      return {
        ...state,
        user: action.payload.user,
        accessToken: action.payload.accessToken,
        refreshToken: action.payload.refreshToken,
      };
    case 'TOKEN_REFRESHED':
      setAccessToken(action.payload.accessToken);
      setRefreshToken(action.payload.refreshToken);
      return {
        ...state,
        accessToken: action.payload.accessToken,
        refreshToken: action.payload.refreshToken,
      };
    case 'LOGOUT':
      clearTokens();
      return {
        ...state,
        user: null,
        accessToken: null,
        refreshToken: null,
      };
    case 'SET_USER':
      return {
        ...state,
        user: action.payload,
      };
    default:
      return state;
  }
}

export const AuthProvider = ({ children }) => {
  const [state, dispatch] = useReducer(authReducer, initialState);

  // При монтировании, если у нас есть токен — забираем профиль
  useEffect(() => {
    if (!state.accessToken) {
      return;
    }

    api.get('/me')
      .then(({ data }) => {
        dispatch({ type: 'SET_USER', payload: data });
      })
      .catch((err) => {
        // Очищаем токены только если это не ошибка 401
        // (401 обрабатывается в interceptor'е)
        if (err.response?.status !== 401) {
          clearTokens();
          dispatch({ type: 'LOGOUT' });
        }
      });
  }, [state.accessToken]);

  // Добавляем обработчик для обновления токена
  useEffect(() => {
    const handleTokenRefresh = (event) => {
      const { accessToken, refreshToken } = event.detail;
      dispatch({ type: 'TOKEN_REFRESHED', payload: { accessToken, refreshToken } });
    };

    window.addEventListener('tokenRefreshed', handleTokenRefresh);
    return () => window.removeEventListener('tokenRefreshed', handleTokenRefresh);
  }, []);

  return (
    <AuthStateContext.Provider value={state}>
      <AuthDispatchContext.Provider value={dispatch}>
         {children}
      </AuthDispatchContext.Provider>
    </AuthStateContext.Provider>
  );
};

export const useAuthState = () => useContext(AuthStateContext);
export const useAuthDispatch = () => useContext(AuthDispatchContext);
