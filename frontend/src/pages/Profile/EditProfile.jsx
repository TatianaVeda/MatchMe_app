// src/pages/Profile/EditProfile.jsx
import React, { useState, useEffect } from 'react';
import { Container, Box, Typography, TextField, Button, CircularProgress, 
  FormControl, InputLabel, Select, MenuItem, FormControlLabel, Switch } from '@mui/material';
import api from '../../api/index';
import { useNavigate } from 'react-router-dom';
import { Formik, Form, Field, ErrorMessage } from 'formik';
import * as Yup from 'yup';
import { getMyProfile, getMyBio, getMyPreferences, updateMyProfile, updateMyBio, deleteMyPhoto } from '../../api/user';
import { toast } from 'react-toastify';


 const cityOptions = [
     { name: 'Helsinki', lat: 60.1699, lon: 24.9384 },
     { name: 'Espoo',    lat: 60.2055, lon: 24.6559 },
     { name: 'Vantaa',   lat: 60.2934, lon: 25.0378 },
     { name: 'Turku',    lat: 60.4518, lon: 22.2666 },
     { name: 'Tampere',  lat: 61.4981, lon: 23.7610 },
     { name: 'Oulu',     lat: 65.0121, lon: 25.4651 },
     { name: 'Lahti',    lat: 60.9827, lon: 25.6615 },
     { name: 'Kuopio',   lat: 62.8924, lon: 27.6770 },
     { name: 'Pori',     lat: 61.4850, lon: 21.7973 },
     { name: 'Jyväskylä',lat: 62.2426, lon: 25.7473 },
   ];

const interestsOptions = ["кино","спорт","музыка","технологии","искусство"];
const hobbiesOptions   = ["чтение","бег","рисование","игры","готовка"];
const musicOptions     = ["рок","джаз","классика","поп","хип-хоп"];
const foodOptions      = ["итальянская","азиатская","русская","французская","мексиканская"];
const travelOptions    = ["пляж","горы","города","экспедиции","экотуризм"];

const EditProfileSchema = Yup.object().shape({
  firstName: Yup.string().max(255, 'Имя слишком длинное').required('Укажите имя'),
  lastName: Yup.string().max(255, 'Фамилия слишком длинная').required('Укажите фамилию'),
  about: Yup.string().max(1000, 'Описание слишком длинное'),

  city: Yup.object({
    name: Yup.string().required(),
    lat: Yup.string().required(),
    lon: Yup.string().required()
  }).required('Выберите город'),
  interests: Yup.array().min(1, 'Укажите хотя бы один интерес'),
  hobbies:   Yup.array().min(1, 'Укажите хотя бы одно хобби'),
  music:     Yup.array().min(1, 'Укажите хотя бы один жанр музыки'),
  food:      Yup.array().min(1, 'Укажите хотя бы одну кухню'),
  travel:    Yup.array().min(1, 'Укажите хотя бы один тип путешествий'),
  lookingFor: Yup.string().required('Укажите, кого вы ищете')  // новое обязательное поле
});

