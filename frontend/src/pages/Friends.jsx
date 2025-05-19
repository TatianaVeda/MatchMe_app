// import React, { useState, useEffect } from 'react';
// import { Container, Typography, Tab, Tabs, Box,
//   Grid, Card, CardMedia, CardContent, CardActions, Button, CircularProgress
// } from '@mui/material';
// import { toast } from 'react-toastify';
// import { useNavigate } from 'react-router-dom';
// import UserCard from '../components/UserCard';
// import { getConnections, getPendingConnections, updateConnectionRequest, deleteConnection } from '../api/connections';
// import { getUser } from '../api/user';
// import { useChatState, useChatDispatch  } from '../contexts/ChatContext';

// const Friends = () => {
//   const navigate = useNavigate();
//   const { setChats } = useChatDispatch();
//   const [tab, setTab] = useState(0);
//   const [friends, setFriends] = useState([]);
//   const [pending, setPending] = useState([]);
//   const [loading, setLoading] = useState(true);
//   const { chats } = useChatState();
//   const handleChatClick = (userId) => {
//     const existingChat = chats.find(c => c.otherUserID === userId);
//     if (existingChat) {
//       navigate(`/chat/${existingChat.id}`);
//     } else {
//       navigate(`/chat/new?other_user_id=${userId}`);
//     }
//   };
//   const fetchFriends = async () => {
//     try {
//       const ids = await getConnections();
//       const data = await Promise.all(ids.map(async id => {
//         const u = await getUser(id);
//         return { id, ...u };
//       }));
//       setFriends(data);
//     } catch {
//       toast.error('Ошибка загрузки списка друзей');
//     }
//   };
//   const fetchPending = async () => {
//     try {
//       const ids = await getPendingConnections();
//       const data = await Promise.all(ids.map(async id => {
//         const u = await getUser(id);
//         return { id, ...u };
//       }));
//       setPending(data);
//     } catch {
//       toast.error('Ошибка загрузки заявок');
//     }
//   };
//   useEffect(() => {
//     setLoading(true);
//     Promise.all([fetchFriends(), fetchPending()])
//       .catch(() => {})
//       .finally(() => setLoading(false));
//   }, []);
//   const handleTabChange = (_, v) => setTab(v);
//   const handleAccept = async id => {
//     try {
//       await updateConnectionRequest(id, 'accept');
//       toast.success('Запрос принят');
//       setPending(p => p.filter(u => u.id !== id));
//       const accepted = pending.find(u => u.id === id);
//       setFriends(f => [...f, accepted]);
//     } catch {
//       toast.error('Ошибка при принятии');
//     }
//   };
//   const handleDecline = async id => {
//     try {
//       await updateConnectionRequest(id, 'decline');
//       toast.info('Запрос отклонён');
//       setPending(p => p.filter(u => u.id !== id));
//     } catch {
//       toast.error('Ошибка при отклонении');
//     }
//   };
//   const handleRemove = async id => {
//     try {
//       await deleteConnection(id);
//       toast.success('Пользователь удалён из друзей');
//       setFriends(f => f.filter(u => u.id !== id));
//       // убрать чат из списка чатов
//       setChats(chs => chs.filter(c => c.otherUserID !== id));
//       // если сейчас открыт этот чат — редиректим назад
//       if (window.location.pathname === `/chat/${id}`) {
//         navigate('/chats');
//       }
//     } catch {
//       toast.error('Ошибка при удалении');
//     }
//   };
//   if (loading) {
//     return (
//       <Container sx={{ textAlign:'center', mt:4 }}>
//         <CircularProgress />
//       </Container>
//     );
//   }
//   return (
//     <Container sx={{ mt:4 }}>
//       <Typography variant="h4" gutterBottom>Друзья</Typography>
//       <Tabs value={tab} onChange={handleTabChange} sx={{ mb:3 }}>
//         <Tab label="Мои друзья" />
//         <Tab label="Запросы" />
//       </Tabs>
// {tab === 0 && (
//         friends.length === 0 ? (
//           <Typography>У вас пока нет друзей.</Typography>
//         ) : (
//           <Grid container spacing={2}>
//             {/* {friends.map(u => (
//               <Grid key={u.id} item xs={12} sm={6} md={4}>
//                 <UserCard
//                   user={{ ...u, connected: true }}
//                   showChat={true}
//                   onChatClick={() => navigate(`/chat/${u.id}`)}
//                   onClick={() => navigate(`/users/${u.id}`)}
//                 /> */}

