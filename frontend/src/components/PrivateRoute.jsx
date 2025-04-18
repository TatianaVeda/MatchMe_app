// m/frontend/src/components/PrivateRoute.jsx
import React from 'react';
import { Navigate } from 'react-router-dom';
import { useAuthState } from '../contexts/AuthContext';

const PrivateRoute = ({ children }) => {
  const { accessToken } = useAuthState();
  if (!accessToken) {
    return <Navigate to="/login" replace />;
  }
   // Если токен есть, отрисовываем дочерние компоненты
   return children;
};

export default PrivateRoute;
