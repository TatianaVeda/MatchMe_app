// //m/frontend/src/pages/Recommendations.jsx

// import React, { useState, useEffect } from 'react';
// import { useNavigate } from 'react-router-dom';
// import { 
//   Container, 
//   Grid, 
//   Card, 
//   CardContent, 
//   CardMedia, 
//   Typography, 
//   Button, 
//   CardActions, 
//   CircularProgress,
//   Box,
//   ToggleButton,
//   ToggleButtonGroup,
//   FormControl,
//   InputLabel,
//   Select,
//   MenuItem,
//   Checkbox,
//   ListItemText,
//   OutlinedInput,
//   TextField,
// } from '@mui/material';
// import { toast } from 'react-toastify';

// // Импортируем методы из модулей API
// import { declineRecommendation } from '../api/recommendations';
// import { getUser, getUserBio } from '../api/user';
// import { sendConnectionRequest } from '../api/connections';
// import api from '../api/index';

// const Recommendations = () => {
//   const navigate = useNavigate(); 
//   // Состояние для хранения объединённых данных пользователя и его биографии
//   const [recommendations, setRecommendations] = useState([]);
//   const [loading, setLoading] = useState(true);
//   const [decliningId, setDecliningId] = useState(null); // <— отслеживаем, что отклоняем
//    // === новое состояние для формы ===
//   const [mode, setMode] = useState('affinity'); // 'affinity' или 'desire'
//   // города с координатами (как в EditProfile)
//   const cityOptions = [
//      { name: 'Helsinki', lat: 60.1699, lon: 24.9384 },
//      { name: 'Espoo',    lat: 60.2055, lon: 24.6559 },
//      { name: 'Vantaa',   lat: 60.2934, lon: 25.0378 },
//      { name: 'Turku',    lat: 60.4518, lon: 22.2666 },
//      { name: 'Tampere',  lat: 61.4981, lon: 23.7610 },
//      { name: 'Oulu',     lat: 65.0121, lon: 25.4651 },
//      { name: 'Lahti',    lat: 60.9827, lon: 25.6615 },
//      { name: 'Kuopio',   lat: 62.8924, lon: 27.6770 },
//      { name: 'Pori',     lat: 61.4850, lon: 21.7973 },
//      { name: 'Jyväskylä',lat: 62.2426, lon: 25.7473 },
//   ];
//   const [selectedCity, setSelectedCity] = useState(cityOptions[0]);

//   // для affinity-режима
//   const bioFields = ['Interests','Hobbies','Music','Food','Travel'];
//   const [selectedFields, setSelectedFields] = useState({
//     Interests: [], Hobbies: [], Music: [], Food: [], Travel: []
//   });
//   const [priority, setPriority] = useState({
//     Interests: false, Hobbies: false, Music: false, Food: false, Travel: false
//   });

//   // для desire-режима
//   const [desireText, setDesireText] = useState('');
//   // переключить режим
//   const handleModeChange = (e, m) => {
//     if (m) {
//       setMode(m);
//       setRecommendations([]);        // очистить старый список
//     }
//   };

//   // общая функция поиска
//   const fetchRecommendations = async () => {
//     setLoading(true);
//     try {
//       // формируем query
//       const params = new URLSearchParams({ mode, withDistance: 'true' });
//       const { data: recIds } = await api.get(`/recommendations?${params.toString()}`);
//       // далее как было — мапим на объекты
//       const recData = await Promise.all(recIds.map(async id => {
//         try {
//           const user = await getUser(id);
//           const bio  = await getUserBio(id);
//           return { id, ...user, bio };
//         } catch {
//          return null;
//         }
//       }));
//       setRecommendations(recData.filter(r => r));
//     } catch (err) {
//       toast.error('Ошибка загрузки рекомендаций');
//     } finally {
//       setLoading(false);
//     }
//   };


//   // Функция для получения списка рекомендованных идентификаторов и соответствующих данных
//   const fetchRecommendations = async () => {
//     try {
//       // Получаем массив идентификаторов с эндпоинта /recommendations
//       const recIds = await getRecommendations();
//       // Для каждого идентификатора последовательно загружаем данные пользователя и его биографию
//       const recData = await Promise.all(
//         recIds.map(async (id) => {
//           try {
//             const user = await getUser(id);
//             const bio = await getUserBio(id);
//             return { id, ...user, bio };
//           } catch (err) {
//             console.error("Ошибка загрузки данных для id", id, err);
//             return null;
//           }
//         })
//       );
//       // Фильтруем полученные данные – исключаем невалидные
//       setRecommendations(recData.filter((rec) => rec !== null));
//     } catch (err) {
//       // Если сервер вернул подробную ошибку (например, о незаполненном профиле/био)
//       const serverMessage = err?.response?.data || '';
//       const isIncompleteProfile = typeof serverMessage === 'string' && (
//         serverMessage.toLowerCase().includes('заполните') ||
//         serverMessage.toLowerCase().includes('профиль') ||
//         serverMessage.toLowerCase().includes('биографию')
//       );

