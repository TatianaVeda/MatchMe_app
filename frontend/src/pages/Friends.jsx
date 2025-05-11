import React, { useState, useEffect } from 'react';
import { Container, Typography, Tab, Tabs, Box,
  Grid, Card, CardMedia, CardContent, CardActions, Button, CircularProgress,
  Dialog, DialogTitle, DialogContent, DialogActions
} from '@mui/material';
import { toast } from 'react-toastify';
import { useNavigate } from 'react-router-dom';
import UserCard from '../components/UserCard';
import { getConnections, getPendingConnections, updateConnectionRequest, deleteConnection } from '../api/connections';
import { getUser } from '../api/user';

const Friends = () => {
  const navigate = useNavigate();
  const [tab, setTab] = useState(0);
  const [friends, setFriends] = useState([]);
  const [pending, setPending] = useState([]);
  const [loading, setLoading] = useState(true);
  const [disconnectDialogOpen, setDisconnectDialogOpen] = useState(false);
  const [selectedUser, setSelectedUser] = useState(null);

  const fetchFriends = async () => {
    try {
      const ids = await getConnections();
      const data = await Promise.all(ids.map(async id => {
        const u = await getUser(id);
        return { id, ...u };
      }));
      setFriends(data);
    } catch {
      toast.error('Ошибка загрузки списка друзей');
    }
  };

  const fetchPending = async () => {
    try {
      const ids = await getPendingConnections();
      const data = await Promise.all(ids.map(async id => {
        const u = await getUser(id);
        return { id, ...u };
      }));
      setPending(data);
    } catch {
      toast.error('Ошибка загрузки заявок');
    }
  };

  useEffect(() => {
    setLoading(true);
    Promise.all([fetchFriends(), fetchPending()])
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const handleTabChange = (_, v) => setTab(v);

  const handleAccept = async id => {
    try {
      await updateConnectionRequest(id, 'accept');
      toast.success('Запрос принят');
      setPending(p => p.filter(u => u.id !== id));
      // добавляем в друзья
      const accepted = pending.find(u => u.id === id);
      setFriends(f => [...f, accepted]);
    } catch {
      toast.error('Ошибка при принятии');
    }
  };

  const handleDecline = async id => {
    try {
      await updateConnectionRequest(id, 'decline');
      toast.info('Запрос отклонён');
      setPending(p => p.filter(u => u.id !== id));
    } catch {
      toast.error('Ошибка при отклонении');
    }
  };

  const handleDisconnectClick = (user) => {
    setSelectedUser(user);
    setDisconnectDialogOpen(true);
  };

  const handleDisconnectConfirm = async () => {
    try {
      await deleteConnection(selectedUser.id);
      toast.success('Отключение выполнено');
      setDisconnectDialogOpen(false);
      fetchFriends(); // Перезагружаем список друзей
    } catch {
      toast.error('Ошибка при отключении');
    }
  };

  if (loading) {
    return (
      <Container sx={{ textAlign:'center', mt:4 }}>
        <CircularProgress />
      </Container>
    );
  }

  return (
    <Container sx={{ mt:4 }}>
      <Typography variant="h4" gutterBottom>Друзья</Typography>
      <Tabs value={tab} onChange={handleTabChange} sx={{ mb:3 }}>
        <Tab label="Мои друзья" />
        <Tab label="Запросы" />
      </Tabs>

      {tab === 0 && (
        friends.length === 0 ? (
          <Typography>У вас пока нет друзей.</Typography>
        ) : (
          <Grid container spacing={2}>
            {friends.map(u => (
              <Grid key={u.id} item xs={12} sm={6} md={4}>
                <UserCard
                  user={{ ...u, connected: true }}
                  showChat={true}
                  showDisconnect={true}
                  onChatClick={() => navigate(`/chat/${u.id}`)}
                  onClick={() => navigate(`/users/${u.id}`)}
                  onDisconnect={() => handleDisconnectClick(u)}
                />
              </Grid>
            ))}
          </Grid>
        )
      )}

{tab === 1 && (
        pending.length === 0 ? (
          <Typography>Нет входящих запросов.</Typography>
        ) : (
          <Grid container spacing={2}>
            {pending.map(u => (
              <Grid key={u.id} item xs={12} sm={6} md={4}>
                <UserCard
                  user={{ ...u, connected: false }}
                  showChat={false}
                  onClick={() => navigate(`/users/${u.id}`)}
                />
                {/* Кнопки «Принять/Отклонить» можно оставить под карточкой */}
                <Box sx={{ display: 'flex', justifyContent: 'center', mt: 1 }}>
                  <Button
                    size="small"
                    variant="contained"
                    onClick={() => handleAccept(u.id)}
                    sx={{ mr: 1 }}
                  >
                    Принять
                  </Button>
                  <Button
                    size="small"
                    variant="outlined"
                    onClick={() => handleDecline(u.id)}
                  >
                    Отклонить
                  </Button>
                </Box>
              </Grid>
            ))}
          </Grid>
        )
      )}

      <Dialog
        open={disconnectDialogOpen}
        onClose={() => setDisconnectDialogOpen(false)}
      >
        <DialogTitle>Отключить пользователя?</DialogTitle>
        <DialogContent>
          <Typography>
            Вы уверены, что хотите отключиться от {selectedUser?.firstName} {selectedUser?.lastName}?
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDisconnectDialogOpen(false)}>
            Отмена
          </Button>
          <Button onClick={handleDisconnectConfirm} color="error">
            Отключить
          </Button>
        </DialogActions>
      </Dialog>
    </Container>
  );
};

export default Friends;
