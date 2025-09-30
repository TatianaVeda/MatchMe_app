// m/frontend/src/components/Header.jsx
import React, { useState, useEffect, useCallback } from 'react';
import {
  AppBar, Toolbar, Typography, IconButton,
  Button, Drawer, List, ListItem, ListItemButton,
  ListItemText, Box, Badge
} from '@mui/material';
import MenuIcon from '@mui/icons-material/Menu';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { useAuthState, useAuthDispatch } from '../contexts/AuthContext';
import { useChatState } from '../contexts/ChatContext';
import useWebSocket from '../hooks/useWebSocket';
import { getPendingConnections } from '../api/connections';
import axios from '../api/index';
import { toast } from 'react-toastify';
import { ADMIN_ID } from '../config';
import RecommendIcon from '@mui/icons-material/Recommend';
import ChatIcon from '@mui/icons-material/Chat';
import PersonIcon from '@mui/icons-material/Person';
import SettingsIcon from '@mui/icons-material/Settings';
import GroupIcon from '@mui/icons-material/Group';
import LogoutIcon from '@mui/icons-material/Logout';
import AdminPanelSettingsIcon from '@mui/icons-material/AdminPanelSettings';

const navItems = [
  { label: 'Recommendations', to: '/recommendations', icon: <RecommendIcon sx={{ mr: 1 }} /> },
  { label: 'Chats',        to: '/chats', icon: <ChatIcon sx={{ mr: 1 }} /> },
  { label: 'Profile',     to: '/me', icon: <PersonIcon sx={{ mr: 1 }} /> },
  { label: 'Settings',   to: '/settings', icon: <SettingsIcon sx={{ mr: 1 }} /> },
  { label: 'Friends',      to: '/friends', icon: <GroupIcon sx={{ mr: 1 }} /> },
];

export default function Header() {
  const { user } = useAuthState();
  const dispatch = useAuthDispatch();
  const navigate = useNavigate();
  const location = useLocation();
  const [mobileOpen, setMobileOpen] = useState(false);

  const { chats, messages } = useChatState();
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
    const chatsArray = Array.isArray(chats) ? chats : [];
    setUnreadMessages(chatsArray.reduce((sum, c) => sum + (c.unreadCount || 0), 0));
  }, [user, chats]);

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
    <Box onClick={toggleDrawer} sx={{ width: 320 }}>
      <List>
        {user && user.id === ADMIN_ID ? (
          <>
            <ListItem disablePadding>
              <ListItemButton component={Link} to="/admin">
                <AdminPanelSettingsIcon sx={{ mr: 1 }} />
                <ListItemText primary="Admin" />
              </ListItemButton>
            </ListItem>
            <ListItem disablePadding>
              <ListItemButton onClick={handleLogout}>
                <LogoutIcon sx={{ mr: 1 }} />
                <ListItemText primary="Logout" />
              </ListItemButton>
            </ListItem>
          </>
        ) : user ? (
          <>
            {navItems.map(item => (
              <ListItem key={item.to} disablePadding>
                <ListItemButton
                  component={Link}
                  to={item.to}
                  sx={{
                    fontWeight: location.pathname === item.to ? 'bold' : 'normal',
                    backgroundColor: location.pathname === item.to ? 'primary.main' : 'inherit',
                    color: location.pathname === item.to ? 'white' : 'inherit',
                    border: location.pathname === item.to ? '1.5px solid #fff' : '1.5px solid transparent',
                    borderRadius: location.pathname === item.to ? '8px' : 0,
                    margin: location.pathname === item.to ? '4px 8px' : '0',
                    transition: 'all 0.2s',
                    boxShadow: location.pathname === item.to ? '0 2px 8px rgba(25, 118, 210, 0.15)' : 'none'
                  }}
                >
                  {item.to === '/chats' ? (
                    <Badge color="error" badgeContent={unreadMessages} invisible={unreadMessages === 0} sx={{ mr: 1 }}>
                      {item.icon}
                    </Badge>
                  ) : item.to === '/friends' ? (
                    <Badge color="error" badgeContent={pendingFriends} invisible={pendingFriends === 0} sx={{ mr: 1 }}>
                      {item.icon}
                    </Badge>
                  ) : (
                    item.icon
                  )}
                  <ListItemText primary={item.label} />
                </ListItemButton>
              </ListItem>
            ))}
            <ListItem disablePadding>
              <ListItemButton onClick={handleLogout}>
                <LogoutIcon sx={{ mr: 1 }} />
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
      <AppBar position="fixed">
        <Toolbar>
          <IconButton
            color="inherit"
            edge="start"
            onClick={toggleDrawer}
            sx={{ display: { md: 'none' }, mr: 2 }}
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

          <Box sx={{ display: { xs: 'none', md: 'block' } }}>
            {user ? (
              user.id === ADMIN_ID ? (
                <>
                  <Button color="inherit" component={Link} to="/admin" sx={{ ml: 1 }}>
                    <AdminPanelSettingsIcon sx={{ mr: 1 }} />
                    Admin
                  </Button>
                  <Button color="inherit" onClick={handleLogout} sx={{ ml: 1 }}>
                    <LogoutIcon sx={{ mr: 1 }} />
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
                            sx={{
                              fontWeight: location.pathname === item.to ? 'bold' : 'normal',
                              backgroundColor: location.pathname === item.to ? 'primary.main' : 'inherit',
                              color: location.pathname === item.to ? 'white' : 'inherit',
                              border: location.pathname === item.to ? '1.5px solid #fff' : '1.5px solid transparent',
                              borderRadius: location.pathname === item.to ? '8px' : 0,
                              margin: location.pathname === item.to ? '4px 8px' : '0',
                              transition: 'all 0.2s',
                              boxShadow: location.pathname === item.to ? '0 2px 8px rgba(25, 118, 210, 0.15)' : 'none'
                            }}
                          >
                            {item.icon}
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
                            sx={{
                              fontWeight: location.pathname === item.to ? 'bold' : 'normal',
                              backgroundColor: location.pathname === item.to ? 'primary.main' : 'inherit',
                              color: location.pathname === item.to ? 'white' : 'inherit',
                              border: location.pathname === item.to ? '1.5px solid #fff' : '1.5px solid transparent',
                              borderRadius: location.pathname === item.to ? '8px' : 0,
                              margin: location.pathname === item.to ? '4px 8px' : '0',
                              transition: 'all 0.2s',
                              boxShadow: location.pathname === item.to ? '0 2px 8px rgba(25, 118, 210, 0.15)' : 'none'
                            }}
                          >
                            {item.icon}
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
                        sx={{
                          ml: 1,
                          fontWeight: location.pathname === item.to ? 'bold' : 'normal',
                          backgroundColor: location.pathname === item.to ? 'primary.main' : 'inherit',
                          color: location.pathname === item.to ? 'white' : 'inherit',
                          border: location.pathname === item.to ? '1.5px solid #fff' : '1.5px solid transparent',
                          borderRadius: location.pathname === item.to ? '8px' : 0,
                          margin: location.pathname === item.to ? '4px 8px' : '0',
                          transition: 'all 0.2s',
                          boxShadow: location.pathname === item.to ? '0 2px 8px rgba(25, 118, 210, 0.15)' : 'none'
                        }}
                      >
                        {item.icon}
                        {item.label}
                      </Button>
                    );
                  })}
                  <Button
                    color="inherit"
                    onClick={handleLogout}
                    sx={{ ml: 1 }}
                  >
                    <LogoutIcon sx={{ mr: 1 }} />
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
