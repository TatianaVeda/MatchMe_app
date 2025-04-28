import React, { useState, useEffect } from 'react';
import { Container, Box, Typography, Button, Avatar, CircularProgress } from '@mui/material';
import { useNavigate } from 'react-router-dom';
import { toast } from 'react-toastify';
import { useAuthState } from '../../contexts/AuthContext';

// Импорт API-методов
import { getMyProfile, getMyBio } from '../../api/user';

const MyProfile = () => {
  const [profile, setProfile] = useState(null);
  const [bio, setBio] = useState(null);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();
  const { accessToken } = useAuthState();

  const fetchProfile = async () => {
    try {
      const data = await getMyProfile();
      setProfile(data);
    } catch (error) {
      toast.error(error.response?.data?.message || 'Ошибка загрузки профиля');
    }
  };

  const fetchBio = async () => {
    try {
      const data = await getMyBio();
      setBio(data);
    } catch (error) {
      toast.error(error.response?.data?.message || 'Ошибка загрузки биографии');
    }
  };

  useEffect(() => {
    if (!accessToken) return;

    const loadData = async () => {
      setLoading(true);
      try {
        await Promise.all([fetchProfile(), fetchBio()]);
      } catch (err) {
        console.error(err);
        window.location.href = '/login';
      } finally {
        setLoading(false);
      }
    };
    loadData();
  }, [accessToken]);

  const handleEdit = () => {
    navigate('/edit-profile');
  };

  if (loading) {
    return (
      <Container sx={{ textAlign: 'center', mt: 4 }}>
        <CircularProgress />
      </Container>
    );
  }

  if (!profile) {
    return (
      <Container sx={{ mt: 4 }}>
        <Typography variant="h6">Профиль не найден.</Typography>
      </Container>
    );
  }

  return (
    <Container maxWidth="md" sx={{ mt: 4 }}>
      <Box sx={{ display: 'flex', alignItems: 'center', mb: 3 }}>
        <Avatar
          alt={`${profile.firstName} ${profile.lastName}`}
          src={profile.photoUrl || '/static/images/default.png'}
          sx={{ width: 80, height: 80, mr: 2 }}
        />
        <Typography variant="h4">
          {profile.firstName} {profile.lastName}
        </Typography>
      </Box>

      {/* О себе */}
      <Box sx={{ mb: 3 }}>
        <Typography variant="body1" color="textSecondary">
          {profile.about || 'Информация о пользователе не заполнена.'}
        </Typography>
      </Box>

      {/* Город */}
      <Box sx={{ mb: 3 }}>
        <Typography variant="body1">
          Город: {profile.city || 'Не указан'}
        </Typography>
      </Box>

      {/* Биография */}
      <Box sx={{ mb: 3 }}>
        <Typography variant="h6" gutterBottom>
          Биография
        </Typography>
        {bio ? (
          <>
            <Typography variant="body1">Интересы: {bio.interests || 'Не указаны'}</Typography>
            <Typography variant="body1">Хобби: {bio.hobbies || 'Не указаны'}</Typography>
            <Typography variant="body1">Музыка: {bio.music || 'Не указана'}</Typography>
            <Typography variant="body1">Еда: {bio.food || 'Не указана'}</Typography>
            <Typography variant="body1">Путешествия: {bio.travel || 'Не указаны'}</Typography>
            <Typography variant="body1">Ищу: {bio.lookingFor || 'Не указано'}</Typography>
          </>
        ) : (
          <Typography variant="body1">Биография не заполнена.</Typography>
        )}
      </Box>

      <Button variant="contained" color="primary" onClick={handleEdit}>
        Редактировать профиль
      </Button>
    </Container>
  );
};

export default MyProfile;
