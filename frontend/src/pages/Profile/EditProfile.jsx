// // /m/frontend/src/pages/Profile/EditProfile.jsx
// import React, { useState, useEffect } from 'react';
// import { Container, Box, Typography, TextField, Button, CircularProgress } from '@mui/material';
// import api from '../api/index'; // или from '../api/index'
// import { useNavigate } from 'react-router-dom';
// import { Formik, Form, Field, ErrorMessage } from 'formik';
// import * as Yup from 'yup';
// import { getMyProfile, getMyBio, updateMyProfile, updateMyBio } from '../api/user';
// import { toast } from 'react-toastify';

// const EditProfileSchema = Yup.object().shape({
//   firstName: Yup.string()
//     .max(255, 'Имя слишком длинное')
//     .required('Укажите имя'),
//   lastName: Yup.string()
//     .max(255, 'Фамилия слишком длинная')
//     .required('Укажите фамилию'),
//   about: Yup.string()
//     .max(1000, 'Описание слишком длинное'),
//   interests: Yup.string()
//     .required('Укажите интересы'),
//   hobbies: Yup.string()
//     .required('Укажите хобби'),
//   music: Yup.string(),
//   food: Yup.string(),
//   travel: Yup.string(),
// });

// const EditProfile = () => {
//   const navigate = useNavigate();
//   const [initialValues, setInitialValues] = useState(null);
//   const [photoFile, setPhotoFile] = useState(null);
//   const [uploading, setUploading] = useState(false);

//    // 1) Обработчик выбора файла
//    const handlePhotoChange = e => {
//     setPhotoFile(e.target.files[0]);
//   };

//   // 2) Кнопка «Загрузить»
//   const handlePhotoUpload = async () => {
//     if (!photoFile) return;
//     setUploading(true);
//     try {
//       const formData = new FormData();
//       formData.append('photo', photoFile);
//       await api.post('/me/photo', formData, {
//         headers: { 'Content-Type': 'multipart/form-data' },
//       });
//       toast.success('Фото успешно загружено');
//       // по желанию: рефрешнуть профиль
//     } catch {
//       toast.error('Ошибка при загрузке фото');
//     } finally {
//       setUploading(false);
//     }
//   };

//   useEffect(() => {
//     const loadData = async () => {
//       try {
//         const profile = await getMyProfile();
//         const bio = await getMyBio();
//         setInitialValues({
//           firstName: profile.firstname || '',
//           lastName:  profile.lastname  || '',
//           about:     profile.about      || '',
//           interests: bio.interests      || '',
//           hobbies:   bio.hobbies        || '',
//           music:     bio.music          || '',
//           food:      bio.food           || '',
//           travel:    bio.travel         || '',
//         });
//       } catch (err) {
//         toast.error('Ошибка загрузки данных профиля');
//       }
//     };
//     loadData();
//   }, []);

//   if (!initialValues) {
//     return (
//       <Container sx={{ textAlign: 'center', mt: 4 }}>
//         <CircularProgress />
//       </Container>
//     );
//   }

//   const handleSubmit = async (values, { setSubmitting }) => {
//     try {
//       await updateMyProfile({
//         firstname: values.firstName,
//         lastname:  values.lastName,
//         about:      values.about,
//       });
//       await updateMyBio({
//         interests: values.interests,
//         hobbies:   values.hobbies,
//         music:     values.music,
//         food:      values.food,
//         travel:    values.travel,
//       });
//       toast.success('Профиль успешно обновлён');
//       navigate('/me');
//     } catch (err) {
//       toast.error(err.response?.data?.message || 'Ошибка при сохранении');
//     } finally {
//       setSubmitting(false);
//     }
//   };

//   return (
//     <Container maxWidth="sm" sx={{ mt: 4 }}>
//       <Box sx={{ p: 3, border: '1px solid #ccc', borderRadius: 2 }}>
//         <Typography variant="h4" gutterBottom>
//           Редактировать профиль
//         </Typography>

//          {/* Загрузка фото */}
//         <Box sx={{ mb: 2 }}>
//           <Typography variant="subtitle1">Загрузить фото</Typography>
//           <input
//             type="file"
//            accept="image/jpeg,image/png"
//             onChange={handlePhotoChange}
//            disabled={uploading}
//           />
//           <Button
//           variant="contained"
//           onClick={handlePhotoUpload}
//           disabled={!photoFile || uploading}
//           sx={{ ml: 1 }}
//         >
//           Загрузить
//         </Button>
//          {uploading && <Typography variant="body2">Загрузка...</Typography>}
//        </Box>

