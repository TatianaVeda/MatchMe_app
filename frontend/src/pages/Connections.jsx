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

// Импортируем методы из модуля API для соединений и пользователя
import { getConnections, deleteConnection } from '../api/connections';
import { getUser } from '../api/user';

const Connections = () => {
  const [connections, setConnections] = useState([]);
  const [loading, setLoading] = useState(true);

  // Функция загрузки списка подключенных идентификаторов, а затем данных пользователя по каждому id
  const fetchConnections = async () => {
    try {
      const connectionIds = await getConnections();
      const connectionDetails = await Promise.all(
        connectionIds.map(async (id) => {
          try {
            const userData = await getUser(id);
            return { id, ...userData };
          } catch (err) {
            console.error("Ошибка загрузки данных для пользователя с id", id, err);
            return null;
          }
        })
      );
      setConnections(connectionDetails.filter(item => item !== null));
    } catch (err) {
      toast.error("Ошибка загрузки подключений");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchConnections();
  }, []);

  // Обработка кнопки "Отключить" – удаляет соединение через API и обновляет локальное состояние
  const handleDisconnect = async (id) => {
    try {
      await deleteConnection(id);
      toast.success("Отключение выполнено успешно");
      setConnections(prev => prev.filter(conn => conn.id !== id));
    } catch (err) {
      toast.error("Ошибка при отключении");
    }
  };

  if (loading) {
    return (
      <Container sx={{ textAlign: 'center', mt: 4 }}>
        <CircularProgress />
      </Container>
    );
  }

  if (connections.length === 0) {
    return (
      <Container sx={{ mt: 4 }}>
        <Typography variant="h6">Нет подключённых профилей.</Typography>
      </Container>
    );
  }

  return (
    <Container sx={{ mt: 4 }}>
      <Typography variant="h4" gutterBottom>
        Подключения
      </Typography>
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
                  {conn.firstName && conn.lastName
                    ? `${conn.firstName} ${conn.lastName}`
                    : conn.name || "Без имени"}
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
    </Container>
  );
};

export default Connections;