//       if (isIncompleteProfile) {
//         toast.error(serverMessage);
//         setTimeout(() => navigate('/edit-profile'), 2000); // подождать 2 сек и перейти
//       } else {
//         toast.error("Ошибка загрузки рекомендаций");
//       }
//     } finally {
//       setLoading(false);
//     }
//   };

//   useEffect(() => {
//     fetchRecommendations();
//   }, []);

//   // Обработка кнопки "Отклонить" – убираем рекомендацию из списка
//   //Новая версия handleDecline
//   const handleDecline = async (id) => {
//    setDecliningId(id);
//     try {
//      await declineRecommendation(id);
//      // удаляем из списка только после успешного API
//      setRecommendations((prev) => prev.filter((rec) => rec.id !== id));
//      toast.success("Рекомендация отклонена");
//     } catch (err) {
//       toast.error("Ошибка при отклонении рекомендации");
//     } finally {
//      setDecliningId(null);
//     }
//   };
//   // Обработка кнопки "Подключиться" – отправляем запрос на подключение
//   const handleConnect = async (id) => {
//     try {
//       await sendConnectionRequest(id);
//       toast.success("Запрос на подключение отправлен");
//       setRecommendations((prev) => prev.filter((rec) => rec.id !== id));
//     } catch (err) {
//       toast.error("Ошибка при отправке запроса");
//     }
//   };

//   if (loading) {
//     return (
//       <Container sx={{ textAlign: 'center', mt: 4 }}>
//         <CircularProgress />
//       </Container>
//     );
//   }

//   // === UI формы поиска ===
//   const renderForm = () => (
//       <Box sx={{ mb: 3 }}>
//         {/* выбор города */}
//         <FormControl sx={{ mr: 2, minWidth: 200 }}>
//           <InputLabel>Город</InputLabel>
//           <Select
//             value={selectedCity.name}
//            input={<OutlinedInput label="Город" />}
//             onChange={e => {
//               const c = cityOptions.find(c=>c.name===e.target.value);
//               setSelectedCity(c);
//             }}
//           >
//             {cityOptions.map(c =>
//               <MenuItem key={c.name} value={c.name}>{c.name}</MenuItem>
//             )}
//           </Select>
//         </FormControl>
  
//         {mode === 'affinity' ? (
//           // мультивыбор для каждого поля
//           bioFields.map(field => (
//             <FormControl key={field} sx={{ mr: 2, minWidth: 160 }}>
//               <InputLabel>{field}</InputLabel>
//               <Select
//                 multiple
//                 value={selectedFields[field]}
//                 onChange={e => {
//                   setSelectedFields(fs => ({
//                     ...fs,
//                     [field]: e.target.value
//                   }));
//                 }}
//                 input={<OutlinedInput label={field} />}
//                 renderValue={vals => vals.join(', ')}
//               >
//                 { /* здесь можно взять options из констант, но для примера дублируем из Bio */ }
//               {selectedFields[field].length === 0 && <MenuItem disabled>Нет данных</MenuItem>}
//               </Select>
//               <Box>
//                 <Checkbox
//                   checked={priority[field]}
//                   onChange={e=>setPriority(p=>({...p,[field]:e.target.checked}))}
//                 /> Приоритет
//               </Box>
//           </FormControl>
//           ))
//         ) : (
//           // desire
//           <FormControl sx={{ minWidth: 300 }}>
//             <TextField
//               label="Кого вы ищете"
//               value={desireText}
//               onChange={e => setDesireText(e.target.value)}
//               fullWidth
//             />
//           </FormControl>
//         )}
  
//         <Button
//           variant="contained"
//           sx={{ ml: 2 }}
//           onClick={fetchRecommendations}
//         >
//         Искать
//         </Button>
//       </Box>
//     );
  
//     return (
//       <Container sx={{ mt: 4 }}>
//         <Typography variant="h4" gutterBottom>Рекомендации</Typography>
  
//         {/* переключатель режимов */}
//         <ToggleButtonGroup
//           value={mode}
//           exclusive
//           onChange={handleModeChange}
//           sx={{ mb: 2 }}
//         >
//           <ToggleButton value="affinity">AffinityMatch</ToggleButton>
//           <ToggleButton value="desire">DesireMatch</ToggleButton>
//         </ToggleButtonGroup>
  
//         {renderForm()}

//   if (recommendations.length === 0) {
//     return (
//       <Container sx={{ mt: 4 }}>
//         <Typography variant="h6">Нет доступных рекомендаций.</Typography>
//       </Container>
//     );
//   }

