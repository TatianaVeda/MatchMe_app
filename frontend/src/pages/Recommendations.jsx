// //m/frontend/src/pages/Recommendations.jsx

import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Container, Grid, Card, CardContent, CardMedia, Typography, Button, CardActions,
CircularProgress, ToggleButton, ToggleButtonGroup, Box, TextField, FormControl, InputLabel, 
Select, MenuItem, Checkbox, ListItemText, FormControlLabel } from '@mui/material';
import { toast } from 'react-toastify';

import { getRecommendations, declineRecommendation } from '../api/recommendations';
import { getUser, getUserBio } from '../api/user';
import { sendConnectionRequest } from '../api/connections';

// Глобальные константы опций (можно вынести в отдельный файл)
const cityOptions = [
  { name: 'Helsinki', lat: 60.1699, lon: 24.9384 },
  { name: 'Espoo', lat: 60.2055, lon: 24.6559 },
  { name: 'Vantaa', lat: 60.2934, lon: 25.0378 },
  { name: 'Turku', lat: 60.4518, lon: 22.2666 },
  { name: 'Tampere', lat: 61.4981, lon: 23.7610 },
  { name: 'Oulu', lat: 65.0121, lon: 25.4651 },
  { name: 'Lahti', lat: 60.9827, lon: 25.6615 },
  { name: 'Kuopio', lat: 62.8924, lon: 27.6770 },
  { name: 'Pori', lat: 61.4850, lon: 21.7973 },
  { name: 'Jyväskylä', lat: 62.2426, lon: 25.7473 },
];

const interestsOptions = ["кино","спорт","музыка","технологии","искусство"];
const hobbiesOptions   = ["чтение","бег","рисование","игры","готовка"];
const musicOptions     = ["рок","джаз","классика","поп","хип-хоп"];
const foodOptions      = ["итальянская","азиатская","русская","французская","мексиканская"];
const travelOptions    = ["пляж","горы","города","экспедиции","экотуризм"];

