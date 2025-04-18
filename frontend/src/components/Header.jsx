// m/frontend/src/components/Header.jsx
import React from 'react';
import { AppBar, Toolbar, Typography, Button } from '@mui/material';
import { Link, useNavigate } from 'react-router-dom';
import { useAuthState, useAuthDispatch } from '../contexts/AuthContext';
import axios from '../api/index';
import { toast } from 'react-toastify';

const Header = () => {
  const { user } = useAuthState();
  const dispatch = useAuthDispatch();
  const navigate = useNavigate();

  const handleLogout = async () => {
    try {
      await axios.post('/logout');
    } catch (error) {
      // Если выход не удался – можно залогировать ошибку и продолжить очистку
    }
    dispatch({ type: 'LOGOUT' });
    toast.info("Вы вышли из системы");
    navigate('/login');
  };

  return (
    <AppBar position="static">
      <Toolbar>
        <Typography variant="h6" component={Link} to="/" sx={{ flexGrow: 1, textDecoration: 'none', color: 'inherit' }}>
          m – Рекомендации
        </Typography>
        {user ? (
          <>
            <Button color="inherit" component={Link} to="/recommendations">Рекомендации</Button>
            <Button color="inherit" component={Link} to="/chats">Чаты</Button>
            <Button color="inherit" onClick={handleLogout}>Выход</Button>
          </>
        ) : (
          <Button color="inherit" component={Link} to="/login">Вход</Button>
        )}
      </Toolbar>
    </AppBar>
  );
};

export default Header;


