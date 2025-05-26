// /m/frontend/src/components/UserCard.jsx
import React from 'react';
import { useChatState } from '../contexts/ChatContext';
import PropTypes from 'prop-types';
import {
  Card,
  CardActionArea,
  CardContent,
  Avatar,
  Typography,
  Badge,
  Box,
  IconButton,
  Tooltip
} from '@mui/material';
import ChatIcon from '@mui/icons-material/Chat';

const UserCard = ({ user, onClick, onChatClick, showChat }) => {
  const { presence } = useChatState();
  const { firstName, lastName, photoUrl, online: propOnline, connected } = user;

  const online = typeof propOnline === 'boolean'
    ? propOnline
    : Boolean(presence?.[user.id]);

  return (
    <Card
      onClick={onClick}
      sx={{
        width: '100%',
        maxWidth: 240,
        cursor: 'pointer',
        position: 'relative',
        '&:hover': { boxShadow: 6 },
      }}
    >
      <CardActionArea>
        <Box sx={{ display: 'flex', justifyContent: 'center', mt: 2 }}>
          <Badge
  color={ online ? "success" : "error" }  
  variant="dot"
  overlap="circular"
  anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
>
            <Avatar
              src={photoUrl || undefined}
              alt={`${firstName} ${lastName}`}
              sx={{ width: 80, height: 80 }}
            >
              {!photoUrl && '👤'}
            </Avatar>
          </Badge>
        </Box>
        <CardContent sx={{ textAlign: 'center', pt: 1 }}>
          <Typography variant="subtitle1" noWrap>
            {firstName} {lastName}
          </Typography>
        </CardContent>
      </CardActionArea>

      {showChat && connected && (
        <Tooltip title="Go to chat">
          <IconButton
            size="small"
            onClick={(e) => {
              e.stopPropagation();  
              onChatClick?.(user);
            }}
            sx={{
                      position: 'absolute',
                      top: 8,
                      left: 8,               
                      backgroundColor: 'white',
                      '&:hover': { backgroundColor: 'lightgray' },
                    }}
          >
            <ChatIcon fontSize="small" />
          </IconButton>
        </Tooltip>
      )}
    </Card>
  );
};

UserCard.propTypes = {
  user: PropTypes.shape({
    id:        PropTypes.string.isRequired,
    firstName: PropTypes.string,
    lastName:  PropTypes.string,
    photoUrl:  PropTypes.string,
    online:    PropTypes.bool,
    connected: PropTypes.bool,
  }).isRequired,
  onClick:     PropTypes.func,
  onChatClick: PropTypes.func,
  showChat:    PropTypes.bool,
};

export default UserCard;
