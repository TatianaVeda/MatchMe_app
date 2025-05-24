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
        toast.error('Error loading chats');
      } finally {
        setLoading(false);
      }
    };
    loadChats();
  }, [setChats]);

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
        <Typography variant="h4" gutterBottom>Chats</Typography>
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
        <Typography variant="h6">No active chats</Typography>
      </Container>
    );
  }

  return (
    <Container sx={{ mt: 4 }}>
      <Typography variant="h4" gutterBottom>Chats</Typography>
      <Grid container spacing={2}>
        {chats.map(chat => (
          <Grid item xs={12} sm={6} md={4} key={chat.otherUserID}>
        
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
                onChatClick={() => handleChatClick(chat)}         
                onClick={() => navigate(`/users/${chat.otherUserID}`)} 
              />
            </Badge>
          </Grid>
        ))}
      </Grid>
    </Container>
  );
};

export default Chats;




