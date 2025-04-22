import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Container, Box, Typography, TextField, Button } from '@mui/material';
import { useAuthDispatch } from '../../contexts/AuthContext';
import { signup } from '../../api/auth';
import { toast } from 'react-toastify';
import { Formik, Form, Field, ErrorMessage } from 'formik';
import * as Yup from 'yup';

const SignupSchema = Yup.object().shape({
  email: Yup.string()
    .email('Некорректный формат email')
    .required('Введите email'),
  password: Yup.string()
    .min(8, 'Пароль должен быть минимум 8 символов')
    .required('Введите пароль'),
  confirmPassword: Yup.string()
    .oneOf([Yup.ref('password'), null], 'Пароли не совпадают')
    .required('Подтвердите пароль'),
});

const Signup = () => {
  const navigate = useNavigate();
  const dispatch = useAuthDispatch();

  const handleSubmit = async (values, { setSubmitting }) => {
    try {
      const data = await signup({
        email: values.email,
        password: values.password,
      });
      dispatch({ type: 'LOGIN_SUCCESS', payload: data });
      toast.success('Регистрация прошла успешно!');
      navigate('/me');
    } catch (err) {
      toast.error(err.response?.data || 'Ошибка регистрации');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Container maxWidth="sm">
      <Box sx={{ mt: 4, p: 3, border: '1px solid #ccc', borderRadius: 2 }}>
        <Typography variant="h4" gutterBottom>
          Регистрация
        </Typography>

        <Formik
          initialValues={{ email: '', password: '', confirmPassword: '' }}
          validationSchema={SignupSchema}
          onSubmit={handleSubmit}
        >
          {({ isSubmitting, touched, errors }) => (
            <Form>
              <Field
                name="email"
                as={TextField}
                label="Email"
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

              <Field
                name="confirmPassword"
                as={TextField}
                label="Подтвердите пароль"
                type="password"
                fullWidth
                margin="normal"
                error={touched.confirmPassword && Boolean(errors.confirmPassword)}
                helperText={<ErrorMessage name="confirmPassword" />}
              />

              <Button
                variant="contained"
                color="primary"
                type="submit"
                fullWidth
                sx={{ mt: 2 }}
                disabled={isSubmitting}
              >
                {isSubmitting ? 'Регистрация...' : 'Зарегистрироваться'}
              </Button>

              <Typography variant="body2" sx={{ mt: 2 }}>
                Уже есть аккаунт? <Link to="/login">Войти</Link>
              </Typography>
            </Form>
          )}
        </Formik>
      </Box>
    </Container>
  );
};

export default Signup;
