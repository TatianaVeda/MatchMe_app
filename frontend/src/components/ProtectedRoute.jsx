// // frontend/src/components/ProtectedRoute.jsx
// import React, { useContext } from 'react';
// import { AuthContext } from '../contexts/AuthContext';
// import { Navigate } from 'react-router-dom';

// const ProtectedRoute = ({ children }) => {
//   const { authData } = useContext(AuthContext);
//   if (!authData) {
//     return <Navigate to="/login" replace />;
//   }
//   return children;
// };

// export default ProtectedRoute;
