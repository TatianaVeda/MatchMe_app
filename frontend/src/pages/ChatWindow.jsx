// import React, { useState, useEffect, useRef } from 'react';
// import { useParams } from 'react-router-dom';
// import {
//   Container,
//   Box,
//   Typography,
//   TextField,
//   Button,
//   CircularProgress,
//   List,
//   Divider
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
//   const { messages: allMessages } = useChatState();
//   const { setMessages } = useChatDispatch();
//   const { subscribe, unsubscribe, sendMessage, sendTyping } = useWebSocket();

//   const messagesEndRef = useRef(null);
//   const [page, setPage] = useState(1);
//   const [loading, setLoading] = useState(true);
//   const [newMessage, setNewMessage] = useState('');

//   const messages = allMessages[chatId] || [];

//   const fetchMessages = async (p = 1) => {
//     try {
//       const { data } = await api.get(`/chats/${chatId}`, {
//         params: { page: p, limit: 20 }
//       });
//       setMessages(chatId, p === 1 ? data : [...data, ...messages]);
//     } catch {
//       toast.error("Ошибка загрузки сообщений");
//     } finally {
//       setLoading(false);
//     }
//   };

//   useEffect(() => {
//     setLoading(true);
//     fetchMessages(page);
//   }, [chatId, page]);

//   // подписываемся на WS-канал этого чата
//   useEffect(() => {
//       if (!chatId) return;
//       subscribe(chatId);
//       return () => {
//         unsubscribe(chatId);
//       };
//     }, [chatId, subscribe, unsubscribe]);

//   useEffect(() => {
//     messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
//   }, [messages]);

//   const handleSend = async (e) => {
//     e.preventDefault();
//     if (!newMessage.trim()) return;

//     try {
//       const { data: saved } = await api.post(`/chats/${chatId}/messages`, { content: newMessage });
//       sendMessage(chatId, newMessage);
//       setMessages(chatId, [...messages, saved]);
//       setNewMessage('');
//     } catch {
//       toast.error("Ошибка отправки сообщения");
//     }
//   };

//   const handleChange = (e) => {
//     setNewMessage(e.target.value);
//     sendTyping(chatId, true);
//   };

//   return (
//     <Container sx={{ mt: 4 }}>
//       <Typography variant="h4" gutterBottom>Чат {chatId}</Typography>
//       <Box sx={{
//         border: '1px solid #ccc',
//         borderRadius: 2,
//         height: '60vh',
//         overflowY: 'auto',
//         p: 2,
//         mb: 2
//       }}>
//         {loading ? (
//           <Box sx={{ textAlign: 'center', mt: 2 }}>
//             <CircularProgress />
//           </Box>
//         ) : (
//           <>
//             {page > 1 && (
//               <Box sx={{ textAlign: 'center', mb: 1 }}>
//                 <Button onClick={() => setPage(p => p + 1)}>Загрузить ещё</Button>
//               </Box>
//             )}
//             <List>
//             {messages.map(msg => (
//                <React.Fragment key={msg.id}>
//                  <ChatBubble
//                   message={msg}
//                   isOwn={msg.sender_id === user.id}
//                  />
//                  <Divider component="li" />
//                </React.Fragment>
//              ))}
//             </List>
//             <div ref={messagesEndRef} />
//           </>
//         )}
//       </Box>
//       <Box component="form" onSubmit={handleSend} sx={{ display: 'flex', gap: 1 }}>
//         <TextField
//           label="Новое сообщение"
//           value={newMessage}
//           onChange={handleChange}
//           fullWidth
//           multiline
//           rows={2}
//         />
//         <Button variant="contained" type="submit">Отправить</Button>
//       </Box>
//     </Container>
//   );
// };

// export default ChatWindow;


import React, { useState, useEffect, useRef, Fragment } from 'react';
import { useParams } from 'react-router-dom';
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
  const { messages: allMessages, typingStatuses } = useChatState();
  const { setMessages, sendMessage, sendTyping } = useChatDispatch();
  const { subscribe, unsubscribe } = useWebSocket();

  const messagesEndRef = useRef(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [newMessage, setNewMessage] = useState('');

  // локальная копия списка сообщений для текущего чата
  const messages = allMessages[chatId] || [];
  // статус набора текста для текущего чата
  const isTyping = typingStatuses[chatId];

  // Загрузка истории сообщений
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

  // Перезагрузка при смене чата или страницы
  useEffect(() => {
    setLoading(true);
    fetchMessages(page);
  }, [chatId, page]);

  // Подписка и отписка по WebSocket
  useEffect(() => {
    if (!chatId) return;
    subscribe(chatId);
    return () => unsubscribe(chatId);
  }, [chatId, subscribe, unsubscribe]);

  // Автоскролл к последнему сообщению
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  // Отправка нового сообщения
  const handleSend = async (e) => {
    e.preventDefault();
    const content = newMessage.trim();
    if (!content) return;
    try {
      const { data: saved } = await api.post(
        `/chats/${chatId}/messages`,
        { content }
      );
      sendMessage(chatId, content); // через WS
      setMessages(chatId, [...messages, saved]); // локально
      setNewMessage('');
    } catch {
      toast.error('Ошибка отправки сообщения');
    }
  };

  // Обработка ввода текста с дебаунсом для статуса "печатает"
  const handleChange = (e) => {
    setNewMessage(e.target.value);
    sendTyping(chatId, true);
    // сброс статуса набора через 1.5 сек бездействия
    clearTimeout(window.typingTimeout);
    window.typingTimeout = setTimeout(() => {
      sendTyping(chatId, false);
    }, 1500);
  };

  return (
    <Container sx={{ mt: 4 }}>
      <Typography variant="h4" gutterBottom>
        Чат {chatId}
      </Typography>

      {/* Индикатор набора текста */}
      {isTyping && (
        <Typography variant="body2" color="textSecondary" sx={{ mb: 1, fontStyle: 'italic' }}>
          Пользователь печатает...
        </Typography>
      )}

      <Box sx={{
        border: '1px solid #ccc', borderRadius: 2,
        height: '60vh', overflowY: 'auto',
        p: 2, mb: 2
      }}>
        {loading ? (
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

      {/* Форма отправки */}
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
    </Container>
  );
};

export default ChatWindow;

