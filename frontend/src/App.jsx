// m/frontend/src/App.jsx
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

function AppRoutes() {
  const { user } = useAuthState();

  return (
    <Routes>
      <Route
  path="/"
  element={
    user?.id
      ? <Navigate to="/me" replace />
      : <Navigate to="/login" replace />
  }
/>
     
<Route
        path="/login"
        element={user ? <Navigate to="/me" replace /> : <Login />}
      />
      <Route
        path="/signup"
        element={user ? <Navigate to="/me" replace /> : <Signup />}
      />

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
