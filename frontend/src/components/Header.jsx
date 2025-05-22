// import React, { useState } from 'react';
// import { AppBar, Toolbar, Typography, IconButton, Button, Drawer, List, ListItem, ListItemButton,
//   ListItemText, Box
// } from '@mui/material';
// import MenuIcon from '@mui/icons-material/Menu';
// import { Link, useNavigate } from 'react-router-dom';
// import { useAuthState, useAuthDispatch } from '../contexts/AuthContext';
// import axios from '../api/index';
// import { toast } from 'react-toastify';
// import WebSocketService from '../services/websocketService';
// import { ADMIN_ID } from '../config';


// const navItems = [
//   { label: 'Рекомендации', to: '/recommendations' },
//   { label: 'Чаты',        to: '/chats' },
//   { label: 'Профиль',     to: '/me' },
//   { label: 'Настройки',   to: '/settings' },
//   { label: 'Друзья',      to: '/friends' },
// ];

// const Header = () => {
//   const { user } = useAuthState();
//   const dispatch = useAuthDispatch();
//   const navigate = useNavigate();
//   const [mobileOpen, setMobileOpen] = useState(false);

//   const handleLogout = async () => {
//     try { await axios.post('/logout'); } catch {}
//     WebSocketService.disconnect();
//     dispatch({ type: 'LOGOUT' });
//     toast.info("Вы вышли из системы");
//     navigate('/login');
//   };

//   const toggleDrawer = () => setMobileOpen(open => !open);

// //   const drawer = (
// //     <Box onClick={toggleDrawer} sx={{ width: 250 }}>
// //       <List>
// //       {user && navItems.map(item => (
// //      <ListItem key={item.to} disablePadding>
// //        <ListItemButton component={Link} to={item.to}>
// //          <ListItemText primary={item.label} />
// //        </ListItemButton>
// //      </ListItem>
// //    ))}
// //         {user ? (
// //    <ListItem disablePadding>
// //      <ListItemButton onClick={handleLogout}>
// //        <ListItemText primary="Выход" />
// //      </ListItemButton>
// //    </ListItem>
// //  ) : (
// //    <ListItem disablePadding>
// //      <ListItemButton component={Link} to="/login">
// //        <ListItemText primary="Вход" />
// //      </ListItemButton>
// //    </ListItem>
// //  )}
// //       </List>
// //     </Box>
// //   );

// const drawer = (
//   <Box onClick={toggleDrawer} sx={{ width: 250 }}>
//     <List>
//       {user && user.id === ADMIN_ID ? (
//         <>
//           <ListItem disablePadding>
//             <ListItemButton component={Link} to="/admin">
//               <ListItemText primary="Admin" />
//             </ListItemButton>
//           </ListItem>
//           <ListItem disablePadding>
//             <ListItemButton onClick={handleLogout}>
//               <ListItemText primary="Выход" />
//             </ListItemButton>
//           </ListItem>
//         </>
//       ) : user ? (
//         <>
//           {navItems.map(item => (
//             <ListItem key={item.to} disablePadding>
//               <ListItemButton component={Link} to={item.to}>
//                 <ListItemText primary={item.label} />
//               </ListItemButton>
//             </ListItem>
//           ))}
//           <ListItem disablePadding>
//             <ListItemButton onClick={handleLogout}>
//               <ListItemText primary="Выход" />
//             </ListItemButton>
//           </ListItem>
//         </>
//       ) : (
//         <ListItem disablePadding>
//           <ListItemButton component={Link} to="/login">
//             <ListItemText primary="Вход" />
//           </ListItemButton>
//         </ListItem>
//       )}
//     </List>
//   </Box>
// ); 

//   return (
//     <>
//       <AppBar position="static">
//         <Toolbar>
//           {/* Гамбургер для мобильных */}
//           <IconButton
//             color="inherit"
//             aria-label="menu"
//             edge="start"
//             onClick={toggleDrawer}
//             sx={{ display: { sm: 'none' }, mr: 2 }}
//           >
//             <MenuIcon />
//           </IconButton>

//           {/* Лого / Название */}
//           <Typography
//             variant="h6"
//             component={Link}
//             to="/"
//             sx={{ flexGrow: 1, textDecoration: 'none', color: 'inherit' }}
//           >
//             m – Рекомендации
//           </Typography>

