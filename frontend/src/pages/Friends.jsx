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
  getPendingConnections,    
  getSentConnections,       
  updateConnectionRequest,
  deleteConnection
} from '../api/connections';
import { getUser, getBatchOnlineStatus } from '../api/user';
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

  const handleChatClick = (userId) => {
    const existing = chats.find(c => c.otherUserID === userId);
    navigate(existing ? `/chat/${existing.id}` : `/chat/new?other_user_id=${userId}`);
  };


  const loadUsers = async (ids) => {
    const rawUsers = await Promise.all(ids.map(id => getUser(id)));
    const presenceMap = await getBatchOnlineStatus(ids);
    return rawUsers.map(u => ({
      ...u,
      online: Boolean(presenceMap[u.id]),
    }));
  };

  useEffect(() => {
    const interval = setInterval(() => {
      fetchFriends();
    }, 60000);
    return () => clearInterval(interval);
  }, []);  
  
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
      .catch(() => toast.error('Error loading data'))
      .finally(() => setLoading(false));
  }, []);

  const handleAccept = async (id) => {
    try {
      await updateConnectionRequest(id, 'accept');
      toast.success('Request accepted');
  
      const acceptedUser = incoming.find(u => u.id === id);
      
      setIncoming(prevIncoming =>
        prevIncoming.filter(u => u.id !== id)
      );
      
      setFriends(prevFriends => [
        ...prevFriends,
        acceptedUser
      ]);
    } catch {
      toast.error('Error accepting request');
    }
  };
  

  const handleDecline = async (id) => {
    try {
      await updateConnectionRequest(id, 'decline');
      toast.info('Friend request declined');
  
      setIncoming(prevIncoming =>
        prevIncoming.filter(u => u.id !== id)
      );
    } catch {
      toast.error('Error declining request');
    }
  };
  

  const handleRemove = async (id) => {
    await deleteConnection(id);
    toast.success('User has been blocked');
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
      <Typography variant="h4" gutterBottom>Friends</Typography>

      <Tabs value={tab} onChange={(_, v) => setTab(v)} sx={{ mb: 3 }}>
        <Tab label="My Friends" />
        <Tab label="Requests" />
      </Tabs>

      {tab === 0 && (
        friends.length === 0
          ? <Typography>You don't have friends yet.</Typography>
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
                      Remove from Friends
                    </Button>
                  </Box>
                </Grid>
              ))}
            </Grid>
          )
      )}

      {tab === 1 && (
        incoming.length === 0 && outgoing.length === 0
          ? <Typography>No friend requests.</Typography>
          : (
            <>
              {incoming.length > 0 && (
                <>
                  <Typography variant="h6">Incoming Requests</Typography>
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
                            Accept
                          </Button>
                          <Button size="small" variant="outlined" onClick={() => handleDecline(u.id)}>
                            Decline
                          </Button>
                        </Box>
                      </Grid>
                    ))}
                  </Grid>
                </>
              )}

              {outgoing.length > 0 && (
                <>
                  <Typography variant="h6">Outgoing Requests</Typography>
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
                            Request Sent
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
