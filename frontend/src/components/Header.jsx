// m/frontend/src/components/Header.jsx
import React, { useState, useEffect, useCallback } from 'react';
import {
  AppBar, Toolbar, Typography, IconButton,
  Button, Drawer, List, ListItem, ListItemButton,
  ListItemText, Box, Badge
} from '@mui/material';
import MenuIcon from '@mui/icons-material/Menu';
import { Link, useNavigate } from 'react-router-dom';
import { useAuthState, useAuthDispatch } from '../contexts/AuthContext';
import { useChatState } from '../contexts/ChatContext';
import useWebSocket from '../hooks/useWebSocket';
import { getPendingConnections } from '../api/connections';
import axios from '../api/index';
import { toast } from 'react-toastify';
import { ADMIN_ID } from '../config';

const navItems = [
  { label: 'Recommendations', to: '/recommendations' },
  { label: 'Chats',        to: '/chats' },
  { label: 'Profile',     to: '/me' },
  { label: 'Settings',   to: '/settings' },
  { label: 'Friends',      to: '/friends' },
];

export default function Header() {
  const { user } = useAuthState();
  const dispatch = useAuthDispatch();
  const navigate = useNavigate();
  const [mobileOpen, setMobileOpen] = useState(false);

  const { chats } = useChatState();
  const [unreadMessages, setUnreadMessages] = useState(0);

  const [pendingFriends, setPendingFriends] = useState(0);

  const { subscribe, unsubscribe } = useWebSocket((data) => {
    if (!user) return;
    switch (data.type) {
      case 'message':
        if (data.sender_id !== user.id) {
          setUnreadMessages(u => u + 1);
        }
        break;
      case 'connection_request':
        setPendingFriends(p => p + 1);
        break;
      default:
        break;
    }
  });

  useEffect(() => {
    if (!user) return;

    setUnreadMessages(chats.reduce((sum, c) => sum + (c.unreadCount || 0), 0));

    chats.forEach(c => subscribe(c.id));

    return () => {
      chats.forEach(c => unsubscribe(c.id));
    };
  }, [user, chats, subscribe, unsubscribe]);

  useEffect(() => {
    if (!user) return;
    getPendingConnections()
      .then(list => setPendingFriends(list.length))
      .catch(() => {/* silent fallback */});
  }, [user]);

  const handleLogout = async () => {
    try { await axios.post('/logout'); } catch {}
    dispatch({ type: 'LOGOUT' });
    toast.info('You have logged out');
    navigate('/login');
  };

  const toggleDrawer = () => setMobileOpen(o => !o);

  const drawer = (
    <Box onClick={toggleDrawer} sx={{ width: 250 }}>
      <List>
        {user && user.id === ADMIN_ID ? (
          <>
            <ListItem disablePadding>
              <ListItemButton component={Link} to="/admin">
                <ListItemText primary="Admin" />
              </ListItemButton>
            </ListItem>
            <ListItem disablePadding>
              <ListItemButton onClick={handleLogout}>
                <ListItemText primary="Logout" />
              </ListItemButton>
            </ListItem>
          </>
        ) : user ? (
          <>
            {navItems.map(item => (
              <ListItem key={item.to} disablePadding>
                <ListItemButton component={Link} to={item.to}>
                  <ListItemText primary={item.label} />
                </ListItemButton>
              </ListItem>
            ))}
            <ListItem disablePadding>
              <ListItemButton onClick={handleLogout}>
                <ListItemText primary="Logout" />
              </ListItemButton>
            </ListItem>
          </>
        ) : (
          <ListItem disablePadding>
            <ListItemButton component={Link} to="/login">
              <ListItemText primary="Login" />
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
          <IconButton
            color="inherit"
            edge="start"
            onClick={toggleDrawer}
            sx={{ display: { sm: 'none' }, mr: 2 }}
          >
            <MenuIcon />
          </IconButton>

          <Typography
            variant="h6"
            component={Link}
            to="/"
            sx={{ flexGrow: 1, textDecoration: 'none', color: 'inherit' }}
          >
            Match Me – Recommendations
          </Typography>

          <Box sx={{ display: { xs: 'none', sm: 'block' } }}>
            {user ? (
              user.id === ADMIN_ID ? (
                <>
                  <Button color="inherit" component={Link} to="/admin">
                    Admin
                  </Button>
                  <Button color="inherit" onClick={handleLogout}>
                    Logout
                  </Button>
                </>
              ) : (
                <>
                  {navItems.map(item => {
                    if (item.to === '/chats') {
                      return (
                        <Badge
                          key={item.to}
                          color="error"
                          badgeContent={unreadMessages}
                          invisible={unreadMessages === 0}
                          sx={{ ml: 1 }}
                        >
                          <Button
                            color="inherit"
                            component={Link}
                            to={item.to}
                          >
                            {item.label}
                          </Button>
                        </Badge>
                      );
                    }
                    if (item.to === '/friends') {
                      return (
                        <Badge
                          key={item.to}
                          color="error"
                          badgeContent={pendingFriends}
                          invisible={pendingFriends === 0}
                          sx={{ ml: 1 }}
                        >
                          <Button
                            color="inherit"
                            component={Link}
                            to={item.to}
                          >
                            {item.label}
                          </Button>
                        </Badge>
                      );
                    }
                    return (
                      <Button
                        key={item.to}
                        color="inherit"
                        component={Link}
                        to={item.to}
                        sx={{ ml: 1 }}
                      >
                        {item.label}
                      </Button>
                    );
                  })}
                  <Button
                    color="inherit"
                    onClick={handleLogout}
                    sx={{ ml: 1 }}
                  >
                    Logout
                  </Button>
                </>
              )
            ) : (
              <Button color="inherit" component={Link} to="/login">
                Login
              </Button>
            )}
          </Box>
        </Toolbar>
      </AppBar>

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
}
