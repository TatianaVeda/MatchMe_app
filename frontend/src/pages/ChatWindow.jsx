// // frontend/src/pages/ChatWindow.jsx
// import React, { useState, useEffect, useRef, Fragment } from 'react';
// import { useParams, useNavigate, useLocation } from 'react-router-dom';
// import {
//   Container, Box, Typography, TextField, Button,
//   CircularProgress, List, Divider
// } from '@mui/material';
// import api from '../api/index';
// import { toast } from 'react-toastify';
// import { useChatState, useChatDispatch } from '../contexts/ChatContext';
// import ChatBubble from '../components/ChatBubble';
// import { useAuthState } from '../contexts/AuthContext';
// import useWebSocket from '../hooks/useWebSocket';

// const ChatWindow = () => {
//   const { user } = useAuthState();
//   const { chatId } = useParams();
//   const navigate = useNavigate();
//   const location = useLocation();

//   const { messages: allMessages, typingStatuses } = useChatState();
//   const { setMessages, sendMessage, sendTyping } = useChatDispatch();
//   const { subscribe, unsubscribe } = useWebSocket();

//   const messagesEndRef = useRef(null);
//   const [page, setPage] = useState(1);
//   const [loading, setLoading] = useState(true);
//   const [newMessage, setNewMessage] = useState('');
//   // const [hasMore, setHasMore] = useState(true);
//   // const chatContainerRef = useRef(null);
//   // const scrollToBottomRef = useRef(true); 

//   // локальные сообщения для текущего чата
//     // стало
//     const raw = allMessages[chatId];
//     // если raw не массив — превращаем в пустой
//     const messages = Array.isArray(raw) ? raw : [];
  
//   // const messages = allMessages[chatId] || [];
//   const isTyping = typingStatuses[chatId];

//   // Если chatId === 'new', создаём чат через POST и сразу редиректим на реальный ID
//   // useEffect(() => {
//   //   if (chatId === 'new') {
//   //     const params = new URLSearchParams(location.search);
//   //     const otherUserId = params.get('other_user_id');
//   //     if (!otherUserId) {
//   //       toast.error('Не указан другой пользователь');
//   //       return;
//   //     }
//   //     api.post('/chats', { otherUserId })
//   //       .then(({ data }) => {
//   //         // data = { chatId: number }
//   //         navigate(`/chat/${data.chatId}`, { replace: true });
//   //       })
//   //       .catch(() => {
//   //         toast.error('Не удалось создать чат');
//   //       });
//   //   }
//   // }, [chatId, location.search, navigate]);

//   useEffect(() => {
//     // если пока новый чат — создаём и редиректим
//     if (chatId === 'new') {
//       const otherUserID = new URLSearchParams(location.search).get('other_user_id');
//       if (!otherUserID) {
//         toast.error('Не указан other_user_id');
//         return;
//       }
//       api.post('/chats', { otherUserId: otherUserID })
//         .then(({ data }) => {
//           // после создания меняем URL на реальный chatId
//           navigate(`/chat/${data.chatId}`, { replace: true });
//         })
//         .catch(() => {
//           toast.error('Не удалось открыть чат');
//         })
//         .finally(() => {
//           setLoading(false);
//         });
//     } else {
//       // для существующего — грузим историю
//       setLoading(true);
//       fetchMessages(page);
//     }
//   }, [chatId, page, location.search, navigate]);

//   // Загрузка истории, но только если уже числовой chatId
//   const fetchMessages = async (p = 1) => {
//     try {
//       const { data } = await api.get(`/chats/${chatId}`, {
//         params: { page: p, limit: 20 }
//       });
//       setMessages(chatId, p === 1 ? data : [...data, ...messages]);
//     } catch {
//       toast.error('Ошибка загрузки сообщений');
//     } finally {
//       setLoading(false);
//     }
//   };

//   //pagination, no anchor
//   // const fetchMessages = async (p = 1) => {
//   //   try {
//   //     const { data } = await api.get(`/chats/${chatId}`, {
//   //       params: { page: p, limit: 20 }
//   //     });
  
