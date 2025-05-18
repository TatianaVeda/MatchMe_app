// // /m/frontend/src/components/UserCard.jsx
// import React from 'react';
// import PropTypes from 'prop-types';
// import { Card, CardActionArea, CardContent, Avatar, Typography, Badge, Box, IconButton, Tooltip } from '@mui/material';
// import ChatIcon from '@mui/icons-material/Chat';

// // const UserCard = ({ user, onClick, onChatClick, showChat }) => {
// //   const { firstName, lastName, photoUrl, online, connected } = user;

// const UserCard = ({
//        user,
//        onClick = () => {},
//        onChatClick = () => {},
//        showChat = false
//      }) => {
//         const { firstName, lastName, photoUrl, online, connected } = user;

//   return (
//     <Card
//       onClick={onClick}
//       sx={{
//         width: '100%',
//         maxWidth: 240,
//         cursor: 'pointer',
//         position: 'relative',
//         '&:hover': { boxShadow: 6 },
//       }}
//     >
//       <CardActionArea>
//         <Box sx={{ display: 'flex', justifyContent: 'center', mt: 2 }}>
//           <Badge
//             color="success"
//             variant="dot"
//             invisible={!online}
//             overlap="circular"
//             anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
//           >
//             <Avatar
//               src={photoUrl || undefined}
//               alt={`${firstName} ${lastName}`}
//               sx={{ width: 80, height: 80 }}
//             >
//               {!photoUrl && '👤'}
//             </Avatar>
//           </Badge>
//         </Box>
//         <CardContent sx={{ textAlign: 'center', pt: 1 }}>
//           <Typography variant="subtitle1" noWrap>
//             {firstName} {lastName}
//           </Typography>
//         </CardContent>
//       </CardActionArea>
 
// {showChat && connected && (
//   <Tooltip title="Перейти в чат">
//     <IconButton
//       size="small"
//       onClick={(e) => {
//         e.stopPropagation(); // чтобы не сработал onClick карточки
//         onChatClick?.(user);
//       }}
//       sx={{
//         position: 'absolute',
//         top: 8,
//         right: 8,
//         backgroundColor: 'white',
//         '&:hover': { backgroundColor: 'lightgray' },
//       }}
//     >
//       <ChatIcon fontSize="small" />
//     </IconButton>
//   </Tooltip>
// )}
// </Card>
// );
// };

// UserCard.propTypes = {
//   user: PropTypes.shape({
//     firstName: PropTypes.string,
//     lastName:  PropTypes.string,
//     photoUrl: PropTypes.string,
//     online:    PropTypes.bool,
//     connected: PropTypes.bool, 
//   }).isRequired,
//   onClick: PropTypes.func,
//   onChatClick: PropTypes.func,
//   showChat: PropTypes.bool,
// };

// // UserCard.defaultProps = {
// //   onClick: () => {},
// //   onChatClick: () => {},
// //   showChat: false,
// // };

// export default UserCard;

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
        <CardContent sx={{ textAlign: 'center', pt: 1 }}>
          <Typography variant="subtitle1" noWrap>
            {firstName} {lastName}
          </Typography>
        </CardContent>
      </CardActions>

      {showChat && connected && (
        <Tooltip title="Перейти в чат">
          <IconButton
            size="small"
            onClick={(e) => {
              e.stopPropagation();  // чтобы не сработал onClick самой карточки
              //onChatClick();
              onChatClick?.(user);
            }}
            sx={{
              position: 'absolute',
              top: 8,
              right: 8,
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
    firstName: PropTypes.string,
    lastName: PropTypes.string,
    photoUrl: PropTypes.string,
    online: PropTypes.bool,
    connected: PropTypes.bool,
  }).isRequired,
  onClick:     PropTypes.func,
  onChatClick: PropTypes.func,
  onDisconnect: PropTypes.func,
  showChat: PropTypes.bool,
  showDisconnect: PropTypes.bool,
};

export default UserCard;
