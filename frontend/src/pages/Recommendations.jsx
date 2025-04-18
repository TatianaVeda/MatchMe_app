import React, { useState, useEffect } from 'react';
import { 
  Container, 
  Grid, 
  Card, 
  CardContent, 
  CardMedia, 
  Typography, 
  Button, 
  CardActions, 
  CircularProgress 
} from '@mui/material';
import { toast } from 'react-toastify';

// Импортируем методы из модулей API
import { getRecommendations } from '../../api/recommendations';
import { getUser, getUserBio } from '../../api/user';
import { sendConnectionRequest } from '../../api/connections';

const Recommendations = () => {
  // Состояние для хранения объединённых данных пользователя и его биографии
  const [recommendations, setRecommendations] = useState([]);
  const [loading, setLoading] = useState(true);

  // Функция для получения списка рекомендованных идентификаторов и соответствующих данных
  const fetchRecommendations = async () => {
    try {
      // Получаем массив идентификаторов с эндпоинта /recommendations
      const recIds = await getRecommendations();
      // Для каждого идентификатора последовательно загружаем данные пользователя и его биографию
      const recData = await Promise.all(
        recIds.map(async (id) => {
          try {
            const user = await getUser(id);
            const bio = await getUserBio(id);
            return { id, ...user, bio };
          } catch (err) {
            console.error("Ошибка загрузки данных для id", id, err);
            return null;
          }
        })
      );
      // Фильтруем полученные данные – исключаем невалидные
      setRecommendations(recData.filter((rec) => rec !== null));
    } catch (err) {
      toast.error("Ошибка загрузки рекомендаций");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchRecommendations();
  }, []);

  // Обработка кнопки "Отклонить" – убираем рекомендацию из списка
  const handleDecline = (id) => {
    setRecommendations((prev) => prev.filter((rec) => rec.id !== id));
  };

  // Обработка кнопки "Подключиться" – отправляем запрос на подключение
  const handleConnect = async (id) => {
    try {
      await sendConnectionRequest(id);
      toast.success("Запрос на подключение отправлен");
      setRecommendations((prev) => prev.filter((rec) => rec.id !== id));
    } catch (err) {
      toast.error("Ошибка при отправке запроса");
    }
  };

  if (loading) {
    return (
      <Container sx={{ textAlign: 'center', mt: 4 }}>
        <CircularProgress />
      </Container>
    );
  }

  if (recommendations.length === 0) {
    return (
      <Container sx={{ mt: 4 }}>
        <Typography variant="h6">Нет доступных рекомендаций.</Typography>
      </Container>
    );
  }

  return (
    <Container sx={{ mt: 4 }}>
      <Typography variant="h4" gutterBottom>
        Рекомендации
      </Typography>
      <Grid container spacing={3}>
        {recommendations.map((rec) => (
          <Grid item xs={12} sm={6} md={4} key={rec.id}>
            <Card>
              <CardMedia
                component="img"
                height="140"
                image={rec.photo_url || '/static/images/default.png'}
                alt={`${rec.firstName} ${rec.lastName}`}
              />
              <CardContent>
                <Typography gutterBottom variant="h5" component="div">
                  {rec.firstName} {rec.lastName}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  {rec.bio && rec.bio.interests 
                    ? `Интересы: ${rec.bio.interests}` 
                    : "Информация отсутствует"}
                </Typography>
              </CardContent>
              <CardActions>
                <Button size="small" variant="contained" onClick={() => handleConnect(rec.id)}>
                  Подключиться
                </Button>
                <Button size="small" variant="outlined" onClick={() => handleDecline(rec.id)}>
                  Отклонить
                </Button>
              </CardActions>
            </Card>
          </Grid>
        ))}
      </Grid>
    </Container>
  );
};

export default Recommendations;
