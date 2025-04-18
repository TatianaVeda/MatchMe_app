import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { toast } from 'react-toastify';
import { Container, Box, Typography, TextField, Button } from '@mui/material';
import { useAuthDispatch } from '../../contexts/AuthContext';
// Импортируем метод login из auth.js
import { login } from '../../api/auth';

const Login = () => {
  const [formData, setFormData] = useState({ email: '', password: '' });
  const navigate = useNavigate();
  const dispatch = useAuthDispatch();
  const [loading, setLoading] = useState(false);

  // Обработка изменения полей формы
  const handleChange = (e) => {
    setFormData({ ...formData, [e.target.name]: e.target.value });
  };

  // Обработка отправки формы
  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    try {
      // Используем метод login вместо прямого вызова axios.post
      const data = await login(formData);
      // Ожидается, что сервер вернет объект с полями user, accessToken, refreshToken
      dispatch({ type: 'LOGIN_SUCCESS', payload: data });
      toast.success("Успешный вход в систему");
      navigate('/me');
    } catch (err) {
      const errorMessage =
        err.response?.data?.message || "Ошибка входа. Проверьте введённые данные.";
      toast.error(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Container maxWidth="sm">
      <Box 
        component="form"
        onSubmit={handleSubmit}
        sx={{ mt: 4, p: 3, border: '1px solid #ccc', borderRadius: 2 }}
      >
        <Typography variant="h4" component="h1" gutterBottom>
          Вход
        </Typography>
        <TextField
          label="Email"
          name="email"
          type="email"
          fullWidth
          margin="normal"
          value={formData.email}
          onChange={handleChange}
          required
        />
        <TextField
          label="Пароль"
          name="password"
          type="password"
          fullWidth
          margin="normal"
          value={formData.password}
          onChange={handleChange}
          required
        />
        <Button
          variant="contained"
          color="primary"
          type="submit"
          fullWidth
          sx={{ mt: 2 }}
          disabled={loading}
        >
          {loading ? "Вход..." : "Войти"}
        </Button>
        <Typography variant="body2" sx={{ mt: 2 }}>
          Нет аккаунта? <Link to="/signup">Зарегистрироваться</Link>
        </Typography>
      </Box>
    </Container>
  );
};

export default Login;
