import React, { useState } from 'react'
import { Container, Typography, Button, Box, TextField } from '@mui/material'
import { toast } from 'react-toastify'
import axios from '../../api/index'
import { useAuthState } from '../../contexts/AuthContext';
import { ADMIN_ID } from '../../config'

const AdminPanel = () => {
  const { user } = useAuthState();
  const [num, setNum] = useState(100)

 // Показываем панель только если это админ
 if (user?.id !== ADMIN_ID) {
  return null;
}

  const handleReset = async () => {
    try {
      await axios.post(`/admin/reset-fixtures?num=${num}`)
      toast.success('База Данных успешно сброшена')
    } catch (e) {
      toast.error('Ошибка сброса БД')
    }
  }
  const handleGenerate = async () => {
    try {
        await axios.post(`/admin/generate-fixtures?num=${num}`)
      toast.success('Фиктивные пользователи созданы')
    } catch {
      toast.error('Ошибка генерации')
    }
  }
  return (
    <Container sx={{ mt: 4 }}>
      <Typography variant="h4" gutterBottom>Admin Panel</Typography>
      <Box sx={{ my: 2 }}>
        <TextField
          label="Количество фиктивных пользователей"
          type="number"
          value={num}
          onChange={e => setNum(+e.target.value)}
        />
      </Box>
      <Box sx={{ display: 'flex', gap: 2 }}>
        <Button variant="contained" color="error" onClick={handleReset}>
          Сброс базы данных
        </Button>
        <Button variant="contained" onClick={handleGenerate}>
          Генерация фиктивных пользователей
        </Button>
      </Box>
    </Container>
  )
}

export default AdminPanel
