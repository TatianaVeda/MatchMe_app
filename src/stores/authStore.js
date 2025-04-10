import { create } from 'zustand'
import axios from 'axios'
import { jwtDecode } from 'jwt-decode'

export const useAuthStore = create((set) => ({
  token: localStorage.getItem('token'),
  user: null,
  isAuthenticated: !!localStorage.getItem('token'),
  
  login: async (email, password) => {
    try {
      const response = await axios.post('/api/login', { email, password })
      const { token, user } = response.data
      localStorage.setItem('token', token)
      set({ token, user, isAuthenticated: true })
      return true
    } catch (error) {
      console.error('Login failed:', error)
      return false
    }
  },

  register: async (userData) => {
    try {
      const response = await axios.post('/api/register', userData)
      return true
    } catch (error) {
      console.error('Registration failed:', error);
      console.error('Registration failed message:', error.message);
      console.error('Registration failed response:', error.response);
      return false
    }
  },

  logout: () => {
    localStorage.removeItem('token')
    set({ token: null, user: null, isAuthenticated: false })
  },

  checkAuth: () => {
    const token = localStorage.getItem('token')
    if (token) {
      try {
        const decoded = jwtDecode(token)
        const currentTime = Date.now() / 1000
        if (decoded.exp < currentTime) {
          localStorage.removeItem('token')
          set({ token: null, user: null, isAuthenticated: false })
        }
      } catch (error) {
        localStorage.removeItem('token')
        set({ token: null, user: null, isAuthenticated: false })
      }
    }
  }
}))

// Add axios interceptor for JWT
axios.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)