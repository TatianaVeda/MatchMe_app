// m/frontend/src/App.jsx

import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './contexts/AuthContext';
import { ChatProvider } from './contexts/ChatContext';
import PrivateRoute from './components/PrivateRoute';
import Login from './pages/Auth/Login';
import Signup from './pages/Auth/Signup';
import MyProfile from './pages/Profile/MyProfile';
import EditProfile from './pages/Profile/EditProfile';
import Recommendations from './pages/Recommendations';
import Connections from './pages/Connections';
import Chats from './pages/Chats';
import ChatWindow from './pages/ChatWindow';
import Settings from './pages/Settings';
import Header from './components/Header';
import { ThemeProvider } from '@mui/material/styles'
import theme from './theme'
import { ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import Friends from './pages/Friends';

function App() {
  return (
    <ThemeProvider theme={theme}>
    <AuthProvider>
      <ChatProvider>
        <Router>
          <Header />
          <Routes>
          <Route path="/" element={<Navigate to="/login" replace />} />
            {/* Public routes */}
            <Route path="/login" element={<Login />} />
            <Route path="/signup" element={<Signup />} />

            {/* Protected routes */}
            <Route path="/me" element={<PrivateRoute><MyProfile /></PrivateRoute>} />
            <Route path="/edit-profile" element={<PrivateRoute><EditProfile /></PrivateRoute>} />
            <Route path="/recommendations" element={<PrivateRoute><Recommendations /></PrivateRoute>} />
            <Route path="/connections" element={<PrivateRoute><Connections /></PrivateRoute>} />
            <Route path="/chats" element={<PrivateRoute><Chats /></PrivateRoute>} />
            <Route path="/chat/:chatId" element={<PrivateRoute><ChatWindow /></PrivateRoute>} />
            <Route path="/settings" element={<PrivateRoute><Settings /></PrivateRoute>} />
            <Route path="/friends"   element={<PrivateRoute><Friends  /></PrivateRoute>} />
            <Route path="*" element={<Navigate to="/login" replace />} />
          </Routes>
          <ToastContainer position="top-right" autoClose={2000} />
        </Router>
      </ChatProvider>
    </AuthProvider>
    </ThemeProvider>
  );
}

export default  App;
