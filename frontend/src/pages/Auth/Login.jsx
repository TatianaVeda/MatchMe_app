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
    .email('Invalid email format. Must be in : example@domain.com')
    .required('Enter email'),
  password: Yup.string()
    .min(8, 'Password must be at least 8 characters, including letters, numbers, and special characters (@$!%*#?&)')
    /* .matches(
      /^(?=.*[A-Za-z])(?=.*\d)(?=.*[@$!%*#?&])[A-Za-z\d@$!%*#?&]{8,}$/,
    ) */
    .required('Enter password'),
});

const Login = () => {
  const navigate = useNavigate();
  const dispatch = useAuthDispatch();

  const handleSubmit = async (values, { setSubmitting }) => {
    try {
      const data = await login({
        email: values.email,
        password: values.password,
      });
  
      if (!data || !data.accessToken) {
        throw new Error('Error: Failed to get token.');
      }
  
      dispatch({ type: 'LOGIN_SUCCESS', payload: data });
      toast.success('Successfully logged in');
      if (values.email.toLowerCase() === ADMIN_EMAIL.toLowerCase()) {
        navigate('/admin');
      } else {
        navigate('/me');
      }
  
    } catch (err) {
      const msg =
        err.response?.data?.message ||
        'Login error. Please check your credentials.';
      toast.error(msg);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Container maxWidth="sm">
      <Box sx={{ mt: 4, p: 3, border: '1px solid #ccc', borderRadius: 2 }}>
        <Typography variant="h4" gutterBottom>
          Login
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
                label="Password"
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
                {isSubmitting ? 'Signing in...' : 'Sign In'}
              </Button>

              <Typography variant="body2" sx={{ mt: 2 }}>
                Don't have an account? <Link to="/signup">Sign Up</Link>
              </Typography>
            </Form>
          )}
        </Formik>
      </Box>
    </Container>
  );
};

export default Login;
