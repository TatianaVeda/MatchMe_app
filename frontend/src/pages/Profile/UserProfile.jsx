// /m/frontend/src/pages/Profile/UserProfile.jsx
import React, { useState, useEffect } from 'react';
import { useParams, useNavigate, useLocation } from 'react-router-dom';
import {
  Container,
  Box,
  Typography,
  Avatar,
  Button,
  CircularProgress,
  Badge
} from '@mui/material';
import { getUser, getUserProfile, getUserBio } from '../../api/user';
import { getConnections, deleteConnection } from '../../api/connections';
import { toast } from 'react-toastify';
import { useChatState, useChatDispatch } from '../../contexts/ChatContext';
//import useWebSocket from '../../hooks/useWebSocket';

const UserProfile = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const location = useLocation();
 const { distance, score } = location.state || {};
 
  console.log('UserProfile location.state →', location.state);
  
  const { chats, presence } = useChatState();
  const { setChats } = useChatDispatch();
  //const { subscribe, unsubscribe } = useWebSocket();
  
  const [user, setUser] = useState(null);
  const [profile, setProfile] = useState(null);
  const [bio, setBio] = useState(null);
  const [connectedIds, setConnectedIds] = useState([]);
  
  const [loading, setLoading] = useState(true);

  const handleRemoveFriend = async (friendId) => {
    try {
      await deleteConnection(friendId);
      toast.success('User removed from friends');
      setConnectedIds(prev => prev.filter(uid => uid !== friendId));
      setChats(chs => chs.filter(c => c.otherUserID !== friendId));
      if (window.location.pathname === `/chat/${friendId}`) {
        navigate('/chats');
      }
    } catch {
      toast.error('Failed to remove friend');
    }
  };
  
  useEffect(() => {
    const load = async () => {
      setLoading(true);
      try {
        const [u, conns] = await Promise.all([
          getUser(id),
          getConnections(),
        ]);
        if (!u) {
          setUser(null);
          setLoading(false);
          return;
        }
        setUser(u);
        setConnectedIds(conns);
      } catch {
        toast.error('Error loading user data');
        navigate('/recommendations');
        return;
      }
      try {
        const p = await getUserProfile(id);
        setProfile(p);
      } catch {
        setProfile(null);
      }
      try {
        const b = await getUserBio(id);
        setBio(b);
      } catch {
        setBio(null);
      }
      setLoading(false);
    };
    load();
  }, [id, navigate]);
 
  const isOnline =
    (typeof user?.online === 'boolean' && user.online) ||
    Boolean(presence?.[user?.id]);
  const handleChat = () => {
    const existing = chats.find(c => c.otherUserID === id);
    navigate(existing ? `/chat/${existing.id}` : `/chat/new?other_user_id=${id}`);
  };

  if (loading) {
    return (
      <Container sx={{ textAlign: 'center', mt: 4 }}>
        <CircularProgress />
      </Container>
    );
  }

  if (!user) {
    return (
      <Container sx={{ textAlign: 'center', mt: 4 }}>
        <Typography variant="h5">User not found</Typography>
      </Container>
    );
  }

  return (
    <Container maxWidth="sm" sx={{ mt: 4 }}>
      <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
        <Badge
          color={isOnline ? 'success' : 'error'}
          variant="dot"
          overlap="circular"
          anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
        >
          <Avatar
            src={user.photoUrl}
            alt={`${user.firstName} ${user.lastName}`}
            sx={{ width: 80, height: 80, mr: 2 }}
          >
            {!user.photoUrl && '👤'}
          </Avatar>
        </Badge>
        <Typography variant="h4">
          {user.firstName} {user.lastName}
        </Typography>
      </Box>

{(distance !== undefined && score !== undefined) && (
  <Box sx={{ mb: 2 }}>
    <Typography variant="body2" color="text.secondary">
    Distance: {distance.toFixed(1)} км
    </Typography>
    <Typography variant="body2" color="text.secondary">
    Score: {(score * 100).toFixed(0)} %
    </Typography>
  </Box>
)}

      <Typography variant="body1" sx={{ mb: 2 }}>
        {profile?.about || 'Information not available'}
      </Typography>

      <Typography variant="h6" gutterBottom>Biography</Typography>
      {bio ? (
        <>
          <Typography>Interests: {bio.interests}</Typography>
          <Typography>Hobbies: {bio.hobbies}</Typography>
          <Typography>Music: {bio.music}</Typography>
          <Typography>Cuisine: {bio.food}</Typography>
          <Typography>Travel: {bio.travel}</Typography>
          <Typography>Looking for: {bio.lookingFor}</Typography>
        </>
      ) : (
        <Typography>Biography not available</Typography>
      )}

      {connectedIds.includes(id) && (
        <Box sx={{ mt: 3 }}>
          <Button
            variant="contained"
            color="primary"
            sx={{ mr: 1 }}
            onClick={handleChat}
          >
            Go to Chat
          </Button>
          <Button
            variant="outlined"
            color="error"
            onClick={() => handleRemoveFriend(id)}
          >
            Block User
          </Button>
        </Box>
      )}
    </Container>
  );
};

export default UserProfile;
