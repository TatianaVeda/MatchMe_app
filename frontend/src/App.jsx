//В ИТОГЕ СРАВНИТЬ ОБА ФАЙЛА APP.JS И APP.JSX!!!! APP.JSX должен быть в проекте а не APP.JS
// m/frontend/src/App.jsx
import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { AuthProvider } from './contexts/AuthContext';
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
// Импорт остальных страниц

function App() {
  return (
    <AuthProvider>
      <Router>
        <Routes>
          {/* Публичные маршруты */}
          <Route path="/login" element={<Login />} />
          <Route path="/signup" element={<Signup />} />
            {/* Защищённые маршруты */}
          <Route path="/me" element={<PrivateRoute><MyProfile /></PrivateRoute>} />
          <Route path="/edit-profile" element={<PrivateRoute><EditProfile /></PrivateRoute>} />
          <Route path="/recommendations" element={<PrivateRoute><Recommendations /></PrivateRoute>} />
          <Route path="/connections" element={<PrivateRoute><Connections /></PrivateRoute>} />
          <Route path="/chats" element={<PrivateRoute><Chats /></PrivateRoute>} />
          <Route path="/chat/:chatId" element={<PrivateRoute><ChatWindow /></PrivateRoute>} />
          <Route path="/settings" element={<PrivateRoute><Settings /></PrivateRoute>} />
        </Routes>
      </Router>
    </AuthProvider>
  );
}

export default App;
