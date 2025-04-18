import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { toast } from 'react-toastify';
import { Container, Box, Typography, TextField, Button } from '@mui/material';
import { useAuthDispatch } from '../../contexts/AuthContext';
// Импортируем метод signup из auth.js
import { signup } from '../../api/auth';

const Signup = () => {
  const [formData, setFormData] = useState({
    email: '',
    password: '',
    confirmPassword: ''
  });
  const navigate = useNavigate();
  const dispatch = useAuthDispatch();
  const [loading, setLoading] = useState(false);
  const [formErrors, setFormErrors] = useState({});

  // Обработка изменения полей формы
  const handleChange = (e) => {
    setFormData({ ...formData, [e.target.name]: e.target.value });
    if (formErrors[e.target.name]) {
      setFormErrors({ ...formErrors, [e.target.name]: '' });
    }
  };

  // Простая валидация данных
  const validate = () => {
    const errors = {};
    if (!formData.email) errors.email = "Введите email";
    if (formData.email && !/\S+@\S+\.\S+/.test(formData.email)) {
      errors.email = "Некорректный формат email";
    }
    if (!formData.password) errors.password = "Введите пароль";
    if (formData.password && formData.password.length < 8) {
      errors.password = "Пароль должен содержать минимум 8 символов";
    }
    if (formData.password !== formData.confirmPassword) {
      errors.confirmPassword = "Пароли не совпадают";
    }
    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  };

  // Обработка отправки формы регистрации
  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!validate()) return;
    setLoading(true);
    try {
      // Используем метод signup для регистрации пользователя
      const data = await signup({
        email: formData.email,
        password: formData.password
      });
      // После успешной регистрации пользователь автоматически аутентифицируется
      dispatch({ type: 'LOGIN_SUCCESS', payload: data });
      toast.success("Регистрация прошла успешно!");
      navigate('/me');
    } catch (err) {
      const errorMessage =
        err.response?.data?.message ||
        "Ошибка регистрации. Проверьте введённые данные.";
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
          Регистрация
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
          error={!!formErrors.email}
          helperText={formErrors.email}
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
          error={!!formErrors.password}
          helperText={formErrors.password}
        />
        <TextField
          label="Подтвердите пароль"
          name="confirmPassword"
          type="password"
          fullWidth
          margin="normal"
          value={formData.confirmPassword}
          onChange={handleChange}
          required
          error={!!formErrors.confirmPassword}
          helperText={formErrors.confirmPassword}
        />
        <Button
          variant="contained"
          color="primary"
          type="submit"
          fullWidth
          sx={{ mt: 2 }}
          disabled={loading}
        >
          {loading ? "Регистрация..." : "Зарегистрироваться"}
        </Button>
        <Typography variant="body2" sx={{ mt: 2 }}>
          Уже есть аккаунт? <Link to="/login">Войти</Link>
        </Typography>
      </Box>
    </Container>
  );
};

export default Signup;
