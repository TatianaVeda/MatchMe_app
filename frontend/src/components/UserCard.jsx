// /m/frontend/src/components/UserCard.jsx
import React from 'react';
import PropTypes from 'prop-types';
import { Card, CardActionArea, CardContent, Avatar, Typography, Badge, Box, IconButton, Tooltip, CardMedia, CardActions, Button } from '@mui/material';
import ChatIcon from '@mui/icons-material/Chat';
import PersonRemoveIcon from '@mui/icons-material/PersonRemove';
import { Link } from 'react-router-dom';

// const UserCard = ({ user, onClick, onChatClick, showChat }) => {
//   const { firstName, lastName, photoUrl, online, connected } = user;

const UserCard = ({
  user,
  onClick = () => {},
  onChatClick = () => {},
  onDisconnect = () => {},
  showChat = false,
  showDisconnect = false
}) => {
  const { firstName, lastName, photoUrl, online, connected } = user;

  return (
    <Card sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      <CardMedia
        component="img"
        height="140"
        image={user.photoUrl || '/default-avatar.png'}
        alt={`${user.firstName} ${user.lastName}`}
      />
      <CardContent sx={{ flexGrow: 1 }}>
        <Typography gutterBottom variant="h5" component="div">
          {user.firstName} {user.lastName}
        </Typography>
        <Typography variant="body2" color="text.secondary">
          {user.bio || 'Нет описания'}
        </Typography>
      </CardContent>
      <CardActions>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', width: '100%' }}>
          <Button size="small" onClick={onClick}>
            Профиль
          </Button>
          {showChat && (
            <Button size="small" onClick={onChatClick}>
              Чат
            </Button>
          )}
          {showDisconnect && (
            <Button 
              size="small" 
              color="error" 
              onClick={onDisconnect}
            >
              Отключить
            </Button>
          )}
        </Box>
      </CardActions>
    </Card>
  );
};

UserCard.propTypes = {
  user: PropTypes.shape({
    firstName: PropTypes.string,
    lastName: PropTypes.string,
    photoUrl: PropTypes.string,
    online: PropTypes.bool,
    connected: PropTypes.bool,
  }).isRequired,
  onClick: PropTypes.func,
  onChatClick: PropTypes.func,
  onDisconnect: PropTypes.func,
  showChat: PropTypes.bool,
  showDisconnect: PropTypes.bool,
};

// UserCard.defaultProps = {
//   onClick: () => {},
//   onChatClick: () => {},
//   showChat: false,
// };

export default UserCard;
