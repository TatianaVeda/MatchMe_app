import React, { useState, useEffect } from 'react';
import { Container, Box, Typography, Button, Avatar, CircularProgress } from '@mui/material';
import { useNavigate } from 'react-router-dom';
import { toast } from 'react-toastify';
import { useAuthState } from '../../contexts/AuthContext';

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
      const status = error.response?.status;
  
      if (status === 404) {
        setProfile(null);
      } else {
        toast.error(error.response?.data?.message || 'Error loading profile');
      }
    } finally {
      setLoading(false);
    }
  };
  

  const fetchBio = async () => {
    try {
      const data = await getMyBio();
      setBio(data);
    } catch (error) {
      toast.error(error.response?.data?.message || 'Error loading biography');
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
        <Typography variant="h6">Profile not found. Please fill in your profile.</Typography>
        <Button variant="contained" color="primary" onClick={handleEdit} sx={{ mt: 2 }}>
          Edit Profile
        </Button>
      </Container>
    );
  }

  return (
    <Container maxWidth="md" sx={{ mt: 4 }}>
      <Box sx={{ display: 'flex', alignItems: 'center', mb: 3 }}>
        <Avatar
          alt={`${profile.firstName} ${profile.lastName}`}
          src={profile.photoUrl || undefined}
          sx={{ width: 80, height: 80, mr: 2 }}
        >
          {!profile.photoUrl && '👤'}
        </Avatar>
        <Typography variant="h4">
          {profile.firstName} {profile.lastName}
        </Typography>
      </Box>

      <Box sx={{ mb: 3 }}>
        <Typography variant="body1" color="textSecondary">
          {profile.about || 'User information is not filled in.'}
        </Typography>
      </Box>

      <Box sx={{ mb: 3 }}>
        <Typography variant="body1">
          City: {profile.city || 'Not specified'}
        </Typography>
      </Box>

      <Box sx={{ mb: 3 }}>
        <Typography variant="h6" gutterBottom>
          Biography
        </Typography>
        {bio ? (
          <>
            <Typography variant="body1">Interests: {bio.interests || 'Not specified'}</Typography>
            <Typography variant="body1">Hobbies: {bio.hobbies || 'Not specified'}</Typography>
            <Typography variant="body1">Music: {bio.music || 'Not specified'}</Typography>
            <Typography variant="body1">Cuisine: {bio.food || 'Not specified'}</Typography>
            <Typography variant="body1">Travel: {bio.travel || 'Not specified'}</Typography>
            <Typography variant="body1">Looking for: {bio.lookingFor || 'Not specified'}</Typography>
          </>
        ) : (
          <Typography variant="body1">Biography is not filled yet.</Typography>
        )}
      </Box>

      <Button variant="contained" color="primary" onClick={handleEdit}>
        Edit Profile
      </Button>
    </Container>
  );
};

export default MyProfile;
