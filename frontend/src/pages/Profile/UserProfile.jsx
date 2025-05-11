import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { 
  Container, Box, Typography, Avatar, Button, CircularProgress,
  Dialog, DialogTitle, DialogContent, DialogActions
} from '@mui/material';
import { getUser, getUserProfile, getUserBio } from '../../api/user';
import { getConnections, deleteConnection } from '../../api/connections';
import { toast } from 'react-toastify';

const UserProfile = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [user, setUser] = useState(null);
  const [profile, setProfile] = useState(null);
  const [bio, setBio] = useState(null);
  const [connectedIds, setConnectedIds] = useState([]);
  const [loading, setLoading] = useState(true);
  const [disconnectDialogOpen, setDisconnectDialogOpen] = useState(false);

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

  const handleDisconnect = async () => {
    try {
      await deleteConnection(id);
      toast.success('Отключение выполнено');
      setDisconnectDialogOpen(false);
      navigate('/connections');
    } catch {
      toast.error('Ошибка при отключении');
    }
  };

  if (loading) {
    return (
      <Container sx={{ textAlign: 'center', mt: 4 }}>
        <CircularProgress />
      </Container>
    );
  }

  if (!user || !profile) {
    return (
      <Container sx={{ textAlign: 'center', mt: 4 }}>
        <Typography variant="h6">Профиль не найден</Typography>
      </Container>
    );
  }

  return (
    <Container maxWidth="sm" sx={{ mt: 4 }}>
      <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
        <Avatar
          src={user.photoUrl || '/default-avatar.png'}
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
        {profile.about || 'Нет информации о пользователе'}
      </Typography>

      <Typography variant="h6" gutterBottom>
        Биография
      </Typography>
      {bio ? (
        <>
          <Typography>Интересы: {bio.interests || 'Не указаны'}</Typography>
          <Typography>Хобби: {bio.hobbies || 'Не указаны'}</Typography>
          <Typography>Музыка: {bio.music || 'Не указана'}</Typography>
          <Typography>Еда: {bio.food || 'Не указана'}</Typography>
          <Typography>Путешествия: {bio.travel || 'Не указаны'}</Typography>
          <Typography>Ищу: {bio.lookingFor || 'Не указано'}</Typography>
        </>
      ) : (
        <Typography>Биография не заполнена</Typography>
      )}

      <Box sx={{ mt: 3, display: 'flex', gap: 2 }}>
        {connectedIds.includes(id) && (
          <>
            <Button
              variant="contained"
              color="primary"
              onClick={handleChat}
            >
              Перейти в чат
            </Button>
            <Button
              variant="outlined"
              color="error"
              onClick={() => setDisconnectDialogOpen(true)}
            >
              Отключить
            </Button>
          </>
        )}
      </Box>

      {/* Модальное окно подтверждения */}
      <Dialog
        open={disconnectDialogOpen}
        onClose={() => setDisconnectDialogOpen(false)}
      >
        <DialogTitle>Отключить пользователя?</DialogTitle>
        <DialogContent>
          <Typography>
            Вы уверены, что хотите отключиться от этого пользователя?
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDisconnectDialogOpen(false)}>
            Отмена
          </Button>
          <Button onClick={handleDisconnect} color="error">
            Отключить
          </Button>
        </DialogActions>
      </Dialog>
    </Container>
  );
};

export default UserProfile;
