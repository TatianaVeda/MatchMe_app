import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuthStore } from './stores/authStore'
import Layout from './components/Layout'
import Login from './pages/Login'
import Register from './pages/Register'
import Profile from './pages/Profile'
import Recommendations from './pages/Recommendations'
import Connections from './pages/Connections'
import Chat from './pages/Chat'

function PrivateRoute({ children }) {
  const isAuthenticated = useAuthStore(state => state.isAuthenticated)
  return isAuthenticated ? children : <Navigate to="/login" />
}

function PublicRoute({ children }) {
  const isAuthenticated = useAuthStore(state => state.isAuthenticated)
  return !isAuthenticated ? children : <Navigate to="/recommendations" />
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={
        <PublicRoute>
          <Login />
        </PublicRoute>
      } />
      <Route path="/register" element={
        <PublicRoute>
          <Register />
        </PublicRoute>
      } />
      <Route path="/" element={
        <PrivateRoute>
          <Layout />
        </PrivateRoute>
      }>
        <Route index element={<Navigate to="/recommendations" replace />} />
        <Route path="profile" element={<Profile />} />
        <Route path="recommendations" element={<Recommendations />} />
        <Route path="connections" element={<Connections />} />
        <Route path="chat" element={<Chat />} />
        <Route path="chat/:userId" element={<Chat />} />
      </Route>
    </Routes>
  )
}