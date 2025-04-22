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
import {
  Container,
  Box,
  Typography,
  TextField,
  Button,
  CircularProgress
} from '@mui/material';
import api from '../../api/index';
import { useNavigate } from 'react-router-dom';
import { Formik, Form, Field, ErrorMessage } from 'formik';
import * as Yup from 'yup';
import {
  getMyProfile,
  getMyBio,
  updateMyProfile,
  updateMyBio
} from '../../api/user';
import { toast } from 'react-toastify';

const EditProfileSchema = Yup.object().shape({
  firstName: Yup.string()
    .max(255, 'Имя слишком длинное')
    .required('Укажите имя'),
  lastName: Yup.string()
    .max(255, 'Фамилия слишком длинная')
    .required('Укажите фамилию'),
  about: Yup.string().max(1000, 'Описание слишком длинное'),
  interests: Yup.string().required('Укажите интересы'),
  hobbies: Yup.string().required('Укажите хобби'),
  music: Yup.string().required('Укажите музыку'),
  food: Yup.string().required('Укажите еду'),
  travel: Yup.string().required('Укажите путешествия')
});

const EditProfile = () => {
  const navigate = useNavigate();
  const [initialValues, setInitialValues] = useState(null);

  // Для загрузки фото
  const [photoFile, setPhotoFile] = useState(null);
  const [uploading, setUploading] = useState(false);

  useEffect(() => {
    const loadData = async () => {
      try {
        const profile = await getMyProfile();
        const bio = await getMyBio();
        setInitialValues({
          firstName: profile.firstName || '',
          lastName: profile.lastName || '',
          about: profile.about || '',
          interests: bio.interests || '',
          hobbies: bio.hobbies || '',
          music: bio.music || '',
          food: bio.food || '',
          travel: bio.travel || ''
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
      await updateMyProfile({
        firstName: values.firstName,
        lastName: values.lastName,
        about: values.about
      });
      await updateMyBio({
        interests: values.interests,
        hobbies: values.hobbies,
        music: values.music,
        food: values.food,
        travel: values.travel
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

        {/* Блок загрузки фото */}
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

        {/* Блок геолокации */}
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

              <Typography variant="h6" sx={{ mt: 3 }}>
                Биография
              </Typography>
              <Field
                name="interests"
                as={TextField}
                label="Интересы"
                fullWidth
                margin="normal"
                error={touched.interests && Boolean(errors.interests)}
                helperText={<ErrorMessage name="interests" />}
              />
              <Field
                name="hobbies"
                as={TextField}
                label="Хобби"
                fullWidth
                margin="normal"
                error={touched.hobbies && Boolean(errors.hobbies)}
                helperText={<ErrorMessage name="hobbies" />}
              />
              <Field
                name="music"
                as={TextField}
                label="Музыка"
                fullWidth
                margin="normal"
                error={touched.music && Boolean(errors.music)}
                helperText={<ErrorMessage name="music" />}
              />
              <Field
                name="food"
                as={TextField}
                label="Еда"
                fullWidth
                margin="normal"
                error={touched.food && Boolean(errors.food)}
                helperText={<ErrorMessage name="food" />}
              />
              <Field
                name="travel"
                as={TextField}
                label="Путешествия"
                fullWidth
                margin="normal"
                error={touched.travel && Boolean(errors.travel)}
                helperText={<ErrorMessage name="travel" />}
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
