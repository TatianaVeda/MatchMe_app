import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Container, Box, Typography, Avatar, Button, CircularProgress } from '@mui/material';
import { getUser, getUserProfile, getUserBio } from '../../api/user';
import { getConnections } from '../../api/connections';
import { toast } from 'react-toastify';

const UserProfile = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [user, setUser] = useState(null);
  const [profile, setProfile] = useState(null);
  const [bio, setBio] = useState(null);
  const [connectedIds, setConnectedIds] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const load = async () => {
      try {
        const [u, p, b, conns] = await Promise.all([
          getUser(id),
          getUserProfile(id),
          getUserBio(id),
          getConnections(), // возвращает массив id подключенных
        ]);
        setUser(u);
        setProfile(p);
        setBio(b);
        setConnectedIds(conns);
      } catch (err) {
        toast.error('Не удалось загрузить профиль');
        navigate('/recommendations');
      } finally {
        setLoading(false);
      }
    };
    load();
  }, [id, navigate]);

  const handleChat = () => {
    // предположим, что чат уже создан при соединении
    navigate(`/chat/${id}`);
  };

  if (loading) {
    return (
      <Container sx={{ textAlign: 'center', mt: 4 }}>
        <CircularProgress />
      </Container>
    );
  }

  return (
    <Container maxWidth="sm" sx={{ mt: 4 }}>
      <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
        <Avatar
          src={user.photoUrl}
          alt={`${user.firstName} ${user.lastName}`}
          sx={{ width: 80, height: 80, mr: 2 }}
        >
          {!user.photoUrl && '👤'}
        </Avatar>
        <Typography variant="h4">
          {user.firstName} {user.lastName}
        </Typography>
      </Box>

      <Typography variant="body1" sx={{ mb: 2 }}>
        {profile.about}
      </Typography>

      <Typography variant="h6" gutterBottom>
        Биография
      </Typography>
      <Typography>Интересы: {bio.interests}</Typography>
      <Typography>Хобби: {bio.hobbies}</Typography>
      <Typography>Музыка: {bio.music}</Typography>
      <Typography>Еда: {bio.food}</Typography>
      <Typography>Путешествия: {bio.travel}</Typography>
      <Typography>Ищу: {bio.lookingFor}</Typography>

      {connectedIds.includes(id) && (
        <Button
          variant="contained"
          color="primary"
          sx={{ mt: 3 }}
          onClick={handleChat}
        >
          Перейти в чат
        </Button>
      )}
    </Container>
  );
};

export default UserProfile;
