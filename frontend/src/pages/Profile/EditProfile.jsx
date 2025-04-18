import React, { useState, useEffect } from 'react';
import { Container, Box, Typography, TextField, Button, CircularProgress } from '@mui/material';
import { useNavigate } from 'react-router-dom';
import { toast } from 'react-toastify';

// Импортируем методы для получения и обновления данных профиля/биографии
import { getMyProfile, getMyBio, updateMyProfile, updateMyBio } from '../../api/user';

const EditProfile = () => {
  const navigate = useNavigate();
  const [profileData, setProfileData] = useState({ firstName: '', lastName: '', about: '' });
  const [bioData, setBioData] = useState({ interests: '', hobbies: '', music: '', food: '', travel: '' });
  const [loading, setLoading] = useState(true);

  // Загрузка текущих данных профиля
  const fetchProfileData = async () => {
    try {
      const data = await getMyProfile();
      setProfileData({
        firstName: data.firstName || '',
        lastName: data.lastName || '',
        about: data.about || '',
      });
    } catch (error) {
      toast.error("Ошибка загрузки профиля");
    }
  };

  // Загрузка текущей биографии
  const fetchBioData = async () => {
    try {
      const data = await getMyBio();
      setBioData({
        interests: data.interests || '',
        hobbies: data.hobbies || '',
        music: data.music || '',
        food: data.food || '',
        travel: data.travel || '',
      });
    } catch (error) {
      toast.error("Ошибка загрузки биографии");
    }
  };

  useEffect(() => {
    const loadData = async () => {
      await Promise.all([fetchProfileData(), fetchBioData()]);
      setLoading(false);
    };
    loadData();
  }, []);

  // Обработка изменения полей профиля
  const handleProfileChange = (e) => {
    setProfileData({ ...profileData, [e.target.name]: e.target.value });
  };

  // Обработка изменения полей биографии
  const handleBioChange = (e) => {
    setBioData({ ...bioData, [e.target.name]: e.target.value });
  };

  // Отправка обновлённых данных профиля и биографии через методы API
  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      await updateMyProfile(profileData);
      await updateMyBio(bioData);
      toast.success("Профиль обновлён успешно");
      navigate('/me');
    } catch (error) {
      toast.error(error.response?.data?.message || "Ошибка при обновлении профиля");
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
          Редактировать профиль
        </Typography>
        {/* Раздел профиля */}
        <Typography variant="h6">Профиль</Typography>
        <TextField
          label="Имя"
          name="firstName"
          fullWidth
          margin="normal"
          value={profileData.firstName}
          onChange={handleProfileChange}
          required
        />
        <TextField
          label="Фамилия"
          name="lastName"
          fullWidth
          margin="normal"
          value={profileData.lastName}
          onChange={handleProfileChange}
          required
        />
        <TextField
          label="О себе"
          name="about"
          fullWidth
          margin="normal"
          value={profileData.about}
          onChange={handleProfileChange}
          multiline
          rows={3}
        />
        {/* Раздел биографии */}
        <Typography variant="h6" sx={{ mt: 3 }}>
          Биография
        </Typography>
        <TextField
          label="Интересы"
          name="interests"
          fullWidth
          margin="normal"
          value={bioData.interests}
          onChange={handleBioChange}
        />
        <TextField
          label="Хобби"
          name="hobbies"
          fullWidth
          margin="normal"
          value={bioData.hobbies}
          onChange={handleBioChange}
        />
        <TextField
          label="Музыка"
          name="music"
          fullWidth
          margin="normal"
          value={bioData.music}
          onChange={handleBioChange}
        />
        <TextField
          label="Еда"
          name="food"
          fullWidth
          margin="normal"
          value={bioData.food}
          onChange={handleBioChange}
        />
        <TextField
          label="Путешествия"
          name="travel"
          fullWidth
          margin="normal"
          value={bioData.travel}
          onChange={handleBioChange}
        />
        <Button variant="contained" color="primary" type="submit" fullWidth sx={{ mt: 2 }}>
          Сохранить изменения
        </Button>
      </Box>
    </Container>
  );
};

export default EditProfile;
