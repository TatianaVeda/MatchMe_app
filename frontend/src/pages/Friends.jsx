import React, { useState, useEffect } from 'react';
import {
  Container, Typography, Box, Tab, Tabs,
  Grid, Card, CardMedia, CardContent, CardActions, Button, CircularProgress
} from '@mui/material';
import { toast } from 'react-toastify';
import { getConnections, getPendingConnections, updateConnectionRequest, deleteConnection } from '../api/connections';
import { getUser } from '../api/user';

const Friends = () => {
  const [tab, setTab] = useState(0);
  const [friends, setFriends] = useState([]);
  const [pending, setPending] = useState([]);
  const [loading, setLoading] = useState(true);

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

  const handleRemove = async id => {
    try {
      await deleteConnection(id);
      toast.success('Пользователь удалён из друзей');
      setFriends(f => f.filter(u => u.id !== id));
    } catch {
      toast.error('Ошибка при удалении');
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
        friends.length === 0
          ? <Typography>У вас пока нет друзей.</Typography>
          : <Grid container spacing={2}>
              {friends.map(u => (
                <Grid key={u.id} item xs={12} sm={6} md={4}>
                  <Card>
                    <CardMedia
                      component="img"
                      height="140"
                      image={u.photoUrl || '/static/images/default.png'}
                      alt={`${u.firstName} ${u.lastName}`}
                    />
                    <CardContent>
                      <Typography variant="h6">
                        {u.firstName} {u.lastName}
                      </Typography>
                    </CardContent>
                    <CardActions>
                      <Button
                        size="small"
                        color="error"
                        variant="outlined"
                        onClick={() => handleRemove(u.id)}
                      >
                        Удалить
                      </Button>
                    </CardActions>
                  </Card>
                </Grid>
              ))}
            </Grid>
      )}

      {tab === 1 && (
        pending.length === 0
          ? <Typography>Нет входящих запросов.</Typography>
          : <Grid container spacing={2}>
              {pending.map(u => (
                <Grid key={u.id} item xs={12} sm={6} md={4}>
                  <Card>
                    <CardMedia
                      component="img"
                      height="140"
                      image={u.photoUrl || '/static/images/default.png'}
                      alt={`${u.firstName} ${u.lastName}`}
                    />
                    <CardContent>
                      <Typography variant="h6">
                        {u.firstName} {u.lastName}
                      </Typography>
                    </CardContent>
                    <CardActions>
                      <Button
                        size="small"
                        variant="contained"
                        onClick={() => handleAccept(u.id)}
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
                    </CardActions>
                  </Card>
                </Grid>
              ))}
            </Grid>
      )}
    </Container>
  );
};

export default Friends;
