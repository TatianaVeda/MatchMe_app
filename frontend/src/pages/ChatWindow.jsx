// m/frontend/src/pages/ChatWindow.jsx

import React, { useState, useEffect, useRef, Fragment } from 'react';
import { useParams, useNavigate, useLocation } from 'react-router-dom';
import {
  Container, Box, Typography, TextField, Button,
  CircularProgress, List, Divider, Pagination
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
  const [totalCount, setTotalCount] = useState(0);    
  const pageSize = 10;                                 
  const pageCount = Math.ceil(totalCount / pageSize);  
  const messages = allMessages[chatId] || [];
  const chatIdNum = Number(chatId);
const isTyping  = typingStatuses[chatIdNum];
  useEffect(() => {
    if (chatId === 'new') {
      const otherUserID = new URLSearchParams(location.search).get('other_user_id');
      if (!otherUserID) {
        toast.error('other_user_id is not specified');
        return;
      }
      api.post('/chats', { otherUserId: otherUserID })
        .then(({ data }) => {
          navigate(`/chat/${data.chatId}`, { replace: true });
        })
        .catch(() => {
          toast.error('Failed to open chat');
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
        params: { page: p, limit: pageSize }
      });
      setTotalCount(data.totalCount);
      setMessages(chatId, data.messages);
    } catch {
      toast.error('Error loading messages');
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
      sendMessage(chatId, content);     
    const normalized = {
        ...saved,
        sender_id: saved.senderId
      };
      setMessages(chatId, [...messages, normalized]);
      setNewMessage('');
    } catch {
      toast.error('Error sending message');
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
        {chatId === 'new' ? 'Creating chat...' : `Chat ${chatId}`}
      </Typography>
      
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
         {isTyping && (
        <Typography variant="body2" color="textSecondary" sx={{ mb: 1, fontStyle: 'italic' }}>
          User is typing...
        </Typography>
      )}
        </List>

        {pageCount > 1 && (
          <Box sx={{ display: 'flex', justifyContent: 'center', mt: 2 }}>
            <Pagination
              count={pageCount}
              page={page}
              onChange={(_, newPage) => {
                setPage(newPage);
                setLoading(true);
                fetchMessages(newPage);
              }}
              color="primary"
            />
          </Box>
        )}

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
            label="New message"
            value={newMessage}
            onChange={handleChange}
            fullWidth
            multiline
            rows={2}
          />
          <Button variant="contained" type="submit">
            Send
          </Button>
        </Box>
      )}
    </Container>
  );
};
export default ChatWindow;