//   //     if (p === 1) {
//   //       setMessages(chatId, data);
//   //     } else {
//   //       setMessages(chatId, [...data, ...messages]);
//   //     }
  
//   //     // Optional: if you want to detect "no more to load"
//   //     if (data.length < 20) {
//   //       setHasMore(false); // you can define this state
//   //     }
  
//   //   } catch (err) {
//   //     console.error(err);
//   //     toast.error("Ошибка загрузки сообщений");
//   //   } finally {
//   //     setLoading(false);
//   //   }
//   // };  

//   // const fetchMessages = async (p = 1) => {
//   //   if (!chatContainerRef.current) return;
  
//   //   const container = chatContainerRef.current;
//   //   const prevScrollHeight = container.scrollHeight;
  
//   //   try {
//   //     const { data } = await api.get(`/chats/${chatId}`, {
//   //       params: { page: p, limit: 20 }
//   //     });
  
//   //     if (p === 1) {
//   //       setMessages(chatId, data);
//   //       //scroll
//   //       scrollToBottomRef.current = true;
//   //     } else {
//   //       setMessages(chatId, [...data, ...messages]);
  
//   //       // Delay scroll adjustment until after DOM updates
//   //       setTimeout(() => {
//   //         const newScrollHeight = container.scrollHeight;
//   //         const diff = newScrollHeight - prevScrollHeight;
//   //         container.scrollTop += diff;
//   //       }, 0);
//   //     }
  
//   //     if (data.length < 20) {
//   //       setHasMore(false);
//   //     }
//   //   } catch (err) {
//   //     console.error(err);
//   //     toast.error("Ошибка загрузки сообщений");
//   //   } finally {
//   //     setLoading(false);
//   //   }
//   // };
  

//   // // При смене chatId или page — грузим истории (если не "new")
//   // useEffect(() => {
//   //   if (chatId !== 'new') {
//   //     setLoading(true);
//   //     fetchMessages(page);
//   //   }
//   // }, [chatId, page]);

//   // Подписываемся на WS только по реальным chatId
//   useEffect(() => {
//     if (chatId && chatId !== 'new') {
//       subscribe(chatId);
//       return () => unsubscribe(chatId);
//     }
//   }, [chatId, subscribe, unsubscribe]);

//   // Автоскролл вниз по новому сообщению
//   useEffect(() => {
//     messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
//   }, [messages]);
// //pagination
//   // useEffect(() => {
//   //   if (scrollToBottomRef.current) {
//   //     messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
//   //   }
//   // }, [messages]);
  

//   // Отправка сообщения
//   const handleSend = async (e) => {
//     e.preventDefault();
//     const content = newMessage.trim();
//     if (!content) return;
//     try {
//       const { data: saved } = await api.post(
//         `/chats/${chatId}/messages`,
//         { content }
//       );

//       // 2) Формируем единый для GET /chats/{id} объект
//     const outgoing = {
//       id: saved.id,
//       content: saved.content,
//       timestamp: saved.timestamp,
//       read: saved.read,
//       // В GET-контроллере это поле называется sender_id
//       sender_id: saved.senderId,
//       // и там же используются sender_name
//       sender_name: `${saved.sender.profile.firstName} ${saved.sender.profile.lastName}`.trim(),
//     };




//       sendMessage(chatId, content);            // через WS
//       setMessages(chatId, [...messages, saved]); // локально
//       //scroll
//       //scrollToBottomRef.current = true;
//       //setMessages(chatId, [...messages, outgoing]);
//       setNewMessage('');
//     } catch {
//       toast.error('Ошибка отправки сообщения');
//     }
//   };

//   // При вводе — шлём typing=true, а через 1.5 с таймаутом «прибиваем»
//   const handleChange = (e) => {
//     setNewMessage(e.target.value);
//     sendTyping(chatId, true);
//     clearTimeout(window.typingTimeout);
//     window.typingTimeout = setTimeout(() => {
//       sendTyping(chatId, false);
//     }, 1500);
//   };

