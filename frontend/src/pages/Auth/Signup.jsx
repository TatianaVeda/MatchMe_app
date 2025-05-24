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
    .email('Invalid email format')
    .required('Enter email'),
  password: Yup.string()
    .min(8, 'Password must be at least 8 characters')
    .required('Enter password'),
  confirmPassword: Yup.string()
    .oneOf([Yup.ref('password'), null], 'Passwords do not match')
    .required('Confirm password'),
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
      toast.success('Registration successful!');
      navigate('/me');
    } catch (err) {
      toast.error(err.response?.data || 'Registration error');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Container maxWidth="sm">
      <Box sx={{ mt: 4, p: 3, border: '1px solid #ccc', borderRadius: 2 }}>
        <Typography variant="h4" gutterBottom>
          Registration
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
                label="Password"
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
                label="Please confirm password"
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
                {isSubmitting ? 'Registering...' : 'Sign Up'}
              </Button>

              <Typography variant="body2" sx={{ mt: 2 }}>
                Already have an account? <Link to="/login">Sign In</Link>
              </Typography>
            </Form>
          )}
        </Formik>
      </Box>
    </Container>
  );
};

export default Signup;
