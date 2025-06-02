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

const interestsOptions = ["movies", "sports", "music", "technology", "art"];
const hobbiesOptions   = ["reading", "running", "drawing", "gaming", "cooking"];
const musicOptions     = ["rock", "jazz", "classical", "pop", "hip-hop"];
const foodOptions      = ["italian", "asian", "russian", "french", "mexican"];
const travelOptions    = ["beach", "mountains", "cities", "expeditions", "ecotourism"];

/**
 * EditProfile.jsx
 *
 * User profile editing page.
 * Implements forms for profile, bio, preferences, photo upload/delete.
 * Uses Formik, Yup, batch data loading, API integration, and validation.
 */
const EditProfileSchema = Yup.object().shape({
  // Yup validation schema for profile editing form
  firstName: Yup.string().max(255, 'First name is too long').required('Enter first name'),
  lastName: Yup.string().max(255, 'Last name is too long').required('Enter last name'),
  about: Yup.string().max(1000, 'Description is too long, max 1000 characters'),

  city: Yup.object({
    name: Yup.string().required(),
    lat: Yup.string().required(),
    lon: Yup.string().required()
  }).required('Select a city'),
  interests: Yup.array().min(1, 'Select at least one interest'),
  hobbies:   Yup.array().min(1, 'Select at least one hobby'),
  music:     Yup.array().min(1, 'Select at least one music genre'),
  food:      Yup.array().min(1, 'Select at least one cuisine'),
  travel:    Yup.array().min(1, 'Select at least one travel type'),
  lookingFor: Yup.string().required('Specify who you are looking for')
});

