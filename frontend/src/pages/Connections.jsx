// // import React, { useState, useEffect } from 'react';
// // import { Container, Typography, Grid, Card, CardContent, CardMedia, CardActions, Button, CircularProgress } from '@mui/material';
// // import { toast } from 'react-toastify';
// // import { useNavigate } from 'react-router-dom';
// // import UserCard from '../components/UserCard';
// // import { getPendingConnections, updateConnectionRequest, getConnections, deleteConnection } from '../api/connections';
// // import { getUser } from '../api/user';
// // const Connections = () => {
// //   const [pending, setPending] = useState([]);
// //   const [connections, setConnections] = useState([]);
// //   const [loading, setLoading] = useState(true);
// //   const navigate = useNavigate();
// //   const fetchAll = async () => {
// //     setLoading(true);
// //     try {
// //       const pendingIds = await getPendingConnections();
// //       const pendingDetails = await Promise.all(
// //         pendingIds.map(async (id) => {
// //           try {
// //             const userData = await getUser(id);
// //             return { id, ...userData };
// //           } catch (err) {
// //             console.error('Ошибка загрузки данных pending для id', id, err);
// //             return null;
// //           }
// //         })
// //       );
// //       setPending(pendingDetails.filter((u) => u !== null));
// //       const acceptedIds = await getConnections();
// //       const acceptedDetails = await Promise.all(
// //         acceptedIds.map(async (id) => {
// //           try {
// //             const userData = await getUser(id);
// //             return { id, ...userData };
// //           } catch (err) {
// //             console.error('Ошибка загрузки данных accepted для id', id, err);
// //             return null;
// //           }
// //         })
// //       );
// //       setConnections(acceptedDetails.filter((u) => u !== null));
// //     } catch (err) {
// //       toast.error('Ошибка загрузки подключений');
// //     } finally {
// //       setLoading(false);
// //     }
// //   };
// //   useEffect(() => {
// //     fetchAll();
// //   }, []);
// //   const handleAccept = async (id) => {
// //     try {
// //       await updateConnectionRequest(id, 'accept');
// //       toast.success('Запрос принят');
// //       const acceptedUser = pending.find((u) => u.id === id);
// //       setConnections((prev) => [...prev, acceptedUser]);
// //       setPending((prev) => prev.filter((u) => u.id !== id));
// //     } catch {
// //       toast.error('Ошибка при принятии запроса');
// //     }
// //   };
// //   const handleDeclinePending = async (id) => {
// //     try {
// //       await updateConnectionRequest(id, 'decline');
// //       toast.info('Запрос отклонён');
// //       setPending((prev) => prev.filter((u) => u.id !== id));
// //     } catch {
// //       toast.error('Ошибка при отклонении запроса');
// //     }
// //   };
// //   const handleDisconnect = async (id) => {
// //     try {
// //       await deleteConnection(id);
// //       toast.success('Отключение выполнено успешно');
// //       setConnections((prev) => prev.filter((conn) => conn.id !== id));
// //     } catch {
// //       toast.error('Ошибка при отключении');
// //     }
// //   };
// //   if (loading) {
// //     return (
// //       <Container sx={{ textAlign: 'center', mt: 4 }}>
// //         <CircularProgress />
// //       </Container>
// //     );
// //   }
// //   return (
// //     <Container sx={{ mt: 4 }}>
// //       {/* Входящие запросы */}
// //       <Typography variant="h4" gutterBottom>
// //         Запросы на подключение
// //       </Typography>
// //       {pending.length === 0 ? (
// //         <Typography sx={{ mb: 4 }}>Нет входящих запросов.</Typography>
// //       ) : (
// //         <Grid container spacing={3} sx={{ mb: 4 }}>
// //           {pending.map(user => (
// //             <Grid key={user.id} item xs={12} sm={6} md={4}>
// //               <UserCard
// //                 user={{ ...user, connected: false }}
// //                 showChat={false}
// //                 onClick={() => navigate(`/users/${user.id}`)}
// //               />
// //               <Grid container spacing={1} justifyContent="center" sx={{ mt: 1 }}>
// //                 <Grid item>
// //                   <Button size="small" variant="contained" onClick={() => handleAccept(user.id)}>
// //                     Принять
// //                   </Button>
// //                 </Grid>
// //                 <Grid item>
// //                   <Button size="small" variant="outlined" onClick={() => handleDeclinePending(user.id)}>
// //                     Отклонить
// //                   </Button>
// //                 </Grid>
// //               </Grid>
// //             </Grid>
// //           ))}
// //         </Grid>
// //       )}
// //       <Typography variant="h4" gutterBottom>
// //         Подключения
// //       </Typography>
// //       {connections.length === 0 ? (
// //         <Typography>Нет подключённых профилей.</Typography>
// //       ) : (
// //         <Grid container spacing={3}>
// //           {connections.map(conn => (
// //             <Grid key={conn.id} item xs={12} sm={6} md={4}>
// //               <UserCard
// //                 user={{ ...conn, connected: true }}
// //                 showChat={true}
// //                 onChatClick={() => navigate(`/chat/${conn.id}`)}
// //                 onClick={() => navigate(`/users/${conn.id}`)}
// //               />
// //               <Grid container justifyContent="center" sx={{ mt: 1 }}>
// //                 <Button
// //                   variant="outlined"
// //                   color="error"
// //                   onClick={() => handleDisconnect(conn.id)}
// //                 >
// //                   Отключить
// //                 </Button>
// //               </Grid>
// //             </Grid>
// //           ))}
// //         </Grid>
// //       )}
// //     </Container>
// //   );
// // };
// // export default Connections;

