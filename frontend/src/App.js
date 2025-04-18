// import logo from './logo.svg';
// import './App.css';

// function App() {
//   return (
//     <div className="App">
//       <header className="App-header">
//         <img src={logo} className="App-logo" alt="logo" />
//         <p>
//           Edit <code>src/App.js</code> and save to reload.
//         </p>
//         <a
//           className="App-link"
//           href="https://reactjs.org"
//           target="_blank"
//           rel="noopener noreferrer"
//         >
//           Learn React
//         </a>
//       </header>
//     </div>
//   );
// }

// export default App;


// // // m/frontend/src/App.jsx
// // import React from 'react';
// // import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
// // import { AuthProvider } from './contexts/AuthContext';
// // import Login from './pages/Auth/Login';
// // import Signup from './pages/Auth/Signup';
// // import Dashboard from './pages/Dashboard'; // Страница, доступная только авторизованным пользователям
// // import ProtectedRoute from './components/ProtectedRoute';
// // import Home from './pages/Home';




// // function App() {
// //   return (
// //     <Router>
// //       <AuthProvider>
// //         <Routes>
// //         <Route path="/" element={<Home />} />
// //           <Route path="/login" element={<Login />} />
// //           <Route path="/signup" element={<Signup />} />
// //           {/* Пример защищённого маршрута */}
// //           <Route
// //             path="/dashboard"
// //             element={
// //               <ProtectedRoute>
// //                 <Dashboard />
// //               </ProtectedRoute>
// //             }
// //           />
// //           {/* Остальные маршруты */}
// //         </Routes>
// //       </AuthProvider>
// //     </Router>
// //   );
// // }

// // export default App;


// // m/frontend/src/App.jsx
// import React from 'react';
// import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
// import { AuthProvider } from './contexts/AuthContext';
// import PrivateRoute from './components/PrivateRoute';
// import Login from './pages/Auth/Login';
// import Signup from './pages/Auth/Signup';
// import MyProfile from './pages/Profile/MyProfile';
// import EditProfile from './pages/Profile/EditProfile';
// import UserProfile from './pages/Profile/UserProfile';
// import Recommendations from './pages/Recommendations';
// import Connections from './pages/Connections';
// import Chats from './pages/Chats';
// import ChatWindow from './pages/ChatWindow';
// import Header from './components/Header';

// function App() {
//   return (
//     <AuthProvider>
//       <Router>
//         <Header />
//         <Routes>
//           <Route path="/login" element={<Login />} />
//           <Route path="/signup" element={<Signup />} />
//           <Route
//             path="/me"
//             element={<PrivateRoute><MyProfile /></PrivateRoute>}
//           />
//           <Route
//             path="/edit-profile"
//             element={<PrivateRoute><EditProfile /></PrivateRoute>}
//           />
//           <Route
//             path="/users/:id"
//             element={<PrivateRoute><UserProfile /></PrivateRoute>}
//           />
//           <Route
//             path="/recommendations"
//             element={<PrivateRoute><Recommendations /></PrivateRoute>}
//           />
//           <Route
//             path="/connections"
//             element={<PrivateRoute><Connections /></PrivateRoute>}
//           />
//           <Route
//             path="/chats"
//             element={<PrivateRoute><Chats /></PrivateRoute>}
//           />
//           <Route
//             path="/chat/:chatId"
//             element={<PrivateRoute><ChatWindow /></PrivateRoute>}
//           />
//           {/* Можно добавить дополнительные маршруты */}
//         </Routes>
//       </Router>
//     </AuthProvider>
//   );
// }

// export default App;

//В ИТОГЕ СРАВНИТЬ ОБА ФАЙЛА APP.JS И APP.JSX!!!! APP.JSX должен быть в проекте а не APP.JS
import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { AuthProvider } from './contexts/AuthContext';
import { ChatProvider } from './contexts/ChatContext';
import PrivateRoute from './components/PrivateRoute';
import Login from './pages/Auth/Login';
import Signup from './pages/Auth/Signup';
import MyProfile from './pages/Profile/MyProfile';
import EditProfile from './pages/Profile/EditProfile';
import Recommendations from './pages/Recommendations';
import Connections from './pages/Connections';
import Chats from './pages/Chats';
import ChatWindow from './pages/ChatWindow';
import Settings from './pages/Settings';
import Header from './components/Header';

function App() {
  return (
    <AuthProvider>
      <ChatProvider>
        <Router>
          <Header />
          <Routes>
            {/* Публичные маршруты */}
            <Route path="/login" element={<Login />} />
            <Route path="/signup" element={<Signup />} />

            {/* Защищённые маршруты */}
            <Route path="/me" element={<PrivateRoute><MyProfile /></PrivateRoute>} />
            <Route path="/edit-profile" element={<PrivateRoute><EditProfile /></PrivateRoute>} />
            <Route path="/recommendations" element={<PrivateRoute><Recommendations /></PrivateRoute>} />
            <Route path="/connections" element={<PrivateRoute><Connections /></PrivateRoute>} />
            <Route path="/chats" element={<PrivateRoute><Chats /></PrivateRoute>} />
            <Route path="/chat/:chatId" element={<PrivateRoute><ChatWindow /></PrivateRoute>} />
            <Route path="/settings" element={<PrivateRoute><Settings /></PrivateRoute>} />
          </Routes>
        </Router>
      </ChatProvider>
    </AuthProvider>
  );
}

export default App;
