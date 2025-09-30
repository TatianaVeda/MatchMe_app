// m/frontend/src/pages/Settings.jsx
import React, { useState, useEffect } from 'react';
import { 
  Container, 
  Box, 
  Typography, 
  TextField, 
  Button, 
  CircularProgress,
  Divider 
} from '@mui/material';
import axios from '../api/index';
import { toast } from 'react-toastify';

/**
 * Settings.jsx
 *
 * User settings page. Allows changing email, password, and recommendation preferences.
 * Integrates with backend API, handles batch loading and error reporting.
 */

const Settings = () => {
  const [loading, setLoading] = useState(true);
  const [savingPrefs, setSavingPrefs] = useState(false);
  const [savingEmail, setSavingEmail] = useState(false);
  const [savingPassword, setSavingPassword] = useState(false);
  const [preferences, setPreferences] = useState({
    maxRadius: ''    
  });
 // const [saving, setSaving] = useState(false);

   const [email, setEmail] = useState({
    currentEmail: '',
    newEmail: ''
  });

  const [passwords, setPasswords] = useState({
    current: '',
    next: '',
    confirm: ''
  });

  useEffect(() => {
    const fetchAll = async () => {
      try {
        const prefRes = await axios.get('/me/preferences');
        setPreferences({ maxRadius: prefRes.data.maxRadius || '' });

        const meRes = await axios.get('/me');
        setEmail(email => ({ ...email, currentEmail: meRes.data.email || '' }));
      } catch (err) {
        toast.error('Error loading settings');
      } finally {
        setLoading(false);
      }
    };
    fetchAll();
  }, []);

  if (loading) {
    return (
      <Container sx={{ textAlign: 'center', mt: 4 }}>
        <CircularProgress />
      </Container>
    );
  }

  const handlePrefChange = e => {
    setPreferences({ maxRadius: e.target.value });
  };
  const submitPreferences = async e => {
    /**
     * Saves recommendation preferences (maxRadius) to backend.
     * Handles API errors and loading state.
     */
    e.preventDefault();
    setSavingPrefs(true);
    try {
      await axios.put('/me/preferences', {
        maxRadius: Number(preferences.maxRadius)
      });
      toast.success('Recommendation settings saved');
    } catch (err) {
      toast.error(err.response?.data?.message || 'Error saving settings');
    } finally {
      setSavingPrefs(false);
    }
  };

  const handleEmailChange = e => {
    setEmail({ ...email, newEmail: e.target.value });
  };
  const submitEmail = async e => {
    /**
     * Updates user email in backend.
     * Handles validation, API errors, and loading state.
     */
    e.preventDefault();
    if (!email.newEmail) {
      toast.error('Please enter new email');
      return;
    }
    setSavingEmail(true);
    try {
      const { data } = await axios.put('/me/email', {
        email: email.newEmail
      });
      setEmail({ currentEmail: data.email, newEmail: '' });
      toast.success('Email successfully updated');
    } catch (err) {
      toast.error(err.response?.data || 'Error updating email');
    } finally {
      setSavingEmail(false);
    }
  };

  const handlePasswordChange = e => {
    setPasswords({ ...passwords, [e.target.name]: e.target.value });
  };
  const submitPassword = async e => {
    /**
     * Updates user password in backend.
     * Handles validation, API errors, and loading state.
     */
    e.preventDefault();
    const { current, next, confirm } = passwords;
    if (!current || !next || next !== confirm) {
      toast.error('Please check password fields');
      return;
    }
    setSavingPassword(true);
    try {
      await axios.put('/me/password', {
        current,
        new: next
      });
      setPasswords({ current: '', next: '', confirm: '' });
      toast.success('Password successfully updated');
    } catch (err) {
      toast.error(err.response?.data || 'Error updating password');
    } finally {
      setSavingPassword(false);
    }
  };

  return (
    <Container maxWidth="sm" sx={{ mt: 4 }}>
      <Typography variant="h4" gutterBottom>Settings</Typography>

      <Box component="form" onSubmit={submitEmail} sx={{ p: 2, mb: 3, border: '1px solid #ccc', borderRadius: 2 }}>
        <Typography variant="h6" gutterBottom>Change Email</Typography>
        <TextField
          label="Current Email"
          value={email.currentEmail}
          fullWidth
          margin="normal"
          InputProps={{ readOnly: true }}
        />
        <TextField
          label="New Email"
          name="newEmail"
          value={email.newEmail}
          onChange={handleEmailChange}
          fullWidth
          margin="normal"
          required
        />
        <Button
          variant="contained"
          type="submit"
          disabled={savingEmail}
          sx={{ mt: 1 }}
        >
          {savingEmail ? 'Saving...' : 'Save Email'}
        </Button>
      </Box>

      <Divider />

      <Box component="form" onSubmit={submitPassword} sx={{ p: 2, my: 3, border: '1px solid #ccc', borderRadius: 2 }}>
        <Typography variant="h6" gutterBottom>Change Password</Typography>
        <TextField
          label="Current Password"
          name="current"
          type="password"
          value={passwords.current}
          onChange={handlePasswordChange}
          fullWidth
          margin="normal"
          required
        />
        <TextField
          label="New Password"
          name="next"
          type="password"
          value={passwords.next}
          onChange={handlePasswordChange}
          fullWidth
          margin="normal"
          required
        />
        <TextField
          label="Confirm New Password"
          name="confirm"
          type="password"
          value={passwords.confirm}
          onChange={handlePasswordChange}
          fullWidth
          margin="normal"
          required
        />
        <Button
          variant="contained"
          type="submit"
          disabled={savingPassword}
          sx={{ mt: 1 }}
        >
          {savingPassword ? 'Saving...' : 'Save Password'}
        </Button>
      </Box>

      <Divider />

      <Box component="form" onSubmit={submitPreferences} sx={{ p: 2, mt: 3, border: '1px solid #ccc', borderRadius: 2 }}>
        <Typography variant="h6" gutterBottom>Recommendation Settings</Typography>
        <TextField
          label="Maximum Radius (km)"
          name="maxRadius"
          type="number"
          value={preferences.maxRadius}
          onChange={handlePrefChange}
          fullWidth
          margin="normal"
          required
        />
        <Button
          variant="contained"
          type="submit"
          disabled={savingPrefs}
          sx={{ mt: 1 }}
        >
          {savingPrefs ? 'Saving...' : 'Save Settings'}
        </Button>
      </Box>
    </Container>
  );
};

export default Settings;