const Recommendations = () => {
  const navigate = useNavigate();

  // UI state
  const [mode, setMode] = useState('affinity');
  const [useProfileFilters, setUseProfileFilters] = useState(true);
  const [form, setForm] = useState({
    city: cityOptions[0],
    interests: [], priorityInterests: false,
    hobbies:   [], priorityHobbies: false,
    music:     [], priorityMusic: false,
    food:      [], priorityFood: false,
    travel:    [], priorityTravel: false,
    lookingFor: ''
  });

  // данные рекомендаций
  const [recommendations, setRecommendations] = useState([]);
  const [loading, setLoading] = useState(false);
  const [decliningId, setDecliningId] = useState(null);
  

  // // при смене режима очищаем список
  // const handleModeChange = (_, newMode) => {
  //   if (newMode && newMode !== mode) {
  //     setMode(newMode);
  //     setRecommendations([]);
  //   }
  // };

  // Обработчик формы поиска
  const handleSearch = async e => {
    e.preventDefault();
    setLoading(true);

    //const params = { mode, withDistance: true };
    const params = { mode, withDistance: true, useProfile: useProfileFilters };

    if (!useProfileFilters) {
      // Всегда передаём координаты
      params.cityLat = form.city.lat;
      params.cityLon = form.city.lon;

      if (mode === 'affinity') {
        params.interests         = form.interests.join(',');
        params.priorityInterests = form.priorityInterests;
        params.hobbies           = form.hobbies.join(',');
        params.priorityHobbies   = form.priorityHobbies;
        params.music             = form.music.join(',');
        params.priorityMusic     = form.priorityMusic;
        params.food              = form.food.join(',');
        params.priorityFood      = form.priorityFood;
        params.travel            = form.travel.join(',');
        params.priorityTravel    = form.priorityTravel;
      } else {
        params.lookingFor = form.lookingFor;
      }
    }

    try {
      const recs = await getRecommendations({ params });
      // const recData = await Promise.all(
      //   recs.map(async ({ id, distance, score }) => {
      //     try {
      //       const user = await getUser(id);
      //       const bio  = await getUserBio(id);
      //       return { id, distance, score, ...user, bio };
      //     } catch {
      //       return null;
      //     }
      //   })
      // );
      const recData = await Promise.all(
        recs.map(async ({ id, distance, score }) => {
          try {
            const user = await getUser(id);
            const bio  = await getUserBio(id);
            return { id, distance, score, ...user, bio };
          } catch (err) {
            console.error(`[ERROR] Failed to load user ${id}:`, err);
            return null;
          }
        })
      );
      
      setRecommendations(recData.filter(r => r));
    } catch (err) {
      const msg = err.response?.data || 'Ошибка загрузки рекомендаций';
      toast.error(msg);
      if (/заполните/i.test(msg)) {
        setTimeout(() => navigate('/edit-profile'), 2000);
      }
    } finally {
      setLoading(false);
    }
  };

  // Отклонить рекомендацию
  const handleDecline = async id => {
    setDecliningId(id);
    try {
      await declineRecommendation(id);
      setRecommendations(prev => prev.filter(r => r.id !== id));
      toast.success('Рекомендация отклонена');
    } catch {
      toast.error('Ошибка при отклонении');
    } finally {
      setDecliningId(null);
    }
  };
  // Отправить запрос на подключение
  const handleConnect = async id => {
    try {
      await sendConnectionRequest(id);
      toast.success('Запрос отправлен');
      setRecommendations(prev => prev.filter(r => r.id !== id));
    } catch {
      toast.error('Ошибка при запросе');
    }
  };

  // переключаем режим и сбрасываем рекомендации
  const switchMode = (newMode) => {
    if (newMode !== mode) {
      setMode(newMode);
      setRecommendations([]);
    }
  };
  
  return (
    <Container sx={{ mt: 4 }}>
      {/* Кнопки режима */}
      <Box sx={{ mb: 2, display: 'flex', alignItems: 'center', gap: 2 }}>
        <Button
          variant={mode === 'affinity' ? 'contained' : 'outlined'}
          onClick={() => switchMode('affinity')}
        >
          AffinityMatch
        </Button>
        <Button
          variant={mode === 'desire' ? 'contained' : 'outlined'}
          onClick={() => switchMode('desire')}
        >
          DesireMatch
        </Button>

        {/* Чекбокс "Использовать данные профиля" */}
        <FormControlLabel
          control={
            <Checkbox
              checked={useProfileFilters}
              onChange={e => setUseProfileFilters(e.target.checked)}
            />
          }
          label="Использовать данные профиля"
          sx={{ ml: 2 }}
        />
      </Box>

      <Typography variant="h4" gutterBottom>Рекомендации</Typography>

      {/* Форма фильтров */}
      <Box component="form" onSubmit={handleSearch} sx={{ mb: 4 }}>
        {/* Город */}
        <FormControl sx={{ minWidth: 200, mr: 2 }}
        disabled={useProfileFilters}>
          <InputLabel>Город</InputLabel>
          <Select
            value={form.city.name}
            label="Город"
            onChange={e => {
              const sel = cityOptions.find(c => c.name === e.target.value);
              setForm(f => ({ ...f, city: sel }));
            }}
          >
            {cityOptions.map(c => (
              <MenuItem key={c.name} value={c.name}>{c.name}</MenuItem>
            ))}
          </Select>
        </FormControl>

        {mode === 'affinity' ? (
          <>
            {[
              ['Интересы', 'interests', interestsOptions, 'priorityInterests'],
              ['Хобби',    'hobbies',   hobbiesOptions,   'priorityHobbies'],
              ['Музыка',   'music',     musicOptions,     'priorityMusic'],
              ['Еда',      'food',      foodOptions,      'priorityFood'],
              ['Путешествия','travel',  travelOptions,    'priorityTravel']
            ].map(([label, key, opts, prioKey]) => (
              <FormControl key={key} sx={{ minWidth: 200, mr: 2, mt: 2 }}
              disabled={useProfileFilters}>
                <InputLabel>{label}</InputLabel>
                <Select
                  multiple
                  value={form[key]}
                  onChange={e => setForm(f => ({ ...f, [key]: e.target.value }))}
                  renderValue={selected => selected.join(', ')}
                  label={label}
                >
                  {opts.map(opt => (
                    <MenuItem key={opt} value={opt}>
                      <Checkbox checked={form[key].includes(opt)} 
                      disabled={useProfileFilters}/>
                      <ListItemText primary={opt} />
                    </MenuItem>
                  ))}
                </Select>
                <Box sx={{ display: 'flex', alignItems: 'center', mt: 1 }}>
                  <Checkbox
                    checked={form[prioKey]}
                    onChange={e => setForm(f => ({ ...f, [prioKey]: e.target.checked }))}
                    disabled={useProfileFilters}/>
                  <Typography variant="body2">Priority</Typography>
                </Box>
              </FormControl>
            ))}
          </>
        ) : (
          <TextField
            label="Кого вы ищете"
            value={form.lookingFor}
            onChange={e => setForm(f => ({ ...f, lookingFor: e.target.value }))}
            disabled={useProfileFilters}
            sx={{ minWidth: 300, mr: 2 }}
          />
        )}

        <Button
          type="submit"
          variant="contained"
          sx={{ mt: 2 }}
          disabled={loading}
        >
          Искать
        </Button>
      </Box>

      {/* Результаты */}
      {loading ? (
        <Box sx={{ textAlign: 'center', mt: 4 }}>
          <CircularProgress />
        </Box>
      ) : recommendations.length === 0 ? (
        <Typography>Нет доступных рекомендаций.</Typography>
      ) : (
        <Grid container spacing={3}>
          {recommendations.map(rec => (
            <Grid item xs={12} sm={6} md={4} key={rec.id}>
              <Card>
                <CardMedia
                  component="img"
                  height="140"
                  image={rec.photoUrl || '/static/images/default.png'}
                  alt={`${rec.firstName} ${rec.lastName}`}
                />
                <CardContent>
                  <Typography variant="h5">
                    {rec.firstName} {rec.lastName}
                  </Typography>
                  {typeof rec.distance === 'number' && (
                    <Typography variant="body2" color="text.secondary">
                      Расстояние: {rec.distance.toFixed(1)} км
                    </Typography>
                  )}
                  {typeof rec.score === 'number' && (
                    <Typography variant="body2" color="text.secondary">
                      Совпадение: {(rec.score * 100).toFixed(0)} %
                    </Typography>
                  )}
                  <Typography variant="body2" color="text.secondary">
                    {rec.bio.interests
                      ? `Интересы: ${rec.bio.interests}`
                      : 'Информация отсутствует'}
                  </Typography>
                </CardContent>
                <CardActions>
                  <Button size="small" variant="contained" onClick={() => handleConnect(rec.id)}>
                    Подключиться
                  </Button>
                  <Button
                    size="small"
                    variant="outlined"
                    onClick={() => handleDecline(rec.id)}
                    disabled={decliningId === rec.id}
                  >
                    {decliningId === rec.id ? 'Отклонение…' : 'Отклонить'}
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

export default Recommendations;
