import React, { useState, useEffect, useRef } from 'react';
import { useParams } from 'react-router-dom';
import {
  Container,
  Box,
  Typography,
  TextField,
  Button,
  CircularProgress,
  List,
  ListItem,
  ListItemText,
  Divider
} from '@mui/material';
import api from '../api/index';
import { toast } from 'react-toastify';
import { useChatState, useChatDispatch } from '../contexts/ChatContext';

const ChatWindow = () => {
  const { chatId } = useParams();
  const { messages: allMessages } = useChatState();
  const { setMessages, sendMessage, sendTyping } = useChatDispatch();

  const messagesEndRef = useRef(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [newMessage, setNewMessage] = useState('');

  const messages = allMessages[chatId] || [];

  const fetchMessages = async (p = 1) => {
    try {
      const { data } = await api.get(`/chats/${chatId}`, {
        params: { page: p, limit: 20 }
      });
      setMessages(chatId, p === 1 ? data : [...data, ...messages]);
    } catch {
      toast.error("Ошибка загрузки сообщений");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    setLoading(true);
    fetchMessages(page);
  }, [chatId, page]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleSend = async (e) => {
    e.preventDefault();
    if (!newMessage.trim()) return;

    try {
      const { data: saved } = await api.post(`/chats/${chatId}/messages`, { content: newMessage });
      sendMessage(chatId, newMessage);
      setMessages(chatId, [...messages, saved]);
      setNewMessage('');
    } catch {
      toast.error("Ошибка отправки сообщения");
    }
  };

  const handleChange = (e) => {
    setNewMessage(e.target.value);
    sendTyping(chatId, true);
  };

  return (
    <Container sx={{ mt: 4 }}>
      <Typography variant="h4" gutterBottom>Чат {chatId}</Typography>
      <Box sx={{ border: '1px solid #ccc', borderRadius: 2, height: '60vh', overflowY: 'auto', p: 2, mb: 2 }}>
        {loading ? (
          <Box sx={{ textAlign: 'center', mt: 2 }}>
            <CircularProgress />
          </Box>
        ) : (
          <>
            {page > 1 && (
              <Box sx={{ textAlign: 'center', mb: 1 }}>
                <Button onClick={() => setPage(p => p + 1)}>Загрузить ещё</Button>
              </Box>
            )}
            <List>
              {messages.map(msg => (
                <React.Fragment key={msg.id}>
                  <ListItem alignItems="flex-start">
                    <ListItemText
                      primary={msg.content}
                      secondary={new Date(msg.timestamp).toLocaleString()}
                    />
                  </ListItem>
                  <Divider component="li" />
                </React.Fragment>
              ))}
            </List>
            <div ref={messagesEndRef} />
          </>
        )}
      </Box>
      <Box component="form" onSubmit={handleSend} sx={{ display: 'flex', gap: 1 }}>
        <TextField
          label="Новое сообщение"
          value={newMessage}
          onChange={handleChange}
          fullWidth
          multiline
          rows={2}
        />
        <Button variant="contained" type="submit">Отправить</Button>
      </Box>
    </Container>
  );
};

export default ChatWindow;
