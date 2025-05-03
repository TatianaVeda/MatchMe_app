// // m/frontend/src/App.jsx

// import React from 'react';
// import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
// import { AuthProvider } from './contexts/AuthContext';
// import { ChatProvider } from './contexts/ChatContext';
// import PrivateRoute from './components/PrivateRoute';
// import Login from './pages/Auth/Login';
// import Signup from './pages/Auth/Signup';
// import MyProfile from './pages/Profile/MyProfile';
// import EditProfile from './pages/Profile/EditProfile';
// import Recommendations from './pages/Recommendations';
// import Connections from './pages/Connections';
// import Chats from './pages/Chats';
// import ChatWindow from './pages/ChatWindow';
// import Settings from './pages/Settings';
// import Header from './components/Header';
// import { ThemeProvider } from '@mui/material/styles'
// import theme from './theme'
// import { ToastContainer } from 'react-toastify';
// import 'react-toastify/dist/ReactToastify.css';
// import Friends from './pages/Friends';
// import UserProfile from './pages/Profile/UserProfile';
// import { useAuthState } from './contexts/AuthContext';
// import { ADMIN_ID } from './config';
// import AdminPanel from './pages/Profile/AdminPanel';

// function App() {
//   const { user } = useAuthState();
//   return (
//     <ThemeProvider theme={theme}>
//     <AuthProvider>
//       <ChatProvider>
//         <Router>
//           <Header />
//           <Routes>
//           <Route path="/" element={<Navigate to="/login" replace />} />
//             {/* Публичные маршруты */}
//             <Route path="/login" element={<Login />} />
//             <Route path="/signup" element={<Signup />} />

//             {/* Защищённые маршруты */}
//             <Route path="/me" element={<PrivateRoute><MyProfile /></PrivateRoute>} />
//             <Route path="/edit-profile" element={<PrivateRoute><EditProfile /></PrivateRoute>} />
//             <Route path="/users/:id" element={<PrivateRoute><UserProfile /></PrivateRoute>} />
//             <Route path="/recommendations" element={<PrivateRoute><Recommendations /></PrivateRoute>} />
//             <Route path="/connections" element={<PrivateRoute><Connections /></PrivateRoute>} />
//             <Route path="/chats" element={<PrivateRoute><Chats /></PrivateRoute>} />
//             <Route path="/chat/:chatId" element={<PrivateRoute><ChatWindow /></PrivateRoute>} />
//             <Route path="/settings" element={<PrivateRoute><Settings /></PrivateRoute>} />
//             <Route path="/friends"   element={<PrivateRoute><Friends  /></PrivateRoute>} />
//             <Route path="*" element={<Navigate to="/login" replace />} />
//             <Route path="/admin" element={<PrivateRoute> {user?.id === ADMIN_ID
//                       ? <AdminPanel />
//                       : <Navigate to="/me" replace />}
//                   </PrivateRoute>
//                 }
//               />
//           </Routes>
//           <ToastContainer position="top-right" autoClose={2000} />
//         </Router>
//       </ChatProvider>
//     </AuthProvider>
//     </ThemeProvider>
//   );
// }

// export default  App;


import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { ThemeProvider } from '@mui/material/styles';
import { ToastContainer } from 'react-toastify';
import theme from './theme';

import { AuthProvider, useAuthState } from './contexts/AuthContext';
import { ChatProvider } from './contexts/ChatContext';

import PrivateRoute from './components/PrivateRoute';
import Header from './components/Header';

import Login from './pages/Auth/Login';
import Signup from './pages/Auth/Signup';
import MyProfile from './pages/Profile/MyProfile';
import EditProfile from './pages/Profile/EditProfile';
import UserProfile from './pages/Profile/UserProfile';
import AdminPanel from './pages/Profile/AdminPanel';
import Recommendations from './pages/Recommendations';
import Connections from './pages/Connections';
import Chats from './pages/Chats';
import ChatWindow from './pages/ChatWindow';
import Settings from './pages/Settings';
import Friends from './pages/Friends';

import { ADMIN_ID } from './config';
import 'react-toastify/dist/ReactToastify.css';

function App() {
  return (
    <ThemeProvider theme={theme}>
      <AuthProvider>
        <ChatProvider>
          <Router>
            <Header />
            <AppRoutes />
            <ToastContainer position="top-right" autoClose={2000} />
          </Router>
        </ChatProvider>
      </AuthProvider>
    </ThemeProvider>
  );
}

// Внутренний компонент с доступом к useAuthState внутри провайдера
function AppRoutes() {
  const { user } = useAuthState();

  return (
    <Routes>
      <Route path="/" element={<Navigate to="/login" replace />} />
      {/* Публичные маршруты */}
      <Route path="/login" element={<Login />} />
      <Route path="/signup" element={<Signup />} />

      {/* Защищённые маршруты */}
      <Route path="/me" element={<PrivateRoute><MyProfile /></PrivateRoute>} />
      <Route path="/edit-profile" element={<PrivateRoute><EditProfile /></PrivateRoute>} />
      <Route path="/users/:id" element={<PrivateRoute><UserProfile /></PrivateRoute>} />
      <Route path="/recommendations" element={<PrivateRoute><Recommendations /></PrivateRoute>} />
      <Route path="/connections" element={<PrivateRoute><Connections /></PrivateRoute>} />
      <Route path="/chats" element={<PrivateRoute><Chats /></PrivateRoute>} />
      <Route path="/chat/:chatId" element={<PrivateRoute><ChatWindow /></PrivateRoute>} />
      <Route path="/settings" element={<PrivateRoute><Settings /></PrivateRoute>} />
      <Route path="/friends" element={<PrivateRoute><Friends /></PrivateRoute>} />
      <Route
        path="/admin"
        element={
          <PrivateRoute>
            {user?.id === ADMIN_ID ? <AdminPanel /> : <Navigate to="/me" replace />}
          </PrivateRoute>
        }
      />
      <Route path="*" element={<Navigate to="/login" replace />} />
    </Routes>
  );
}

export default App;