//   return (
//     <Container sx={{ mt: 4 }}>
//       <Typography variant="h4" gutterBottom>
//         {chatId === 'new' ? 'Создание чата...' : `Чат ${chatId}`}
//       </Typography>

//       {/* индикатор «печатает» */}
//       {isTyping && (
//         <Typography variant="body2" color="textSecondary" sx={{ mb: 1, fontStyle: 'italic' }}>
//           Пользователь печатает...
//         </Typography>
//       )}

//       <Box sx={{
//         border: '1px solid #ccc',
//         borderRadius: 2,
//         height: '60vh',
//         overflowY: 'auto',
//         p: 2,
//         mb: 2
//       }}>
//         {/* <Box
//   ref={chatContainerRef}
//   sx={{
//     border: '1px solid #ccc',
//     borderRadius: 2,
//     height: '60vh',
//     overflowY: 'auto',
//     p: 2,
//     mb: 2
//   }}
// > */}

//         {loading || chatId === 'new' ? (
//           <Box sx={{ textAlign: 'center', mt: 2 }}>
//             <CircularProgress />
//           </Box>
//         ) : (
//           <>
//             {page > 1 && (
//               <Box sx={{ textAlign: 'center', mb: 1 }}>
//                 <Button onClick={() => setPage(p => p + 1)}>
//                   Загрузить ещё
//                 </Button>
//               </Box>
//             )}
//             {/* //pagination */}
//             {/* {hasMore && (
//               <Box sx={{ textAlign: "center", mb: 1 }}>
//                 <Button onClick={() => {
//   scrollToBottomRef.current = false;   // ⬅️ Prevent auto-scroll
//   setPage((prev) => prev + 1);
// }}>
//   Загрузить ещё
// </Button>

//               </Box>
//             )} */}

//             {/* <List>
//               {/* {messages.map(msg => ( */}
//               {/* {messages
//                .filter(msg => msg !== null)
//                .map(msg => (
//                 <Fragment key={msg.id}>
//                   <ChatBubble
//                     message={msg}
//                     isOwn={msg.sender_id === user.id}
//                   />
//                   <Divider component="li" />
//                 </Fragment>
//               ))}
//             </List> */}             <List>
//                          {/*
//                            Отфильтровываем любые null/undefined,
//                            и только потом делаем map
//                          */}
//                          {messages
//                            .filter(msg => msg != null)
//                            .map(msg => (
//                              <Fragment key={msg.id}>
//                                <ChatBubble
//                                  message={msg}
//                                  isOwn={msg.sender_id === user.id}
//                                />
//                                <Divider component="li" />
//                              </Fragment>
//                            ))
//                         }
//                        </List>
//             <div ref={messagesEndRef} />
//           </>
//         )}
//       </Box>

//       {/* если чат уже реальный — показываем форму */}
//       {chatId !== 'new' && (
//         <Box
//           component="form"
//           onSubmit={handleSend}
//           sx={{ display: 'flex', gap: 1 }}
//         >
//           <TextField
//             label="Новое сообщение"
//             value={newMessage}
//             onChange={handleChange}
//             fullWidth
//             multiline
//             rows={2}
//           />
//           <Button variant="contained" type="submit">
//             Отправить
//           </Button>
//         </Box>
//       )}
//     </Container>
//   );
// };

// export default ChatWindow;


