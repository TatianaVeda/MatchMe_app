// /m/frontend/src/pages/Chats.jsx
import React, { useState, useEffect } from 'react';
import { Container, Typography, Grid, Skeleton, Badge } from '@mui/material';
import { useNavigate } from 'react-router-dom';
import api from '../api/index';
import { toast } from 'react-toastify';
import { useChatState, useChatDispatch } from '../contexts/ChatContext';
import UserCard from '../components/UserCard';

const Chats = () => {
  const navigate = useNavigate();
  const { chats } = useChatState();
  const { setChats } = useChatDispatch();
  const [loading, setLoading] = useState(true);

  // 1) Загружаем список чатов
  useEffect(() => {
    const loadChats = async () => {
      try {
        const { data } = await api.get('/chats');
        const normalized = data.map(c => ({
          id:               c.chatId,
          otherUserID:      c.otherUserId,
          otherUser:        c.otherUser,
          unreadCount:      c.unreadCount,
          otherUserOnline:  c.otherUserOnline,
        }));
        setChats(normalized);
      } catch (err) {
        toast.error('Ошибка загрузки чатов');
      } finally {
        setLoading(false);
      }
    };
    loadChats();
  }, [setChats]);

  // 2) Функция перехода в чат
  const handleChatClick = (chat) => {
    if (chat.id) {
      navigate(`/chat/${chat.id}`);
    } else {
      navigate(`/chat/new?other_user_id=${chat.otherUserID}`);
    }
  };

  if (loading) {
    return (
      <Container sx={{ mt: 4 }}>
        <Typography variant="h4" gutterBottom>Чаты</Typography>
        <Grid container spacing={2}>
          {[...Array(6)].map((_, i) => (
            <Grid item xs={12} sm={6} md={4} key={i}>
              <Skeleton variant="rectangular" height={120} />
            </Grid>
          ))}
        </Grid>
      </Container>
    );
  }

  if (chats.length === 0) {
    return (
      <Container sx={{ mt: 4 }}>
        <Typography variant="h6">Нет активных чатов</Typography>
      </Container>
    );
  }

  return (
    <Container sx={{ mt: 4 }}>
      <Typography variant="h4" gutterBottom>Чаты</Typography>
      <Grid container spacing={2}>
        {chats.map(chat => (
          <Grid item xs={12} sm={6} md={4} key={chat.otherUserID}>
            {/* Badge поверх аватарки */}
            <Badge
              badgeContent={chat.unreadCount}
              color="error"
              overlap="circular"
              invisible={chat.unreadCount === 0}
              anchorOrigin={{
                vertical: 'top',
                horizontal: 'right',
              }}
            >
              <UserCard
                user={{
                  id:        chat.otherUserID,
                  firstName: chat.otherUser?.firstName,
                  lastName:  chat.otherUser?.lastName,
                  photoUrl:  chat.otherUser?.photoUrl,
                  online:    chat.otherUserOnline,
                  connected: true, 
                }}
                showChat
                onChatClick={() => handleChatClick(chat)}         // клик по иконке чата
                onClick={() => navigate(`/users/${chat.otherUserID}`)} // клик по имени/аватару
              />
            </Badge>
          </Grid>
        ))}
      </Grid>
    </Container>
  );
};

export default Chats;