// {friends.map(u => (
//     <Grid key={u.id} item xs={12} sm={6} md={4}>
//       <UserCard
//         user={{ ...u, connected: true }}
//         showChat={true}
//         // onChatClick={() => navigate(`/chat/${u.id}`)}
//         onChatClick={() => handleChatClick(u.id)}
//         onClick={() => navigate(`/users/${u.id}`)}
//       />
//       {/* новая кнопка удаления */}
//       <Box sx={{ display: 'flex', justifyContent: 'center', mt: 1 }}>
//         <Button
//           size="small"
//           variant="outlined"
//           color="error"
//           onClick={() => handleRemove(u.id)}
//         >
//           Удалить из друзей
//         </Button>
//       </Box>
//               </Grid>
//             ))}
//           </Grid>
//         )
//       )}
// {tab === 1 && (
//         pending.length === 0 ? (
//           <Typography>Нет входящих запросов.</Typography>
//         ) : (
//           <Grid container spacing={2}>
//             {pending.map(u => (
//               <Grid key={u.id} item xs={12} sm={6} md={4}>
//                 <UserCard
//                   user={{ ...u, connected: false }}
//                   showChat={false}
//                   onClick={() => navigate(`/users/${u.id}`)}
//                 />
//                 <Box sx={{ display: 'flex', justifyContent: 'center', mt: 1 }}>
//                   <Button
//                     size="small"
//                     variant="contained"
//                     onClick={() => handleAccept(u.id)}
//                     sx={{ mr: 1 }}
//                   >
//                     Принять
//                   </Button>
//                   <Button
//                     size="small"
//                     variant="outlined"
//                     onClick={() => handleDecline(u.id)}
//                   >
//                     Отклонить
//                   </Button>
//                 </Box>
//               </Grid>
//             ))}
//           </Grid>
//         )
//       )}
//     </Container>
//   );
// };
// export default Friends;

import React, { useState, useEffect } from 'react';
import {
  Container, Typography, Tab, Tabs, Box,
  Grid, Button, CircularProgress
} from '@mui/material';
import { toast } from 'react-toastify';
import { useNavigate } from 'react-router-dom';
import UserCard from '../components/UserCard';
import {
  getConnections,
  getPendingConnections,    // incoming
  getSentConnections,       // outgoing
  updateConnectionRequest,
  deleteConnection
} from '../api/connections';
import { getUser } from '../api/user';
import { useChatState, useChatDispatch } from '../contexts/ChatContext';