import React, { useState, useEffect } from 'react';
import {
   Container, Typography, Tab, Tabs, Box, Grid, Button, CircularProgress 
} from '@mui/material';
import { toast } from 'react-toastify';
import { useNavigate } from 'react-router-dom';
import UserCard from '../components/UserCard';
import { getConnections, getPendingConnections, updateConnectionRequest, deleteConnection
} from '../api/connections';
import { getUser } from '../api/user';
import { useChatState, useChatDispatch } from '../contexts/ChatContext';

const Connections = () => {
  const navigate = useNavigate();
  const { setChats } = useChatDispatch();
  const { chats } = useChatState();

  const [tab, setTab] = useState(0);
  const [connections, setConnections] = useState([]);  // accepted
  const [pending, setPending] = useState([]);        // incoming
  const [loading, setLoading] = useState(true);

  const handleChatClick = (userId) => {
    const existing = chats.find(c => c.otherUserID === userId);
    navigate(existing ? `/chat/${existing.id}` : `/chat/new?other_user_id=${userId}`);
  };

  const loadUsers = async (ids) => {
    const raw = await Promise.all(ids.map(async id => {
      const u = await getUser(id);
      return { id, ...u };
    }));
    return fetchOnlineStatusForUsers(raw);
  };

  const fetchOnlineStatusForUsers = async (users) => {
    const updated = await Promise.all(users.map(async u => {
      try {
        const res = await fetch(`/api/user/online?user_id=${u.id}`);
        const data = await res.json();
        return { ...u, online: data.online };
      } catch (err) {
        console.error(`Error fetching status for user ${u.id}:`, err);
        return { ...u, online: false };
      }
    }));
    return updated;
  };

  const fetchData = async () => {
    setLoading(true);
    try {
      const pendingIds = await getPendingConnections();
      setPending(await loadUsers(pendingIds));

      const connIds = await getConnections();
      setConnections(await loadUsers(connIds));
    } catch {
      toast.error('Ошибка загрузки подключений');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 120000);
    return () => clearInterval(interval);
  }, []);

  const handleAccept = async (id) => {
    try {
      await updateConnectionRequest(id, 'accept');
      toast.success('Запрос принят');
      const user = pending.find(u => u.id === id);
      setPending(p => p.filter(u => u.id !== id));
      setConnections(c => [...c, user]);
    } catch {
      toast.error('Ошибка при принятии');
    }
  };

  const handleDecline = async (id) => {
    try {
      await updateConnectionRequest(id, 'decline');
      toast.info('Запрос отклонён');
      setPending(p => p.filter(u => u.id !== id));
    } catch {
      toast.error('Ошибка при отклонении');
    }
  };

  const handleDisconnect = async (id) => {
    try {
      await deleteConnection(id);
      toast.success('Подключение удалено');
      setConnections(c => c.filter(u => u.id !== id));
      setChats(chs => chs.filter(c => c.otherUserID !== id));
      if (window.location.pathname === `/chat/${id}`) navigate('/chats');
    } catch {
      toast.error('Ошибка при отключении');
    }
  };

  if (loading) {
    return (
      <Container sx={{ textAlign: 'center', mt: 4 }}>
        <CircularProgress />
      </Container>
    );
  }

  return (
    <Container sx={{ mt: 4 }}>
      <Typography variant="h4" gutterBottom>Подключения</Typography>
      <Tabs value={tab} onChange={(e, v) => setTab(v)} sx={{ mb: 3 }}>
        <Tab label="Существующие" />
        <Tab label="Запросы" />
      </Tabs>

      {tab === 0 && (
        connections.length === 0
          ? <Typography>Нет подключённых профилей.</Typography>
          : (
            <Grid container spacing={2}>
              {connections.map(u => (
                <Grid key={u.id} item xs={12} sm={6} md={4}>
                  <UserCard
                    user={{ ...u, connected: true }}
                    showChat
                    onChatClick={() => handleChatClick(u.id)}
                    onClick={() => navigate(`/users/${u.id}`)}
                  />
                  <Box sx={{ display: 'flex', justifyContent: 'center', mt: 1 }}>
                    <Button variant="outlined" color="error" size="small" onClick={() => handleDisconnect(u.id)}>
                      Отключить
                    </Button>
                  </Box>
                </Grid>
              ))}
            </Grid>
          )
      )}

      {tab === 1 && (
        pending.length === 0
          ? <Typography>Нет входящих запросов.</Typography>
          : (
            <Grid container spacing={2}>
              {pending.map(u => (
                <Grid key={u.id} item xs={12} sm={6} md={4}>
                  <UserCard
                    user={{ ...u, connected: false }}
                    showChat={false}
                    onClick={() => navigate(`/users/${u.id}`)}
                  />
                  <Box sx={{ display: 'flex', justifyContent: 'center', mt: 1 }}>
                    <Button size="small" variant="contained" sx={{ mr: 1 }} onClick={() => handleAccept(u.id)}>
                      Принять
                    </Button>
                    <Button size="small" variant="outlined" onClick={() => handleDecline(u.id)}>
                      Отклонить
                    </Button>
                  </Box>
                </Grid>
              ))}
            </Grid>
          )
      )}
    </Container>
  );
};

export default Connections;