import React, { useState, useEffect, useRef, Fragment } from 'react';
import { useParams, useNavigate, useLocation } from 'react-router-dom';
import {
  Container, Box, Typography, TextField, Button,
  CircularProgress, List, Divider
} from '@mui/material';
import api from '../api/index';
import { toast } from 'react-toastify';
import { useChatState, useChatDispatch } from '../contexts/ChatContext';
import ChatBubble from '../components/ChatBubble';
import { useAuthState } from '../contexts/AuthContext';
import useWebSocket from '../hooks/useWebSocket';
const ChatWindow = () => {
  const { user } = useAuthState();
  const { chatId } = useParams();
  const navigate = useNavigate();
  const location = useLocation();
  const { messages: allMessages, typingStatuses } = useChatState();
  const { setMessages, sendMessage, sendTyping } = useChatDispatch();
  const { subscribe, unsubscribe } = useWebSocket();
  const messagesEndRef = useRef(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [newMessage, setNewMessage] = useState('');
  const messages = allMessages[chatId] || [];
  const isTyping = typingStatuses[chatId];
  useEffect(() => {
    if (chatId === 'new') {
      const otherUserID = new URLSearchParams(location.search).get('other_user_id');
      if (!otherUserID) {
        toast.error('Не указан other_user_id');
        return;
      }
      api.post('/chats', { otherUserId: otherUserID })
        .then(({ data }) => {
          navigate(`/chat/${data.chatId}`, { replace: true });
        })
        .catch(() => {
          toast.error('Не удалось открыть чат');
        })
        .finally(() => {
          setLoading(false);
        });
    } else {
      setLoading(true);
      fetchMessages(page);
    }
  }, [chatId, page, location.search, navigate]);
  const fetchMessages = async (p = 1) => {
    try {
      const { data } = await api.get(`/chats/${chatId}`, {
        params: { page: p, limit: 20 }
      });
      setMessages(chatId, p === 1 ? data : [...data, ...messages]);
    } catch {
      toast.error('Ошибка загрузки сообщений');
    } finally {
      setLoading(false);
    }
  };
  useEffect(() => {
    if (chatId && chatId !== 'new') {
      subscribe(chatId);
      return () => unsubscribe(chatId);
    }
  }, [chatId, subscribe, unsubscribe]);
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);
  const handleSend = async (e) => {
    e.preventDefault();
    const content = newMessage.trim();
    if (!content) return;
    try {
      const { data: saved } = await api.post(
        `/chats/${chatId}/messages`,
        { content }
      );
      sendMessage(chatId, content);            // через WS
      setMessages(chatId, [...messages, saved]); // локально
      setNewMessage('');
    } catch {
      toast.error('Ошибка отправки сообщения');
    }
  };
  const handleChange = (e) => {
    setNewMessage(e.target.value);
    sendTyping(chatId, true);
    clearTimeout(window.typingTimeout);
    window.typingTimeout = setTimeout(() => {
      sendTyping(chatId, false);
    }, 1500);
  };
  return (
    <Container sx={{ mt: 4 }}>
      <Typography variant="h4" gutterBottom>
        {chatId === 'new' ? 'Создание чата...' : `Чат ${chatId}`}
      </Typography>
      {isTyping && (
        <Typography variant="body2" color="textSecondary" sx={{ mb: 1, fontStyle: 'italic' }}>
          Пользователь печатает...
        </Typography>
      )}
      <Box sx={{
        border: '1px solid #ccc',
        borderRadius: 2,
        height: '60vh',
        overflowY: 'auto',
        p: 2,
        mb: 2
      }}>
        {loading || chatId === 'new' ? (
          <Box sx={{ textAlign: 'center', mt: 2 }}>
            <CircularProgress />
          </Box>
        ) : (
          <>
            {page > 1 && (
              <Box sx={{ textAlign: 'center', mb: 1 }}>
                <Button onClick={() => setPage(p => p + 1)}>
                  Загрузить ещё
                </Button>
              </Box>
            )}
            <List>
              {messages.map(msg => (
                <Fragment key={msg.id}>
                  <ChatBubble
                    message={msg}
                    isOwn={msg.sender_id === user.id}
                  />
                  <Divider component="li" />
                </Fragment>
              ))}
            </List>
            <div ref={messagesEndRef} />
          </>
        )}
      </Box>
      {chatId !== 'new' && (
        <Box
          component="form"
          onSubmit={handleSend}
          sx={{ display: 'flex', gap: 1 }}
        >
          <TextField
            label="Новое сообщение"
            value={newMessage}
            onChange={handleChange}
            fullWidth
            multiline
            rows={2}
          />
          <Button variant="contained" type="submit">
            Отправить
          </Button>
        </Box>
      )}
    </Container>
  );
};
export default ChatWindow;