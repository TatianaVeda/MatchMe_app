// m/frontend/src/contexts/AuthContext.jsx
import React, { createContext, useReducer, useContext, useEffect } from 'react';
import { setAccessToken, setRefreshToken, clearTokens } from '../services/tokenService';
import api from '../api/index';

const AuthStateContext = createContext();
const AuthDispatchContext = createContext();

const initialState = {
  user: null, 
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

    useEffect(() => {
      /**
       * On mount or when accessToken changes: restores user session if token exists.
       * Fetches user info from backend, handles token expiration and logout.
       */
          if (!state.accessToken) {
            return;
          }
      
          api.get('/me')
            .then(({ data }) => {
              dispatch({ type: 'SET_USER', payload: data });
            })
            .catch((err) => {
              clearTokens();
              dispatch({ type: 'LOGOUT' });
            });
        }, [state.accessToken]);

  return (
    <AuthStateContext.Provider value={state}>
      <AuthDispatchContext.Provider value={dispatch}>
         {children}
      </AuthDispatchContext.Provider>
    </AuthStateContext.Provider>
  );
};

/**
 * useAuthState
 * Returns current authentication state (user, tokens).
 */
export const useAuthState = () => useContext(AuthStateContext);
/**
 * useAuthDispatch
 * Returns dispatch function for updating authentication state.
 */
export const useAuthDispatch = () => useContext(AuthDispatchContext);
