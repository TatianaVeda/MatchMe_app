import React, { useState } from 'react';
import { AppBar, Toolbar, Typography, IconButton, Button, Drawer, List, ListItem, ListItemButton,
  ListItemText, Box
} from '@mui/material';
import MenuIcon from '@mui/icons-material/Menu';
import { Link, useNavigate } from 'react-router-dom';
import { useAuthState, useAuthDispatch } from '../contexts/AuthContext';
import axios from '../api/index';
import { toast } from 'react-toastify';
import WebSocketService from '../services/websocketService';
import { ADMIN_ID } from '../config';


const navItems = [
  { label: 'Рекомендации', to: '/recommendations' },
  { label: 'Чаты',        to: '/chats' },
  { label: 'Профиль',     to: '/me' },
  { label: 'Настройки',   to: '/settings' },
  { label: 'Друзья',      to: '/friends' },
];

const Header = () => {
  const { user } = useAuthState();
  const dispatch = useAuthDispatch();
  const navigate = useNavigate();
  const [mobileOpen, setMobileOpen] = useState(false);

  const handleLogout = async () => {
    try { await axios.post('/logout'); } catch {}
    WebSocketService.disconnect();
    dispatch({ type: 'LOGOUT' });
    toast.info("Вы вышли из системы");
    navigate('/login');
  };

  const toggleDrawer = () => setMobileOpen(open => !open);

  const drawer = (
    <Box onClick={toggleDrawer} sx={{ width: 250 }}>
      <List>
      {user && navItems.map(item => (
     <ListItem key={item.to} disablePadding>
       <ListItemButton component={Link} to={item.to}>
         <ListItemText primary={item.label} />
       </ListItemButton>
     </ListItem>
   ))}
        {user ? (
   <ListItem disablePadding>
     <ListItemButton onClick={handleLogout}>
       <ListItemText primary="Выход" />
     </ListItemButton>
   </ListItem>
 ) : (
   <ListItem disablePadding>
     <ListItemButton component={Link} to="/login">
       <ListItemText primary="Вход" />
     </ListItemButton>
   </ListItem>
 )}
      </List>
    </Box>
  );

  return (
    <>
      <AppBar position="static">
        <Toolbar>
          {/* Гамбургер для мобильных */}
          <IconButton
            color="inherit"
            aria-label="menu"
            edge="start"
            onClick={toggleDrawer}
            sx={{ display: { sm: 'none' }, mr: 2 }}
          >
            <MenuIcon />
          </IconButton>

          {/* Лого / Название */}
          <Typography
            variant="h6"
            component={Link}
            to="/"
            sx={{ flexGrow: 1, textDecoration: 'none', color: 'inherit' }}
          >
            m – Рекомендации
          </Typography>

          {/* Десктопное меню */}
          <Box sx={{ display: { xs: 'none', sm: 'block' } }}>
            {user ? (
              <>
                {navItems.map(item => (
                  <Button key={item.to}
                    color="inherit" component={Link}  to={item.to}>
                    {item.label}
                  </Button>
                ))}
                  {user.id === ADMIN_ID && (
                <Button color="inherit" component={Link} to="/admin">
                  Admin
                </Button>
              )}
                <Button color="inherit" onClick={handleLogout}>
                  Выход
                </Button>
              </>
            ) : (
              <Button color="inherit" component={Link} to="/login">
                Вход
              </Button>
            )}
          </Box>
        </Toolbar>
      </AppBar>

      {/* Мобильный Drawer */}
      <Drawer
        anchor="left"
        open={mobileOpen}
        onClose={toggleDrawer}
        ModalProps={{ keepMounted: true }}
      >
        {drawer}
      </Drawer>
    </>
  );
};

export default Header;
