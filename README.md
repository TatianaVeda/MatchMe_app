# Match Me

A modern web application for connecting people with similar movie interests. Built with React, Vite, and Go.

## Features

- User authentication (register/login)
- Profile management
- Movie preference matching
- Real-time chat
- Connection management
- Recommendation system

## Tech Stack

### Frontend
- React
- Vite
- TailwindCSS
- Zustand (State Management)
- React Router
- Axios
- WebSocket for real-time chat

### Backend
- Go
- Gin (Web Framework)
- GORM (ORM)
- PostgreSQL
- JWT Authentication
- WebSocket

## Prerequisites

- Node.js (v18 or higher)
- Go (v1.23.3 or higher)
- PostgreSQL

## Getting Started

### Frontend Setup

1. Install dependencies:
```bash
npm install
```

2. Start the development server:
```bash
npm run dev
```

The frontend will be available at `http://localhost:5173`

### Backend Setup

1. Set up PostgreSQL database and update the connection string in `backend/database/database.go`

2. Install Go dependencies:
```bash
cd backend
go mod download
```

3. Start the backend server:
```bash
go run main.go
```

The backend API will be available at `http://localhost:8080`

## Project Structure

```
.
├── backend/
│   ├── database/       # Database connection and configuration
│   ├── handlers/       # HTTP request handlers
│   ├── middlewares/    # Middleware functions
│   ├── models/         # Database models
│   ├── routes/         # Route definitions
│   ├── utils/          # Utility functions
│   └── websocket/      # WebSocket implementation
├── src/
│   ├── components/     # React components
│   ├── pages/          # Page components
│   ├── stores/         # Zustand stores
│   └── main.jsx        # Application entry point
```

## API Endpoints

### Authentication
- `POST /api/register` - Register new user
- `POST /api/login` - User login

### Profile
- `GET /api/me` - Get current user profile
- `PUT /api/me` - Update profile
- `PUT /api/me/avatar` - Update avatar

### Recommendations
- `GET /api/recommendations` - Get user recommendations
- `POST /api/recommendations/dismiss` - Dismiss a recommendation

### Connections
- `GET /api/connections` - Get user connections
- `POST /api/connections/request` - Send connection request
- `POST /api/connections/accept` - Accept connection request
- `POST /api/connections/reject` - Reject connection request

### Chat
- `GET /api/chat/messages` - Get chat messages
- `POST /api/chat/send` - Send message
- `GET /api/chat/unread` - Get unread message count

## Contributing

1. Fork the repository
2. Create a new branch
3. Make your changes
4. Submit a pull request

## License

This project is licensed under the MIT License.