//         <Formik
//           initialValues={initialValues}
//           validationSchema={EditProfileSchema}
//           onSubmit={handleSubmit}
//         >
//           {({ isSubmitting, touched, errors }) => (
//             <Form>
//               {/* Профиль */}
//               <Typography variant="h6">Профиль</Typography>
//               <Field
//                 name="firstName"
//                 as={TextField}
//                 label="Имя"
//                 fullWidth
//                 margin="normal"
//                 error={touched.firstName && Boolean(errors.firstName)}
//                 helperText={<ErrorMessage name="firstName" />}
//               />
//               <Field
//                 name="lastName"
//                 as={TextField}
//                 label="Фамилия"
//                 fullWidth
//                 margin="normal"
//                 error={touched.lastName && Boolean(errors.lastName)}
//                 helperText={<ErrorMessage name="lastName" />}
//               />
//               <Field
//                 name="about"
//                 as={TextField}
//                 label="О себе"
//                 fullWidth
//                 margin="normal"
//                 multiline
//                 rows={3}
//                 error={touched.about && Boolean(errors.about)}
//                 helperText={<ErrorMessage name="about" />}
//               />

//               {/* Биография */}
//               <Typography variant="h6" sx={{ mt: 3 }}>Биография</Typography>
//               <Field
//                 name="interests"
//                 as={TextField}
//                 label="Интересы"
//                 fullWidth
//                 margin="normal"
//                 error={touched.interests && Boolean(errors.interests)}
//                 helperText={<ErrorMessage name="interests" />}
//               />
//               <Field
//                 name="hobbies"
//                 as={TextField}
//                 label="Хобби"
//                 fullWidth
//                 margin="normal"
//                 error={touched.hobbies && Boolean(errors.hobbies)}
//                 helperText={<ErrorMessage name="hobbies" />}
//               />
//               <Field
//                 name="music"
//                 as={TextField}
//                 label="Музыка"
//                 fullWidth
//                 margin="normal"
//                 error={touched.music && Boolean(errors.music)}
//                 helperText={<ErrorMessage name="music" />}
//               />
//               <Field
//                 name="food"
//                 as={TextField}
//                 label="Еда"
//                 fullWidth
//                 margin="normal"
//                 error={touched.food && Boolean(errors.food)}
//                 helperText={<ErrorMessage name="food" />}
//               />
//               <Field
//                 name="travel"
//                 as={TextField}
//                 label="Путешествия"
//                 fullWidth
//                 margin="normal"
//                 error={touched.travel && Boolean(errors.travel)}
//                 helperText={<ErrorMessage name="travel" />}
//               />

//               <Button
//                 variant="contained"
//                 color="primary"
//                 type="submit"
//                 fullWidth
//                 sx={{ mt: 2 }}
//                 disabled={isSubmitting}
//               >
//                 {isSubmitting ? 'Сохранение...' : 'Сохранить изменения'}
//               </Button>
//             </Form>
//           )}
//         </Formik>
//       </Box>
//     </Container>
//   );
// };

// export default EditProfile;


// src/pages/Profile/EditProfile.jsx
import React, { useState, useEffect } from 'react';
import { Container, Box, Typography, TextField, Button, CircularProgress, 
  FormControl, InputLabel, Select, MenuItem, FormControlLabel, Switch } from '@mui/material';
import api from '../../api/index';
import { useNavigate } from 'react-router-dom';
import { Formik, Form, Field, ErrorMessage } from 'formik';
import * as Yup from 'yup';
import { getMyProfile, getMyBio, updateMyProfile, updateMyBio } from '../../api/user';
import { toast } from 'react-toastify';


// Новые константы с опциями 
const cityOptions = [
  "Helsinki","Espoo","Vantaa","Turku","Tampere",
  "Oulu","Lahti","Kuopio","Pori","Jyväskylä"
];

const interestsOptions = ["кино","спорт","музыка","технологии","искусство"];
const hobbiesOptions   = ["чтение","бег","рисование","игры","готовка"];
const musicOptions     = ["рок","джаз","классика","поп","хип-хоп"];
const foodOptions      = ["итальянская","азиатская","русская","французская","мексиканская"];
const travelOptions    = ["пляж","горы","города","экспедиции","экотуризм"];

