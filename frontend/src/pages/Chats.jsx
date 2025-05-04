// /m/frontend/src/pages/Chats.jsx
import React, { useState, useEffect } from 'react';
import {
  Container,
  Typography,
  Grid,
  Skeleton,
  Box
} from '@mui/material';
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

  useEffect(() => {
    const loadChats = async () => {
      try {
        const { data } = await api.get('/chats');
        setChats(data);
      } catch {
        toast.error('Ошибка загрузки чатов');
      } finally {
        setLoading(false);
      }
    };
    loadChats();
  }, [setChats]);

  if (loading) {
    return (
      <Container sx={{ mt: 4 }}>
        <Typography variant="h4" gutterBottom>
          Чаты
        </Typography>
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
      <Typography variant="h4" gutterBottom>
        Чаты
      </Typography>
      <Grid container spacing={2}>
        {chats.map(chat => (
          // <Grid item xs={12} sm={6} md={4} key={chat.chat_id}>
          <Grid item xs={12} sm={6} md={4} key={chat.id}>
            <Box
              sx={{ position: 'relative', cursor: 'pointer' }}
              // onClick={() => navigate(`/chat/${chat.chat_id}`)}
              onClick={() => navigate(`/chat/${chat.id}`)}
            >
              <UserCard
                user={{
                  id:        chat.otherUserID,
                  firstName: chat.otherUser?.firstName,
                  lastName:  chat.otherUser?.lastName,
                  photoUrl: chat.otherUser?.photoUrl,
                  online:    chat.otherUserOnline
                }}
              />
              {chat.unreadCount > 0 && (
                <Box sx={{ position: 'absolute', top: 8, right: 16 }}>
                  <Typography variant="caption" color="error">
                    {chat.unreadCount}
                  </Typography>
                </Box>
              )}
            </Box>
          </Grid>
        ))}
      </Grid>
    </Container>
  );
};

export default Chats;
