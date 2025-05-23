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

const Settings = () => {
  const [loading, setLoading] = useState(true);
  const [savingPrefs, setSavingPrefs] = useState(false);
  const [savingEmail, setSavingEmail] = useState(false);
  const [savingPassword, setSavingPassword] = useState(false);
  const [preferences, setPreferences] = useState({
    maxRadius: ''    
  });
  const [saving, setSaving] = useState(false);

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
        toast.error('Ошибка загрузки настроек');
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
    e.preventDefault();
    setSavingPrefs(true);
    try {
      await axios.put('/me/preferences', {
        maxRadius: Number(preferences.maxRadius)
      });
      toast.success('Настройки рекомендаций сохранены');
    } catch (err) {
      toast.error(err.response?.data?.message || 'Ошибка сохранения настроек');
    } finally {
      setSavingPrefs(false);
    }
  };

  const handleEmailChange = e => {
    setEmail({ ...email, newEmail: e.target.value });
  };
  const submitEmail = async e => {
    e.preventDefault();
    if (!email.newEmail) {
      toast.error('Введите новый e-mail');
      return;
    }
    setSavingEmail(true);
    try {
      const { data } = await axios.put('/me/email', {
        email: email.newEmail
      });
      setEmail({ currentEmail: data.email, newEmail: '' });
      toast.success('E-mail успешно изменён');
    } catch (err) {
      toast.error(err.response?.data || 'Ошибка изменения e-mail');
    } finally {
      setSavingEmail(false);
    }
  };

  const handlePasswordChange = e => {
    setPasswords({ ...passwords, [e.target.name]: e.target.value });
  };
  const submitPassword = async e => {
    e.preventDefault();
    const { current, next, confirm } = passwords;
    if (!current || !next || next !== confirm) {
      toast.error('Проверьте поля пароля');
      return;
    }
    setSavingPassword(true);
    try {
      await axios.put('/me/password', {
        current,
        new: next
      });
      setPasswords({ current: '', next: '', confirm: '' });
      toast.success('Пароль успешно изменён');
    } catch (err) {
      toast.error(err.response?.data || 'Ошибка изменения пароля');
    } finally {
      setSavingPassword(false);
    }
  };

  return (
    <Container maxWidth="sm" sx={{ mt: 4 }}>
      <Typography variant="h4" gutterBottom>Настройки</Typography>

      <Box component="form" onSubmit={submitEmail} sx={{ p: 2, mb: 3, border: '1px solid #ccc', borderRadius: 2 }}>
        <Typography variant="h6" gutterBottom>Сменить e-mail</Typography>
        <TextField
          label="Текущий e-mail"
          value={email.currentEmail}
          fullWidth
          margin="normal"
          InputProps={{ readOnly: true }}
        />
        <TextField
          label="Новый e-mail"
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
          {savingEmail ? 'Сохраняем...' : 'Сохранить e-mail'}
        </Button>
      </Box>

      <Divider />

      <Box component="form" onSubmit={submitPassword} sx={{ p: 2, my: 3, border: '1px solid #ccc', borderRadius: 2 }}>
        <Typography variant="h6" gutterBottom>Сменить пароль</Typography>
        <TextField
          label="Текущий пароль"
          name="current"
          type="password"
          value={passwords.current}
          onChange={handlePasswordChange}
          fullWidth
          margin="normal"
          required
        />
        <TextField
          label="Новый пароль"
          name="next"
          type="password"
          value={passwords.next}
          onChange={handlePasswordChange}
          fullWidth
          margin="normal"
          required
        />
        <TextField
          label="Подтвердите новый пароль"
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
          {savingPassword ? 'Сохраняем...' : 'Сохранить пароль'}
        </Button>
      </Box>

      <Divider />

      <Box component="form" onSubmit={submitPreferences} sx={{ p: 2, mt: 3, border: '1px solid #ccc', borderRadius: 2 }}>
        <Typography variant="h6" gutterBottom>Параметры рекомендаций</Typography>
        <TextField
          label="Максимальный радиус (км)"
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
          {savingPrefs ? 'Сохраняем...' : 'Сохранить параметры'}
        </Button>
      </Box>
    </Container>
  );
};

export default Settings;