// Добавлено поле lookingFor в схему валидации
const EditProfileSchema = Yup.object().shape({
  firstName: Yup.string().max(255, 'Имя слишком длинное').required('Укажите имя'),
  lastName: Yup.string().max(255, 'Фамилия слишком длинная').required('Укажите фамилию'),
  about: Yup.string().max(1000, 'Описание слишком длинное'),
  city: Yup.string().required('Выберите город'),
  interests: Yup.string().required('Укажите интересы'),
  hobbies: Yup.string().required('Укажите хобби'),
  music: Yup.string().required('Укажите музыку'),
  food: Yup.string().required('Укажите еду'),
  travel: Yup.string().required('Укажите путешествия'),
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
        const profile = await getMyProfile();
        const bio = await getMyBio();
        // Расширенные initialValues с lookingFor
        setInitialValues({
          firstName: profile.firstName || '',
          lastName: profile.lastName || '',
          about: profile.about || '',
          city:      profile.city || '',
          interests: bio.interests || '',
          hobbies: bio.hobbies || '',
          music: bio.music || '',
          food: bio.food || '',
          travel: bio.travel || '',
          lookingFor: bio.lookingFor || '',
          priorityInterests: false,
          priorityHobbies:   false,
          priorityMusic:     false,
          priorityFood:      false,
          priorityTravel:    false,
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

  const handleSubmit = async (values, { setSubmitting }) => {
    try {
      // Обновляем профиль
      await updateMyProfile({
        firstName: values.firstName,
        lastName: values.lastName,
        about: values.about,
        city:      values.city,
      });
      // Обновляем биографию, включая lookingFor
      await updateMyBio({
        interests: values.interests,
        hobbies: values.hobbies,
        music: values.music,
        food: values.food,
        travel: values.travel,
        lookingFor: values.lookingFor,  // сохраняем новое поле
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

        {/* Загрузка фото */}
        <Box sx={{ mb: 2 }}>
          <Typography variant="subtitle1">Загрузить фото</Typography>
          <input
            type="file"
            accept="image/jpeg,image/png"
            onChange={handlePhotoChange}
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
          {uploading && <Typography variant="body2">Загрузка...</Typography>}
        </Box>

        {/* Геолокация */}
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
                  api.put('/me/profile', {
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

               {/* Город */}
     <FormControl fullWidth margin="normal" error={touched.city && Boolean(errors.city)}>
       <InputLabel id="city-label">Город</InputLabel>
       <Field
         name="city"
         as={Select}
         labelId="city-label"
         label="Город"
       >
         {cityOptions.map(city => (
           <MenuItem key={city} value={city}>{city}</MenuItem>
         ))}
       </Field>
       <ErrorMessage name="city" component="div" style={{ color: 'red' }} />
     </FormControl>

              <Typography variant="h6" sx={{ mt: 3 }}>
                Биография
              </Typography>

              {/* Interests */}
     <FormControl fullWidth margin="normal" error={touched.interests && Boolean(errors.interests)}>
       <InputLabel id="interests-label">Интересы</InputLabel>
      <Field
         name="interests"
         as={Select}
         labelId="interests-label"
         label="Интересы"
       >
         {interestsOptions.map(opt => (
           <MenuItem key={opt} value={opt}>{opt}</MenuItem>
         ))}
       </Field>
       <FormControlLabel
         control={<Field name="priorityInterests" as={Switch} />}
         label="Приоритетные интересы"
       />
       <ErrorMessage name="interests" component="div" style={{ color: 'red' }} />
     </FormControl>

     {/* Hobbies */}
     <FormControl fullWidth margin="normal" error={touched.hobbies && Boolean(errors.hobbies)}>
       <InputLabel id="hobbies-label">Хобби</InputLabel>
       <Field name="hobbies" as={Select} labelId="hobbies-label" label="Хобби">
         {hobbiesOptions.map(opt => <MenuItem key={opt} value={opt}>{opt}</MenuItem>)}
       </Field>
       <FormControlLabel
         control={<Field name="priorityHobbies" as={Switch} />}
         label="Приоритетное хобби"
       />
       <ErrorMessage name="hobbies" component="div" style={{ color: 'red' }} />
     </FormControl>

     {/* Music */}
     <FormControl fullWidth margin="normal" error={touched.music && Boolean(errors.music)}>
       <InputLabel id="music-label">Музыка</InputLabel>
       <Field name="music" as={Select} labelId="music-label" label="Музыка">
         {musicOptions.map(opt => <MenuItem key={opt} value={opt}>{opt}</MenuItem>)}
       </Field>
       <FormControlLabel
         control={<Field name="priorityMusic" as={Switch} />}
         label="Приоритетная музыка"
       />
       <ErrorMessage name="music" component="div" style={{ color: 'red' }} />
     </FormControl>

     {/* Food */}
     <FormControl fullWidth margin="normal" error={touched.food && Boolean(errors.food)}>
       <InputLabel id="food-label">Еда</InputLabel>
       <Field name="food" as={Select} labelId="food-label" label="Еда">
         {foodOptions.map(opt => <MenuItem key={opt} value={opt}>{opt}</MenuItem>)}
       </Field>
       <FormControlLabel
         control={<Field name="priorityFood" as={Switch} />}
         label="Приоритетная еда"
       />
       <ErrorMessage name="food" component="div" style={{ color: 'red' }} />
     </FormControl>

     {/* Travel */}
     <FormControl fullWidth margin="normal" error={touched.travel && Boolean(errors.travel)}>
       <InputLabel id="travel-label">Путешествия</InputLabel>
       <Field name="travel" as={Select} labelId="travel-label" label="Путешествия">
         {travelOptions.map(opt => <MenuItem key={opt} value={opt}>{opt}</MenuItem>)}
       </Field>
       <FormControlLabel
         control={<Field name="priorityTravel" as={Switch} />}
         label="Приоритетные путешествия"
       />
       <ErrorMessage name="travel" component="div" style={{ color: 'red' }} />
     </FormControl>
              
              {/* Поле кого ищу */}
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