const EditProfile = () => {
  const navigate = useNavigate();
  const [initialValues, setInitialValues] = useState(null);
  const [photoFile, setPhotoFile] = useState(null);
  const [uploading, setUploading] = useState(false);

  useEffect(() => {
  /**
   * Batch loads profile, bio, and preferences on mount.
   * Handles loading errors and sets initial form values.
   */
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
        toast.error('Error loading profile data');
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

  //const handlePhotoChange = e => setPhotoFile(e.target.files[0]);

  const handlePhotoUpload = async () => {
    /**
     * Uploads a new user photo using multipart/form-data.
     * Handles errors and updates uploading state.
     */
    if (!photoFile) return;
    setUploading(true);
    try {
      const formData = new FormData();
      formData.append('photo', photoFile);
      await api.post('/me/photo', formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
      });
      toast.success('Photo uploaded successfully');
    } catch {
      toast.error('Error uploading photo');
    } finally {
      setUploading(false);
    }
  };

  
  const handlePhotoDelete = async () => {
    /**
     * Deletes the user photo via API, updates state and reloads the page.
     * Handles errors and uploading state.
     */
    setUploading(true);
    try {
      await deleteMyPhoto();             
      toast.success('Photo deleted');
      navigate(0);                       
    } catch {
      toast.error('Error deleting photo');
    } finally {
      setUploading(false);
    }
  };

  const handleSubmit = async (values, { setSubmitting }) => {
    /**
     * Saves changes to profile, bio, and preferences.
     * Integrates with API, handles errors, updates state.
     */
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
      toast.success('Profile updated successfully');
      navigate('/me');
    } catch (err) {
      toast.error(err.response?.data?.message || 'Error saving changes');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Container maxWidth="sm" sx={{ mt: 4 }}>
      <Box sx={{ p: 3, border: '1px solid #ccc', borderRadius: 2 }}>
        <Typography variant="h4" gutterBottom>
          Edit Profile
        </Typography>

        <Box sx={{ mb: 2 }}>
          <Typography variant="subtitle1">Upload Photo</Typography>
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
            Upload
          </Button>
          <Button
            variant="outlined"
            color="error"
            onClick={handlePhotoDelete}
            disabled={uploading}
            sx={{ ml: 1 }}
          >
            Delete Photo
          </Button>
          {uploading && <Typography variant="body2">Uploading...</Typography>}
        </Box>

        <Box sx={{ mb: 3 }}>
          <Typography variant="h6">Location</Typography>
          <Button
            variant="outlined"
            fullWidth
            sx={{ mt: 1 }}
            onClick={() => {
              if (!navigator.geolocation) {
                toast.error('Geolocation is not supported');
                return;
              }
              
              navigator.geolocation.getCurrentPosition(
                ({ coords }) => {
                  api.put('/me/location', {
                    latitude: coords.latitude,
                    longitude: coords.longitude
                  })
                  .then(() => toast.success('Location saved'))
                  .catch(() => toast.error('Failed to save coordinates'));
                },
                () => toast.error('Failed to get location')
              );              
            }}
          >
            Use My Location
          </Button>
        </Box>

        <Formik
          initialValues={initialValues}
          validationSchema={EditProfileSchema}
          onSubmit={handleSubmit}
        >
          {({ isSubmitting, touched, errors }) => (
            <Form>
              <Typography variant="h6">Profile</Typography>
              <Field
                name="firstName"
                as={TextField}
                label="First Name"
                fullWidth
                margin="normal"
                error={touched.firstName && Boolean(errors.firstName)}
                helperText={<ErrorMessage name="firstName" />}
              />
              <Field
                name="lastName"
                as={TextField}
                label="Last Name"
                fullWidth
                margin="normal"
                error={touched.lastName && Boolean(errors.lastName)}
                helperText={<ErrorMessage name="lastName" />}
              />
              <Field
                name="about"
                as={TextField}
                label="About Me"
                fullWidth
                margin="normal"
                multiline
                rows={3}
                error={touched.about && Boolean(errors.about)}
                helperText={<ErrorMessage name="about" />}
              />

     <FormControl fullWidth margin="normal" error={touched.city && Boolean(errors.city)}>
       <InputLabel id="city-label">City</InputLabel>
       <Field name="city">
         {({ field, form }) => (
           <Select
             {...field}
             labelId="city-label"
             label="City"
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
                Biography
              </Typography>

     <FormControl fullWidth margin="normal" error={touched.interests && Boolean(errors.interests)}>
       <InputLabel id="interests-label">Interests</InputLabel>
       <Field name="interests">
    {({ field, form }) => (
      <Select
        {...field}
        multiple
        labelId="interests-label"
        label="Interests"
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
         label="Priority Interests"
       />
       <ErrorMessage name="interests" component="div" style={{ color: 'red' }} />
     </FormControl>

     <FormControl fullWidth margin="normal" error={touched.hobbies && Boolean(errors.hobbies)}>
  <InputLabel id="hobbies-label">Hobbies</InputLabel>
  <Field name="hobbies">
    {({ field, form }) => (
      <Select
        {...field}
        multiple
        labelId="hobbies-label"
        label="Hobbies"
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
    label="Priority Hobby"
  />
  <ErrorMessage name="hobbies" component="div" style={{ color: 'red' }} />
</FormControl>

     <FormControl fullWidth margin="normal" error={touched.music && Boolean(errors.music)}>
  <InputLabel id="music-label">Music</InputLabel>
  <Field name="music">
    {({ field, form }) => (
      <Select
        {...field}
        multiple
        labelId="music-label"
        label="Music"
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
    label="Priority Music"
  />
  <ErrorMessage name="music" component="div" style={{ color: 'red' }} />
</FormControl>

     <FormControl fullWidth margin="normal" error={touched.food && Boolean(errors.food)}>
  <InputLabel id="food-label">Food</InputLabel>
  <Field name="food">
    {({ field, form }) => (
      <Select
        {...field}
        multiple
        labelId="food-label"
        label="Food"
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
    label="Priority Cuisine"
  />
  <ErrorMessage name="food" component="div" style={{ color: 'red' }} />
</FormControl>

     <FormControl fullWidth margin="normal" error={touched.travel && Boolean(errors.travel)}>
  <InputLabel id="travel-label">Travel</InputLabel>
  <Field name="travel">
    {({ field, form }) => (
      <Select
        {...field}
        multiple
        labelId="travel-label"
        label="Travel"
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
    label="Priority Travel"
  />
  <ErrorMessage name="travel" component="div" style={{ color: 'red' }} />
</FormControl>
              
              <Field
                name="lookingFor"
                as={TextField}
                label="Who are you looking for"
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
                {isSubmitting ? 'Saving...' : 'Save Changes'}
              </Button>
            </Form>
          )}
        </Formik>
      </Box>
    </Container>
  );
};

export default EditProfile;
