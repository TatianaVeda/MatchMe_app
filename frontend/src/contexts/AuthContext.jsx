// m/frontend/src/contexts/AuthContext.jsx
import React, { createContext, useReducer, useContext } from 'react';
import { setAccessToken, setRefreshToken, clearTokens } from '../services/tokenService';

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