const Friends = () => {
  const navigate = useNavigate();
  const { setChats } = useChatDispatch();
  const { chats } = useChatState();

  const [tab, setTab] = useState(0);
  const [friends, setFriends] = useState([]);
  const [incoming, setIncoming] = useState([]);
  const [outgoing, setOutgoing] = useState([]);
  const [loading, setLoading] = useState(true);

  // navigate or create chat
  const handleChatClick = (userId) => {
    const existing = chats.find(c => c.otherUserID === userId);
    navigate(existing ? `/chat/${existing.id}` : `/chat/new?other_user_id=${userId}`);
  };

  // load full user objects given an array of IDs
  const loadUsers = async (ids) => {
    return Promise.all(ids.map(async id => {
      const u = await getUser(id);
      return { id, ...u };
    }));
  };

  const fetchFriends = async () => {
    const ids = await getConnections();
    setFriends(await loadUsers(ids));
  };

  const fetchIncoming = async () => {
    const ids = await getPendingConnections();
    setIncoming(await loadUsers(ids));
  };

  const fetchOutgoing = async () => {
    const ids = await getSentConnections();
    setOutgoing(await loadUsers(ids));
  };

  useEffect(() => {
    setLoading(true);
    Promise.all([fetchFriends(), fetchIncoming(), fetchOutgoing()])
      .catch(() => toast.error('Ошибка загрузки данных'))
      .finally(() => setLoading(false));
  }, []);

  const handleAccept = async (id) => {
    try {
      await updateConnectionRequest(id, 'accept');
      toast.success('Запрос принят');
  
      // Найдём пользователя в текущем списке входящих сразу
      const acceptedUser = incoming.find(u => u.id === id);
      
      // Убираем его из incoming
      setIncoming(prevIncoming =>
        prevIncoming.filter(u => u.id !== id)
      );
      
      // Добавляем в друзья
      setFriends(prevFriends => [
        ...prevFriends,
        acceptedUser
      ]);
    } catch {
      toast.error('Ошибка при принятии');
    }
  };
  

  const handleDecline = async (id) => {
    try {
      await updateConnectionRequest(id, 'decline');
      toast.info('Запрос отклонён');
  
      setIncoming(prevIncoming =>
        prevIncoming.filter(u => u.id !== id)
      );
    } catch {
      toast.error('Ошибка при отклонении');
    }
  };
  

  const handleRemove = async (id) => {
    await deleteConnection(id);
    toast.success('Пользователь удалён из друзей');
    setFriends(f => f.filter(u => u.id !== id));
    setChats(chs => chs.filter(c => c.otherUserID !== id));
    if (window.location.pathname === `/chat/${id}`) navigate('/chats');
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
      <Typography variant="h4" gutterBottom>Друзья</Typography>

      <Tabs value={tab} onChange={(_, v) => setTab(v)} sx={{ mb: 3 }}>
        <Tab label="Мои друзья" />
        <Tab label="Запросы" />
      </Tabs>

      {tab === 0 && (
        friends.length === 0
          ? <Typography>У вас пока нет друзей.</Typography>
          : (
            <Grid container spacing={2}>
              {friends.map(u => (
                <Grid key={u.id} item xs={12} sm={6} md={4}>
                  <UserCard
                    user={{ ...u, connected: true }}
                    showChat
                    onChatClick={() => handleChatClick(u.id)}
                    onClick={() => navigate(`/users/${u.id}`)}
                  />
                  <Box sx={{ display: 'flex', justifyContent: 'center', mt: 1 }}>
                    <Button
                      size="small"
                      variant="outlined"
                      color="error"
                      onClick={() => handleRemove(u.id)}
                    >
                      Удалить из друзей
                    </Button>
                  </Box>
                </Grid>
              ))}
            </Grid>
          )
      )}

      {tab === 1 && (
        incoming.length === 0 && outgoing.length === 0
          ? <Typography>Нет запросов.</Typography>
          : (
            <>
              {/* INCOMING */}
              {incoming.length > 0 && (
                <>
                  <Typography variant="h6">Входящие запросы</Typography>
                  <Grid container spacing={2} sx={{ mb: 4 }}>
                    {incoming.map(u => (
                      <Grid key={u.id} item xs={12} sm={6} md={4}>
                        <UserCard
                          user={{ ...u, connected: false }}
                          showChat={false}
                          onClick={() => navigate(`/users/${u.id}`)}
                        />
                        <Box sx={{ display: 'flex', justifyContent: 'center', mt: 1 }}>
                          <Button size="small" variant="contained" onClick={() => handleAccept(u.id)} sx={{ mr: 1 }}>
                            Принять
                          </Button>
                          <Button size="small" variant="outlined" onClick={() => handleDecline(u.id)}>
                            Отклонить
                          </Button>
                        </Box>
                      </Grid>
                    ))}
                  </Grid>
                </>
              )}

              {/* OUTGOING */}
              {outgoing.length > 0 && (
                <>
                  <Typography variant="h6">Исходящие запросы</Typography>
                  <Grid container spacing={2}>
                    {outgoing.map(u => (
                      <Grid key={u.id} item xs={12} sm={6} md={4}>
                        <UserCard
                          user={{ ...u, connected: false }}
                          showChat={false}
                          onClick={() => navigate(`/users/${u.id}`)}
                        />
                        <Box sx={{ display: 'flex', justifyContent: 'center', mt: 1 }}>
                          <Button size="small" variant="outlined" disabled>
                            Запрос отправлен
                          </Button>
                        </Box>
                      </Grid>
                    ))}
                  </Grid>
                </>
              )}
            </>
          )
      )}
    </Container>
  );
};

export default Friends;
