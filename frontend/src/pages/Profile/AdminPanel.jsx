import React, { useState } from 'react'
import { Container, Typography, Button, Box, TextField } from '@mui/material'
import axios from '../api/index'
import { toast } from 'react-toastify'

const AdminPanel = () => {
  const [num, setNum] = useState(100)
  const handleReset = async () => {
    try {
      await axios.post(`/fixtures/reset?num=${num}`)
      toast.success('БД сброшена и фиктивные пользователи созданы')
    } catch (e) {
      toast.error('Ошибка сброса БД')
    }
  }
  const handleGenerate = async () => {
    try {
      await axios.post(`/fixtures/generate?num=${num}`)
      toast.success('Фиктивные пользователи созданы')
    } catch {
      toast.error('Ошибка генерации')
    }
  }
  return (
    <Container sx={{ mt: 4 }}>
      <Typography variant="h4">Admin Panel</Typography>
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
