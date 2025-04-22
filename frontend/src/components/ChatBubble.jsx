// /m/frontend/src/components/ChatBubble.jsx
import React from 'react';
import PropTypes from 'prop-types';
import { Box, Typography } from '@mui/material';

const ChatBubble = ({ message, isOwn }) => {
  const time = new Date(message.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });

  return (
    <Box
      sx={{
        display: 'flex',
        justifyContent: isOwn ? 'flex-end' : 'flex-start',
        mb: 1,
        px: 1,
      }}
    >
      <Box
        sx={{
          maxWidth: '75%',
          p: 1.5,
          bgcolor: isOwn ? 'primary.main' : 'grey.200',
          color: isOwn ? 'primary.contrastText' : 'text.primary',
          borderRadius: 2,
          borderTopRightRadius: isOwn ? 0 : 8,
          borderTopLeftRadius: isOwn ? 8 : 0,
        }}
      >
        <Typography variant="body2" sx={{ wordBreak: 'break-word' }}>
          {message.content}
        </Typography>
        <Typography
          variant="caption"
          component="div"
          sx={{ textAlign: 'right', mt: 0.5 }}
        >
          {time} {isOwn && (message.read ? '✓✓' : '✓')}
        </Typography>
      </Box>
    </Box>
  );
};

ChatBubble.propTypes = {
  message: PropTypes.shape({
    content:   PropTypes.string,
    timestamp: PropTypes.string,
    read:      PropTypes.bool,
  }).isRequired,
  isOwn: PropTypes.bool,
};

ChatBubble.defaultProps = {
  isOwn: false,
};

export default ChatBubble;
