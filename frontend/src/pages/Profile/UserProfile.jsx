import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { 
  Container, Box, Typography, Avatar, Button, CircularProgress,
  Dialog, DialogTitle, DialogContent, DialogActions
} from '@mui/material';
import { getUser, getUserProfile, getUserBio } from '../../api/user';
import { getConnections, deleteConnection} from '../../api/connections';
import { toast } from 'react-toastify';
import { getChats } from '../../api/chat';
const UserProfile = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [user, setUser] = useState(null);
  const [profile, setProfile] = useState(null);
  const [bio, setBio] = useState(null);
  const [connectedIds, setConnectedIds] = useState([]);
  const handleRemoveFriend = async (id) => {
    try {
      await deleteConnection(id);
      toast.success('Пользователь удалён из друзей');
      // обновляем список connectedIds в локальном состоянии
      setConnectedIds(prev => prev.filter(uid => uid !== id));
      setConnectedIds(prev => prev.filter(uid => uid !== id));
      // убрать чат из списка чатов
      getChats(chs => chs.filter(c => c.otherUserID !== id));
      if (window.location.pathname === `/chat/${id}`) {
        navigate('/chats');
      }
    } catch {
      toast.error('Не удалось удалить друга');
    }
  };
  

  const [loading, setLoading] = useState(true);
  const [disconnectDialogOpen, setDisconnectDialogOpen] = useState(false);
  // useEffect(() => {
  //   const load = async () => {
  //     try {
  //       const [u, p, b, conns] = await Promise.all([
  //         getUser(id),
  //         getUserProfile(id),
  //         getUserBio(id),
  //         getConnections(),
  //       ]);
  //       setUser(u);
  //       setProfile(p);
  //       setBio(b);
  //       setConnectedIds(conns);
  //     } catch (err) {
  //       toast.error('Не удалось загрузить профиль');
  //       navigate('/recommendations');
  //     } finally {
  //       setLoading(false);
  //     }
  //   };
  //   load();
  // }, [id, navigate]);

  // useEffect(() => {
  //   const load = async () => {
  //     setLoading(true);
  //     try {
  //       const [u, conns] = await Promise.all([
  //         getUser(id),
  //         getConnections(),
  //       ]);
  //       setUser(u);
  //       setConnectedIds(conns);
  //     } catch (err) {
  //       toast.error('Не удалось загрузить пользователя');
  //       navigate('/recommendations');
  //       return;
  //     }
  
  //     try {
  //       const p = await getUserProfile(id);
  //       setProfile(p);
  //     } catch (err) {
  //       setProfile(null); // Or leave it null to handle conditionally in UI
  //     }
  
  //     try {
  //       const b = await getUserBio(id);
  //       setBio(b);
  //     } catch (err) {
  //       setBio(null); // Or default object
  //     }
  
  //     setLoading(false);
  //   };
  //   load();
  // }, [id, navigate]);
  
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
      } catch (err) {
        toast.error('Ошибка загрузки данных пользователя');
        navigate('/recommendations');
        return;
      }
  
      try {
        const p = await getUserProfile(id);
        setProfile(p);
      } catch (err) {
        setProfile(null);
      }
  
      try {
        const b = await getUserBio(id);
        setBio(b);
      } catch (err) {
        setBio(null);
      }
  
      setLoading(false);
    };
  
    load();
  }, [id, navigate]);
  

  const handleChat = async () => {
    // Получаем список чатов
    const chats = await getChats();
    // ищем по otherUserId, а для перехода используем chatId!
    const chat = chats.find(c => String(c.otherUserId) === String(id));
    if (chat && chat.chatId) {
      navigate(`/chat/${chat.chatId}`);
    } else {
      toast.error('Чат с этим пользователем не найден');
    }
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
        {profile?.about || 'Нет информации о пользователе' || 'Информация недоступна'}
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
              onClick={() =>handleRemoveFriend(id)} //setDisconnectDialogOpen(true)
            >
              Delete Friend
            </Button>
          </>
        )}
      </Box>

      {/* Модальное окно подтверждения */}
      <Dialog
        open={disconnectDialogOpen}
        onClose={() => setDisconnectDialogOpen(false)}
      >
        <DialogTitle>Delete Friend?</DialogTitle>
        <DialogContent>
          <Typography>
            Are you sure you want to delete this friend?
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDisconnectDialogOpen(false)}>
            Отмена
          </Button>
          <Button onClick={handleDisconnect} color="error">
            Delete
          </Button>
        </DialogActions>
      </Dialog>
    </Container>
  );
};
export default UserProfile;