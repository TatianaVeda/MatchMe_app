import React, { useEffect } from 'react';
import {
  Container,
  Typography,
  Grid,
  Card,
  CardMedia,
  Badge,
  Box,
  CircularProgress
} from '@mui/material';
import api from '../api/index';
import { toast } from 'react-toastify';
import { useChatState, useChatDispatch } from '../contexts/ChatContext';

const Chats = () => {
  const { chats } = useChatState();
  const { setChats } = useChatDispatch();

  useEffect(() => {
    const loadChats = async () => {
      try {
        const { data } = await api.get('/chats');
        setChats(data);
      } catch {
        toast.error("Ошибка загрузки чатов");
      }
    };
    loadChats();
  }, [setChats]);

  if (chats === null) {
    return (
      <Container sx={{ mt: 4, textAlign: 'center' }}>
        <CircularProgress />
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
        {chats.map((chat) => (
          <Grid item xs={12} key={chat.chat_id}>
            <Card sx={{ display: 'flex', alignItems: 'center', p: 2 }}>
              <Badge
                color="success"
                variant="dot"
                invisible={!chat.otherUserOnline}
                anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
              >
                <CardMedia
                  component="img"
                  sx={{ width: 80, height: 80, borderRadius: '50%', mr: 2 }}
                  image={chat.otherUser?.photo_url || '/static/images/default.png'}
                  alt={chat.otherUser ? `${chat.otherUser.firstName} ${chat.otherUser.lastName}` : chat.otherUserID}
                />
              </Badge>
              <Box sx={{ flexGrow: 1 }}>
                <Typography variant="h6">
                  {chat.otherUser
                    ? `${chat.otherUser.firstName} ${chat.otherUser.lastName}`
                    : chat.otherUserID}
                </Typography>
                {chat.lastMessage && (
                  <Typography variant="body2" color="textSecondary">
                    {chat.lastMessage.content}
                  </Typography>
                )}
              </Box>
              <Box sx={{ textAlign: 'right' }}>
                {chat.unreadCount > 0 && (
                  <Badge color="error" badgeContent={chat.unreadCount} />
                )}
                {chat.isTyping && (
                  <Typography variant="caption" color="primary">
                    Набирает...
                  </Typography>
                )}
              </Box>
            </Card>
          </Grid>
        ))}
      </Grid>
    </Container>
  );
};

export default Chats;
