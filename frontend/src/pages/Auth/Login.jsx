// /m/frontend/src/pages/Auth/Login.jsx
import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Container, Box, Typography, TextField, Button } from '@mui/material';
import { Formik, Form, Field, ErrorMessage } from 'formik';
import * as Yup from 'yup';
import { useAuthDispatch } from '../../contexts/AuthContext';
import { login } from '../../api/auth';
import { toast } from 'react-toastify';
export const ADMIN_EMAIL = "admin@first.av";

const LoginSchema = Yup.object().shape({
  email: Yup.string()
    .email('Некорректный формат email')
    .required('Введите email'),
  password: Yup.string()
    .min(8, 'Пароль должен быть минимум 8 символов')
    .required('Введите пароль'),
});

const Login = () => {
  const navigate = useNavigate();
  const dispatch = useAuthDispatch();

  // const handleSubmit = async (values, { setSubmitting }) => {
  //   try {
  //     const data = await login({
  //       email: values.email,
  //       password: values.password,
  //     });
  //     dispatch({ type: 'LOGIN_SUCCESS', payload: data });
  //     toast.success('Успешный вход в систему');
  //     navigate('/me');
  //   } catch (err) {
  //     const msg =
  //       err.response?.data?.message ||
  //       'Ошибка входа. Проверьте введённые данные.';
  //     toast.error(msg);
  //   } finally {
  //     setSubmitting(false);
  //   }
  // };

  const handleSubmit = async (values, { setSubmitting }) => {
    try {
      const data = await login({
        email: values.email,
        password: values.password,
      });
  
      // 🔒 Validate presence of token
      if (!data || !data.accessToken) {
        throw new Error('Ошибка: не удалось получить токен.');
      }
  
      dispatch({ type: 'LOGIN_SUCCESS', payload: data });
      toast.success('Успешный вход в систему');
  
      // 🔍 Optionally fetch profile here BEFORE navigating
      // const profile = await GetCurrentUserProfile().catch(() => null);
      // if (!profile) {
      //   toast.error('Ошибка входа. Проверьте введённые данные.');
      //   return;
      // }
  
      // ✅ Now route only if profile exists
      if (values.email.toLowerCase() === ADMIN_EMAIL.toLowerCase()) {
        navigate('/admin');
      } else {
        navigate('/me');
      }
  
    } catch (err) {
      const msg =
        err.response?.data?.message ||
        'Ошибка входа. Проверьте введённые данные.';
      toast.error(msg);
    } finally {
      setSubmitting(false);
    }
  };
  
  

  // const handleSubmit = async (values, { setSubmitting }) => {
  //   try {
  //     const data = await login({
  //       email: values.email,
  //       password: values.password,
  //     });
  
  //     // Make sure login was truly successful
  //     if (!data || !data.accessToken) {
  //       throw new Error('Ошибка: пользователь не найден или токен отсутствует');
  //     }
  
  //     dispatch({ type: 'LOGIN_SUCCESS', payload: data });
  //     toast.success('Успешный вход в систему');
  
  //     if (values.email.toLowerCase() === ADMIN_EMAIL.toLowerCase()) {
  //       navigate('/admin');
  //     } else {
  //       navigate('/me');
  //     }
  //   } catch (err) {
  //     const msg =
  //       err.response?.data?.message ||
  //       err.message ||
  //       'Ошибка входа. Проверьте введённые данные.';
  //     toast.error(msg);
  //   } finally {
  //     setSubmitting(false);
  //   }
  // }; 

  return (
    <Container maxWidth="sm">
      <Box sx={{ mt: 4, p: 3, border: '1px solid #ccc', borderRadius: 2 }}>
        <Typography variant="h4" gutterBottom>
          Вход
        </Typography>

        <Formik
          initialValues={{ email: '', password: '' }}
          validationSchema={LoginSchema}
          onSubmit={handleSubmit}
        >
          {({ isSubmitting, touched, errors }) => (
            <Form>
              <Field
                name="email"
                as={TextField}
                label="Email"
                type="email"
                autoComplete="username" 
                fullWidth
                margin="normal"
                error={touched.email && Boolean(errors.email)}
                helperText={<ErrorMessage name="email" />}
              />

              <Field
                name="password"
                as={TextField}
                label="Пароль"
                type="password"
                autoComplete="current-password"
                fullWidth
                margin="normal"
                error={touched.password && Boolean(errors.password)}
                helperText={<ErrorMessage name="password" />}
              />

              <Button
                variant="contained"
                color="primary"
                type="submit"
                fullWidth
                sx={{ mt: 2 }}
                disabled={isSubmitting}
              >
                {isSubmitting ? 'Вход...' : 'Войти'}
              </Button>

              <Typography variant="body2" sx={{ mt: 2 }}>
                Нет аккаунта? <Link to="/signup">Зарегистрироваться</Link>
              </Typography>
            </Form>
          )}
        </Formik>
      </Box>
    </Container>
  );
};

export default Login;