//           {/* Десктопное меню */}
//           <Box sx={{ display: { xs: 'none', sm: 'block' } }}>
//             {/* {user ? (
//               <>
//                 {navItems.map(item => (
//                   <Button key={item.to}
//                     color="inherit" component={Link}  to={item.to}>
//                     {item.label}
//                   </Button>
//                 ))}
//                   {user.id === ADMIN_ID && (
//                 <Button color="inherit" component={Link} to="/admin">
//                   Admin
//                 </Button>
//               )}
//                 <Button color="inherit" onClick={handleLogout}>
//                   Выход
//                 </Button>
//               </>
//             ) : (
//               <Button color="inherit" component={Link} to="/login">
//                 Вход
//               </Button>
//             )} */}
//             {user ? (
//   user.id === ADMIN_ID ? (
//     <>
//       <Button color="inherit" component={Link} to="/admin">
//         Admin
//       </Button>
//       <Button color="inherit" onClick={handleLogout}>
//         Выход
//       </Button>
//     </>
//   ) : (
//     <>
//       {navItems.map(item => (
//         <Button key={item.to} color="inherit" component={Link} to={item.to}>
//           {item.label}
//         </Button>
//       ))}
//       <Button color="inherit" onClick={handleLogout}>
//         Выход
//       </Button>
//     </>
//   )
// ) : (
//   <Button color="inherit" component={Link} to="/login">
//     Вход
//   </Button>
// )}

//           </Box>
//         </Toolbar>
//       </AppBar>

//       {/* Мобильный Drawer */}
//       <Drawer
//         anchor="left"
//         open={mobileOpen}
//         onClose={toggleDrawer}
//         ModalProps={{ keepMounted: true }}
//       >
//         {drawer}
//       </Drawer>
//     </>
//   );
// };

// export default Header;

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
  { label: 'Рекомендации', to: '/recommendations' },
  { label: 'Чаты',        to: '/chats' },
  { label: 'Профиль',     to: '/me' },
  { label: 'Настройки',   to: '/settings' },
  { label: 'Друзья',      to: '/friends' },
];

export default function Header() {
  const { user } = useAuthState();
  const dispatch = useAuthDispatch();
  const navigate = useNavigate();
  const [mobileOpen, setMobileOpen] = useState(false);

  // 1) состояние чатов и бейджей
  const { chats } = useChatState();
  const [unreadMessages, setUnreadMessages] = useState(0);

  // 2) заявки в друзья
  const [pendingFriends, setPendingFriends] = useState(0);

  // WS-хук: подключаемся и получаем callback для чистки
  const { subscribe, unsubscribe } = useWebSocket((data) => {
    if (!user) return;
    switch (data.type) {
      case 'message':
        // сообщение из чата, в который мы подписаны
        if (data.sender_id !== user.id) {
          setUnreadMessages(u => u + 1);
        }
        break;
      case 'connection_request':
        // новое заявка в друзья
        setPendingFriends(p => p + 1);
        break;
      default:
        break;
    }
  });

  // при каждом изменении списка chats — обновляем бейдж и подписываемся на WS
  useEffect(() => {
    if (!user) return;

    // a) пересчёт уже полученных
    setUnreadMessages(chats.reduce((sum, c) => sum + (c.unreadCount || 0), 0));

    // b) подписываемся на каждую комнату
    chats.forEach(c => subscribe(c.id));

    // c) при анмонт/обновлении — отписываемся
    return () => {
      chats.forEach(c => unsubscribe(c.id));
    };
  }, [user, chats, subscribe, unsubscribe]);

  // при монтировании Header — инициализируем pendingFriends
  useEffect(() => {
    if (!user) return;
    getPendingConnections()
      .then(list => setPendingFriends(list.length))
      .catch(() => {/* тихий фоллбек */});
  }, [user]);

  const handleLogout = async () => {
    try { await axios.post('/logout'); } catch {}
    dispatch({ type: 'LOGOUT' });
    toast.info('Вы вышли из системы');
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
                <ListItemText primary="Выход" />
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
                <ListItemText primary="Выход" />
              </ListItemButton>
            </ListItem>
          </>
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
            m – Рекомендации
          </Typography>

          <Box sx={{ display: { xs: 'none', sm: 'block' } }}>
            {user ? (
              user.id === ADMIN_ID ? (
                <>
                  <Button color="inherit" component={Link} to="/admin">
                    Admin
                  </Button>
                  <Button color="inherit" onClick={handleLogout}>
                    Выход
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
                    Выход
                  </Button>
                </>
              )
            ) : (
              <Button color="inherit" component={Link} to="/login">
                Вход
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
