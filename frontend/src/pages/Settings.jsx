// m/frontend/src/pages/Settings.jsx
import React, { useState, useEffect } from 'react';
import { 
  Container, 
  Box, 
  Typography, 
  TextField, 
  Button, 
  CircularProgress 
} from '@mui/material';
import axios from '../api/index';
import { toast } from 'react-toastify';

const Settings = () => {
  // Начальное состояние настроек (предпочтений)
  const [preferences, setPreferences] = useState({
    maxRadius: ''    // Максимальный радиус для рекомендаций
    //location: ''      // Местоположение (например, название города)
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  // Функция загрузки предпочтений пользователя
  const fetchPreferences = async () => {
    try {
      // Пример GET запроса к эндпоинту для получения настроек
      const { data } = await axios.get('/me/preferences');
      setPreferences({
        maxRadius: data.maxRadius || ''
        //location: data.location || ''
      });
    } catch (error) {
      toast.error("Ошибка загрузки настроек");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchPreferences();
  }, []);

  // Обработка изменения полей формы
  // const handleChange = (e) => {
  //   const { name, value } = e.target;
  //   setPreferences((prev) => ({
  //     ...prev,
  //     [name]: value
  //   }));
  // };

  const handleChange = (e) => {
    setPreferences({ maxRadius: e.target.value });
  };
  

  // Отправка формы для сохранения изменений
  const handleSubmit = async (e) => {
    e.preventDefault();
    setSaving(true);
    try {
      // Пример PUT запроса для обновления настроек
      await axios.put('/me/preferences', {
              // Подаём именно то, что ждёт бэкенд
              maxRadius: Number(preferences.maxRadius)
            });
      toast.success("Настройки сохранены");
    } catch (error) {
      toast.error(error.response?.data?.message || "Ошибка сохранения настроек");
    } finally {
      setSaving(false);
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
    <Container maxWidth="sm" sx={{ mt: 4 }}>
      <Box component="form" onSubmit={handleSubmit} sx={{ p: 3, border: '1px solid #ccc', borderRadius: 2 }}>
        <Typography variant="h4" component="h1" gutterBottom>
          Настройки
        </Typography>
        <TextField
          label="Максимальный радиус (км)"
          name="maxRadius"
          type="number"
          fullWidth
          margin="normal"
          value={preferences.maxRadius}
          onChange={handleChange}
          required
        />
        {/* <TextField
          label="Местоположение"
          name="location"
          type="text"
          fullWidth
          margin="normal"
          value={preferences.location}
          onChange={handleChange}
          required
        /> */}
        <Button
          variant="contained"
          color="primary"
          type="submit"
          fullWidth
          sx={{ mt: 2 }}
          disabled={saving}
        >
          {saving ? "Сохранение..." : "Сохранить настройки"}
        </Button>
      </Box>
    </Container>
  );
};

export default Settings;
