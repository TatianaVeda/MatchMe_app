import React, { useState, useEffect } from 'react';
import { 
  Container, 
  Typography, 
  Grid, 
  Card, 
  CardContent, 
  CardMedia, 
  CardActions, 
  Button, 
  CircularProgress 
} from '@mui/material';
import { toast } from 'react-toastify';
import {
  getPendingConnections,
  updateConnectionRequest,
  getConnections,
  deleteConnection
} from '../api/connections';
import { getUser } from '../api/user';

const Connections = () => {
  const [pending, setPending] = useState([]);
  const [connections, setConnections] = useState([]);
  const [loading, setLoading] = useState(true);

  // Загрузка входящих и принятых подключений
  const fetchAll = async () => {
    setLoading(true);
    try {
      // Входящие (pending)
      const pendingIds = await getPendingConnections();
      const pendingDetails = await Promise.all(
        pendingIds.map(async (id) => {
          try {
            const userData = await getUser(id);
            return { id, ...userData };
          } catch (err) {
            console.error('Ошибка загрузки данных pending для id', id, err);
            return null;
          }
        })
      );
      setPending(pendingDetails.filter((u) => u !== null));

      // Принятые
      const acceptedIds = await getConnections();
      const acceptedDetails = await Promise.all(
        acceptedIds.map(async (id) => {
          try {
            const userData = await getUser(id);
            return { id, ...userData };
          } catch (err) {
            console.error('Ошибка загрузки данных accepted для id', id, err);
            return null;
          }
        })
      );
      setConnections(acceptedDetails.filter((u) => u !== null));
    } catch (err) {
      toast.error('Ошибка загрузки подключений');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAll();
  }, []);

  const handleAccept = async (id) => {
    try {
      await updateConnectionRequest(id, 'accept');
      toast.success('Запрос принят');
      const acceptedUser = pending.find((u) => u.id === id);
      setConnections((prev) => [...prev, acceptedUser]);
      setPending((prev) => prev.filter((u) => u.id !== id));
    } catch {
      toast.error('Ошибка при принятии запроса');
    }
  };

  const handleDeclinePending = async (id) => {
    try {
      await updateConnectionRequest(id, 'decline');
      toast.info('Запрос отклонён');
      setPending((prev) => prev.filter((u) => u.id !== id));
    } catch {
      toast.error('Ошибка при отклонении запроса');
    }
  };

  const handleDisconnect = async (id) => {
    try {
      await deleteConnection(id);
      toast.success('Отключение выполнено успешно');
      setConnections((prev) => prev.filter((conn) => conn.id !== id));
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

  return (
    <Container sx={{ mt: 4 }}>
      {/* Входящие запросы */}
      <Typography variant="h4" gutterBottom>
        Запросы на подключение
      </Typography>
      {pending.length === 0 ? (
        <Typography sx={{ mb: 4 }}>Нет входящих запросов.</Typography>
      ) : (
        <Grid container spacing={3} sx={{ mb: 4 }}>
          {pending.map((user) => (
            <Grid item xs={12} sm={6} md={4} key={user.id}>
              <Card>
                <CardMedia
                  component="img"
                  height="140"
                  image={user.photoUrl || '/static/images/default.png'}
                  alt={`${user.firstName} ${user.lastName}`}
                />
                <CardContent>
                  <Typography variant="h6">
                    {user.firstName} {user.lastName}
                  </Typography>
                </CardContent>
                <CardActions>
                  <Button size="small" variant="contained" onClick={() => handleAccept(user.id)}>
                    Принять
                  </Button>
                  <Button size="small" variant="outlined" onClick={() => handleDeclinePending(user.id)}>
                    Отклонить
                  </Button>
                </CardActions>
              </Card>
            </Grid>
          ))}
        </Grid>
      )}

      {/* Принятые подключения */}
      <Typography variant="h4" gutterBottom>
        Подключения
      </Typography>
      {connections.length === 0 ? (
        <Typography>Нет подключённых профилей.</Typography>
      ) : (
        <Grid container spacing={3}>
          {connections.map((conn) => (
            <Grid item xs={12} sm={6} md={4} key={conn.id}>
              <Card>
                <CardMedia
                  component="img"
                  height="140"
                  image={conn.photoUrl || '/static/images/default.png'}
                  alt={`${conn.firstName} ${conn.lastName}`}
                />
                <CardContent>
                  <Typography variant="h6">
                    {conn.firstName} {conn.lastName}
                  </Typography>
                </CardContent>
                <CardActions>
                  <Button variant="outlined" color="error" onClick={() => handleDisconnect(conn.id)}>
                    Отключить
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

export default Connections;