//   return (
//     <Container sx={{ mt: 4 }}>
//       <Typography variant="h4" gutterBottom>
//         Рекомендации
//       </Typography>
//       <Grid container spacing={3}>
//         {recommendations.map((rec) => (
//           <Grid item xs={12} sm={6} md={4} key={rec.id}>
//             <Card>
//               <CardMedia
//                 component="img"
//                 height="140"
//                 image={rec.photoUrl || '/static/images/default.png'}
//                 alt={`${rec.firstName} ${rec.lastName}`}
//               />
//               <CardContent>
//                 <Typography gutterBottom variant="h5" component="div">
//                   {rec.firstName} {rec.lastName}
//                 </Typography>
//                 <Typography variant="body2" color="text.secondary">
//                   {rec.bio && rec.bio.interests 
//                     ? `Интересы: ${rec.bio.interests}` 
//                     : "Информация отсутствует"}
//                 </Typography>
//               </CardContent>
//               <CardActions>
//                 <Button size="small" variant="contained" onClick={() => handleConnect(rec.id)}>
//                   Подключиться
//                 </Button>
//                 <Button size="small" variant="outlined" onClick={() => handleDecline(rec.id)}
//                 disabled={decliningId === rec.id}>
//                  {decliningId === rec.id ? 'Отклонение...' : 'Отклонить'}
//                 </Button>
//               </CardActions>
//             </Card>
//           </Grid>
//         ))}
//       </Grid>
//     </Container>
//   );
// };

// export default Recommendations;


import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Container, Grid, Card, CardContent, CardMedia,
  Typography, Button, CardActions, CircularProgress,
  ToggleButton, ToggleButtonGroup,
  Box, TextField, FormControl, InputLabel, Select, MenuItem,
  Checkbox, ListItemText
} from '@mui/material';
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
  

  // при смене режима очищаем список
  const handleModeChange = (_, newMode) => {
    if (newMode && newMode !== mode) {
      setMode(newMode);
      setRecommendations([]);
    }
  };

  // общая функция поиска (для обоих режимов)
  const fetchRecommendations = async () => {
    setLoading(true);
    try {
      const recs = await getRecommendations({ mode, withDistance: true });
      const recData = await Promise.all(
      recs.map(async ({ id, distance, score }) => {
          try {
            const user = await getUser(id);
            const bio  = await getUserBio(id);
            return { id, distance, score, ...user, bio };
          } catch {
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

  // отправка формы поиска
  const handleSearch = e => {
    e.preventDefault();
    fetchRecommendations();
  };

  // decline & connect
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
  const handleConnect = async id => {
    try {
      await sendConnectionRequest(id);
      toast.success('Запрос отправлен');
      setRecommendations(prev => prev.filter(r => r.id !== id));
    } catch {
      toast.error('Ошибка при запросе');
    }
  };

  return (
    <Container sx={{ mt: 4 }}>
      {/* переключатель режима */}
      <Box sx={{ mb: 2, display: 'flex', gap: 2 }}>
        <Button
          variant={mode === 'affinity' ? 'contained' : 'outlined'}
          onClick={() => { setMode('affinity'); setRecommendations([]); setLoading(true); fetchRecommendations(); }}
        >
          Affinity
        </Button>
        <Button
          variant={mode === 'desire' ? 'contained' : 'outlined'}
          onClick={() => { setMode('desire'); setRecommendations([]); setLoading(true); fetchRecommendations(); }}
        >
          Desire
        </Button>
      </Box>
      <Typography variant="h4" gutterBottom>Рекомендации</Typography>

      {/* Переключатель режимов */}
      <ToggleButtonGroup
        value={mode}
        exclusive
        onChange={handleModeChange}
        sx={{ mb: 3 }}
      >
        <ToggleButton value="affinity">AffinityMatch</ToggleButton>
        <ToggleButton value="desire">DesireMatch</ToggleButton>
      </ToggleButtonGroup>

      {/* Форма поиска */}
      <Box component="form" onSubmit={handleSearch} sx={{ mb: 4 }}>
        {/* Город */}
        <FormControl sx={{ minWidth: 200, mr: 2 }}>
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
            {/* 5 мультивыборов + чекбоксы */}
            {[
              ['Интересы', 'interests', interestsOptions, 'priorityInterests'],
              ['Хобби',    'hobbies',   hobbiesOptions,   'priorityHobbies'],
              ['Музыка',   'music',     musicOptions,     'priorityMusic'],
              ['Еда',      'food',      foodOptions,      'priorityFood'],
              ['Путешествия','travel',  travelOptions,    'priorityTravel']
            ].map(([label, key, opts, prioKey]) => (
              <FormControl key={key} sx={{ minWidth: 200, mr: 2, mt: 2 }}>
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
                      <Checkbox checked={form[key].includes(opt)} />
                      <ListItemText primary={opt} />
                    </MenuItem>
                  ))}
                </Select>
                <Box sx={{ display: 'flex', alignItems: 'center', mt: 1 }}>
                  <Checkbox
                    checked={form[prioKey]}
                    onChange={e => setForm(f => ({ ...f, [prioKey]: e.target.checked }))}
                  />
                  <Typography variant="body2">Priority</Typography>
                </Box>
              </FormControl>
            ))}
          </>
        ) : (
          /* mode === 'desire' */
          <TextField
            label="Кого вы ищете"
            value={form.lookingFor}
            onChange={e => setForm(f => ({ ...f, lookingFor: e.target.value }))}
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

      {/* Состояние загрузки */}
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
