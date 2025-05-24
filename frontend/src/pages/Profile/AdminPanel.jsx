import React, { useState } from 'react'
import { Container, Typography, Button, Box, TextField } from '@mui/material'
import { toast } from 'react-toastify'
import axios from '../../api/index'
import { useAuthState } from '../../contexts/AuthContext';
import { ADMIN_ID } from '../../config'

const AdminPanel = () => {
  const { user } = useAuthState();
  const [num, setNum] = useState(100)

 if (user?.id !== ADMIN_ID) {
  return null;
}

  const handleReset = async () => {
    try {
      await axios.post(`/admin/reset-fixtures?num=${num}`)
      toast.success('Database has been successfully reset')
    } catch (e) {
      toast.error('Database reset error')
    }
  }
  const handleGenerate = async () => {
    try {
        await axios.post(`/admin/generate-fixtures?num=${num}`)
      toast.success('Test users have been created')
    } catch {
      toast.error('Generation error')
    }
  }
  return (
    <Container sx={{ mt: 4 }}>
      <Typography variant="h4" gutterBottom>Admin Panel</Typography>
      <Box sx={{ my: 2 }}>
        <TextField
          label="Number of test users"
          type="number"
          value={num}
          onChange={e => setNum(+e.target.value)}
        />
      </Box>
      <Box sx={{ display: 'flex', gap: 2 }}>
        <Button variant="contained" color="error" onClick={handleReset}>
          Reset Database
        </Button>
        <Button variant="contained" onClick={handleGenerate}>
          Generate Test Users
        </Button>
      </Box>
    </Container>
  )
}

export default AdminPanel
