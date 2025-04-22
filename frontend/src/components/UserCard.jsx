// /m/frontend/src/components/UserCard.jsx
import React from 'react';
import PropTypes from 'prop-types';
import { Card, CardActionArea, CardContent, Avatar, Typography, Badge, Box } from '@mui/material';

const UserCard = ({ user, onClick }) => {
  const { firstName, lastName, photoUrl, online } = user;

  return (
    <Card
      onClick={onClick}
      sx={{
        width: '100%',
        maxWidth: 240,
        cursor: 'pointer',
        '&:hover': { boxShadow: 6 },
      }}
    >
      <CardActionArea>
        <Box sx={{ display: 'flex', justifyContent: 'center', mt: 2 }}>
          <Badge
            color="success"
            variant="dot"
            invisible={!online}
            overlap="circular"
            anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
          >
            <Avatar
              src={photoUrl}
              alt={`${firstName} ${lastName}`}
              sx={{ width: 80, height: 80 }}
            />
          </Badge>
        </Box>
        <CardContent sx={{ textAlign: 'center', pt: 1 }}>
          <Typography variant="subtitle1" noWrap>
            {firstName} {lastName}
          </Typography>
        </CardContent>
      </CardActionArea>
    </Card>
  );
};

UserCard.propTypes = {
  user: PropTypes.shape({
    firstName: PropTypes.string,
    lastName:  PropTypes.string,
    photoUrl: PropTypes.string,
    online:    PropTypes.bool,
  }).isRequired,
  onClick: PropTypes.func,
};

UserCard.defaultProps = {
  onClick: () => {},
};

export default UserCard;
