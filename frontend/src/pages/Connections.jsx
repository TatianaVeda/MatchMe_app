import React, { useState, useEffect } from 'react';
import {
   Container, Typography, Tab, Tabs, Box, Grid, Button, CircularProgress 
} from '@mui/material';
import { toast } from 'react-toastify';
import { useNavigate } from 'react-router-dom';
import UserCard from '../components/UserCard';
import { getConnections, getPendingConnections, updateConnectionRequest, deleteConnection
} from '../api/connections';
import { getUser, getBatchOnlineStatus } from '../api/user';
import { useChatState, useChatDispatch } from '../contexts/ChatContext';

const Connections = () => {
  const navigate = useNavigate();
  const { setChats } = useChatDispatch();
  const { chats } = useChatState();

  const [tab, setTab] = useState(0);
  const [connections, setConnections] = useState([]);  
  const [pending, setPending] = useState([]);        
  const [loading, setLoading] = useState(true);

  const handleChatClick = (userId) => {
    const existing = chats.find(c => c.otherUserID === userId);
    navigate(existing ? `/chat/${existing.id}` : `/chat/new?other_user_id=${userId}`);
  };

  const loadUsers = async (ids) => {
        const rawUsers = await Promise.all(ids.map(id => getUser(id)));

        const presenceMap = await getBatchOnlineStatus(ids);

        return rawUsers.map(u => ({
          ...u,
          online: Boolean(presenceMap[u.id])
        }));
      };

  const fetchData = async () => {
    setLoading(true);
    try {
      const pendingIds = await getPendingConnections();
      setPending(await loadUsers(pendingIds));

      const connIds = await getConnections();
      setConnections(await loadUsers(connIds));
    } catch {
      toast.error('Error loading connections');
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
      toast.success('Request accepted');
      const user = pending.find(u => u.id === id);
      setPending(p => p.filter(u => u.id !== id));
      setConnections(c => [...c, user]);
    } catch {
      toast.error('Error accepting request');
    }
  };

  const handleDecline = async (id) => {
    try {
      await updateConnectionRequest(id, 'decline');
      toast.info('Request declined');
      setPending(p => p.filter(u => u.id !== id));
    } catch {
      toast.error('Error declining request');
    }
  };

  const handleDisconnect = async (id) => {
    try {
      await deleteConnection(id);
      toast.success('Connection removed');
      setConnections(c => c.filter(u => u.id !== id));
      setChats(chs => chs.filter(c => c.otherUserID !== id));
      if (window.location.pathname === `/chat/${id}`) navigate('/chats');
    } catch {
      toast.error('Error disconnecting');
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
      <Typography variant="h4" gutterBottom>Connections</Typography>
      <Tabs value={tab} onChange={(e, v) => setTab(v)} sx={{ mb: 3 }}>
        <Tab label="Existing" />
        <Tab label="Requests" />
      </Tabs>

      {tab === 0 && (
        connections.length === 0
          ? <Typography>No connected profiles.</Typography>
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
                      Block
                    </Button>
                  </Box>
                </Grid>
              ))}
            </Grid>
          )
      )}

      {tab === 1 && (
        pending.length === 0
          ? <Typography>No incoming requests.</Typography>
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
                      Accept
                    </Button>
                    <Button size="small" variant="outlined" onClick={() => handleDecline(u.id)}>
                      Decline
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