const EditProfile = () => {
  const navigate = useNavigate();
  const [initialValues, setInitialValues] = useState(null);
  const [photoFile, setPhotoFile] = useState(null);
  const [uploading, setUploading] = useState(false);

  useEffect(() => {
  
    const loadData = async () => {
      try {
        const [profileRaw, bioRaw, prefsRaw] = await Promise.all([
          getMyProfile().catch(() => null),
          getMyBio().catch(() => null),
          getMyPreferences().catch(() => null),
        ]);
    
        const profile = profileRaw || {};
        const bio = bioRaw || {};
        const prefs = prefsRaw || {};
    
        setInitialValues({
          firstName: profile.firstName || '',
          lastName: profile.lastName || '',
          about: profile.about || '',
          city: cityOptions.find(c => c.name === profile.city) || {
            name:  profile.city || cityOptions[0].name,
            lat:   profile.latitude  || cityOptions[0].lat,
            lon:   profile.longitude || cityOptions[0].lon,
          },
          interests: bio.interests ? bio.interests.split(' ') : [],
          hobbies:   bio.hobbies   ? bio.hobbies.split(' ')   : [],
          music:     bio.music     ? bio.music.split(' ')     : [],
          food:      bio.food      ? bio.food.split(' ')      : [],
          travel:    bio.travel    ? bio.travel.split(' ')    : [],
          lookingFor: bio.lookingFor || '',
          priorityInterests: prefs.priorityInterests || false,
          priorityHobbies:   prefs.priorityHobbies   || false,
          priorityMusic:     prefs.priorityMusic     || false,
          priorityFood:      prefs.priorityFood      || false,
          priorityTravel:    prefs.priorityTravel    || false,
        });
      } catch {
        toast.error('Ошибка загрузки данных профиля');
      }
    };
    
    loadData();
  }, []);

  if (!initialValues) {
    return (
      <Container sx={{ textAlign: 'center', mt: 4 }}>
        <CircularProgress />
      </Container>
    );
  }

  const handlePhotoChange = e => setPhotoFile(e.target.files[0]);

  const handlePhotoUpload = async () => {
    if (!photoFile) return;
    setUploading(true);
    try {
      const formData = new FormData();
      formData.append('photo', photoFile);
      await api.post('/me/photo', formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
      });
      toast.success('Фото успешно загружено');
    } catch {
      toast.error('Ошибка при загрузке фото');
    } finally {
      setUploading(false);
    }
  };

  
  const handlePhotoDelete = async () => {
    setUploading(true);
    try {
      await deleteMyPhoto();             
      toast.success('Фото удалено');
      navigate(0);                       
    } catch {
      toast.error('Ошибка при удалении фото');
    } finally {
      setUploading(false);
    }
  };

  const handleSubmit = async (values, { setSubmitting }) => {
    try {
      let latitude = values.city.lat;
      let longitude = values.city.lon;

      if (!latitude || !longitude) {
        const fallbackCity = cityOptions.find(c => c.name === values.city.name);
        latitude = fallbackCity?.lat;
        longitude = fallbackCity?.lon;
      }

      await updateMyProfile({
                firstName: values.firstName,
                lastName:  values.lastName,
                about:     values.about,
                city:      values.city.name,
                latitude:  values.city.lat,
                longitude: values.city.lon
              });
      await updateMyBio({
        interests: values.interests.join(' '),
        hobbies:   values.hobbies.join(' '),
        music:     values.music.join(' '),
        food:      values.food.join(' '),
        travel:    values.travel.join(' '),
        lookingFor: values.lookingFor,  
        priorityInterests:   values.priorityInterests,
        priorityHobbies:     values.priorityHobbies,
        priorityMusic:       values.priorityMusic,
        priorityFood:        values.priorityFood,
        priorityTravel:      values.priorityTravel,
      });
      toast.success('Профиль успешно обновлён');
      navigate('/me');
    } catch (err) {
      toast.error(err.response?.data?.message || 'Ошибка при сохранении');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Container maxWidth="sm" sx={{ mt: 4 }}>
      <Box sx={{ p: 3, border: '1px solid #ccc', borderRadius: 2 }}>
        <Typography variant="h4" gutterBottom>
          Редактировать профиль
        </Typography>

        <Box sx={{ mb: 2 }}>
          <Typography variant="subtitle1">Загрузить фото</Typography>
          <input
            type="file"
            accept="image/jpeg,image/png"
            onChange={e => setPhotoFile(e.target.files[0])}
            disabled={uploading}
          />
          <Button
            variant="contained"
            onClick={handlePhotoUpload}
            disabled={!photoFile || uploading}
            sx={{ ml: 1 }}
          >
            Загрузить
          </Button>
          <Button
            variant="outlined"
            color="error"
            onClick={handlePhotoDelete}
            disabled={uploading}
            sx={{ ml: 1 }}
          >
            Удалить фото
          </Button>
          {uploading && <Typography variant="body2">Загрузка...</Typography>}
        </Box>

        <Box sx={{ mb: 3 }}>
          <Typography variant="h6">Местоположение</Typography>
          <Button
            variant="outlined"
            fullWidth
            sx={{ mt: 1 }}
            onClick={() => {
              if (!navigator.geolocation) {
                toast.error('Геолокация не поддерживается');
                return;
              }
              
              navigator.geolocation.getCurrentPosition(
                ({ coords }) => {
                  api.put('/me/location', {
                    latitude: coords.latitude,
                    longitude: coords.longitude
                  })
                  .then(() => toast.success('Локация сохранена'))
                  .catch(() => toast.error('Не удалось сохранить координаты'));
                },
                () => toast.error('Не удалось получить местоположение')
              );              
            }}
          >
            Использовать моё местоположение
          </Button>
        </Box>

        <Formik
          initialValues={initialValues}
          validationSchema={EditProfileSchema}
          onSubmit={handleSubmit}
        >
          {({ isSubmitting, touched, errors }) => (
            <Form>
              <Typography variant="h6">Профиль</Typography>
              <Field
                name="firstName"
                as={TextField}
                label="Имя"
                fullWidth
                margin="normal"
                error={touched.firstName && Boolean(errors.firstName)}
                helperText={<ErrorMessage name="firstName" />}
              />
              <Field
                name="lastName"
                as={TextField}
                label="Фамилия"
                fullWidth
                margin="normal"
                error={touched.lastName && Boolean(errors.lastName)}
                helperText={<ErrorMessage name="lastName" />}
              />
              <Field
                name="about"
                as={TextField}
                label="О себе"
                fullWidth
                margin="normal"
                multiline
                rows={3}
                error={touched.about && Boolean(errors.about)}
                helperText={<ErrorMessage name="about" />}
              />

     <FormControl fullWidth margin="normal" error={touched.city && Boolean(errors.city)}>
       <InputLabel id="city-label">Город</InputLabel>
       <Field name="city">
         {({ field, form }) => (
           <Select
             {...field}
             labelId="city-label"
             label="Город"
             value={field.value.name}
             onChange={e => {
               const sel = cityOptions.find(c => c.name === e.target.value);
               form.setFieldValue('city', sel);
             }}
           >
             {cityOptions.map(c => (
               <MenuItem key={c.name} value={c.name}>
                 {c.name}
               </MenuItem>
             ))}
           </Select>
         )}
       </Field>
       <ErrorMessage name="city" component="div" style={{ color: 'red' }} />       
     </FormControl>

              <Typography variant="h6" sx={{ mt: 3 }}>
                Биография
              </Typography>

     <FormControl fullWidth margin="normal" error={touched.interests && Boolean(errors.interests)}>
       <InputLabel id="interests-label">Интересы</InputLabel>
       <Field name="interests">
    {({ field, form }) => (
      <Select
        {...field}
        multiple
        labelId="interests-label"
        label="Интересы"
        value={field.value}
        onChange={e => form.setFieldValue('interests', e.target.value)}
      >
        {interestsOptions.map(opt => (
          <MenuItem key={opt} value={opt}>{opt}</MenuItem>
        ))}
      </Select>
    )}
  </Field>
       <FormControlLabel
         control={<Field name="priorityInterests" as={Switch} />}
         label="Приоритетные интересы"
       />
       <ErrorMessage name="interests" component="div" style={{ color: 'red' }} />
     </FormControl>

     <FormControl fullWidth margin="normal" error={touched.hobbies && Boolean(errors.hobbies)}>
  <InputLabel id="hobbies-label">Хобби</InputLabel>
  <Field name="hobbies">
    {({ field, form }) => (
      <Select
        {...field}
        multiple
        labelId="hobbies-label"
        label="Хобби"
        value={field.value}
        onChange={e => form.setFieldValue('hobbies', e.target.value)}
      >
        {hobbiesOptions.map(opt => (
          <MenuItem key={opt} value={opt}>{opt}</MenuItem>
        ))}
      </Select>
    )}
  </Field>
  <FormControlLabel
    control={<Field name="priorityHobbies" as={Switch} />}
    label="Приоритетное хобби"
  />
  <ErrorMessage name="hobbies" component="div" style={{ color: 'red' }} />
</FormControl>

     <FormControl fullWidth margin="normal" error={touched.music && Boolean(errors.music)}>
  <InputLabel id="music-label">Музыка</InputLabel>
  <Field name="music">
    {({ field, form }) => (
      <Select
        {...field}
        multiple
        labelId="music-label"
        label="Музыка"
        value={field.value}
        onChange={e => form.setFieldValue('music', e.target.value)}
      >
        {musicOptions.map(opt => (
          <MenuItem key={opt} value={opt}>{opt}</MenuItem>
        ))}
      </Select>
    )}
  </Field>
  <FormControlLabel
    control={<Field name="priorityMusic" as={Switch} />}
    label="Приоритетная музыка"
  />
  <ErrorMessage name="music" component="div" style={{ color: 'red' }} />
</FormControl>

     <FormControl fullWidth margin="normal" error={touched.food && Boolean(errors.food)}>
  <InputLabel id="food-label">Еда</InputLabel>
  <Field name="food">
    {({ field, form }) => (
      <Select
        {...field}
        multiple
        labelId="food-label"
        label="Еда"
        value={field.value}
        onChange={e => form.setFieldValue('food', e.target.value)}
      >
        {foodOptions.map(opt => (
          <MenuItem key={opt} value={opt}>{opt}</MenuItem>
        ))}
      </Select>
    )}
  </Field>
  <FormControlLabel
    control={<Field name="priorityFood" as={Switch} />}
    label="Приоритетная еда"
  />
  <ErrorMessage name="food" component="div" style={{ color: 'red' }} />
</FormControl>

     <FormControl fullWidth margin="normal" error={touched.travel && Boolean(errors.travel)}>
  <InputLabel id="travel-label">Путешествия</InputLabel>
  <Field name="travel">
    {({ field, form }) => (
      <Select
        {...field}
        multiple
        labelId="travel-label"
        label="Путешествия"
        value={field.value}
        onChange={e => form.setFieldValue('travel', e.target.value)}
      >
        {travelOptions.map(opt => (
          <MenuItem key={opt} value={opt}>{opt}</MenuItem>
        ))}
      </Select>
    )}
  </Field>
  <FormControlLabel
    control={<Field name="priorityTravel" as={Switch} />}
    label="Приоритетные путешествия"
  />
  <ErrorMessage name="travel" component="div" style={{ color: 'red' }} />
</FormControl>
              
              <Field
                name="lookingFor"
                as={TextField}
                label="Кого вы ищете"
                fullWidth
                margin="normal"
                error={touched.lookingFor && Boolean(errors.lookingFor)}
                helperText={<ErrorMessage name="lookingFor" />}
              />

              <Button
                variant="contained"
                color="primary"
                type="submit"
                fullWidth
                sx={{ mt: 2 }}
                disabled={isSubmitting}
              >
                {isSubmitting ? 'Сохранение...' : 'Сохранить изменения'}
              </Button>
            </Form>
          )}
        </Formik>
      </Box>
    </Container>
  );
};

export default EditProfile;
