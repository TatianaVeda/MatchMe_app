// // frontend/src/pages/Dashboard.jsx
// import React, { useContext } from 'react';
// import { AuthContext } from '../contexts/AuthContext';
// import { Link } from 'react-router-dom';

// const Dashboard = () => {
//   const { authData, logout } = useContext(AuthContext);

//   // Предполагается, что authData.user содержит информацию о пользователе,
//   // например, email (при необходимости можно добавить имя и другие свойства).
//   const userEmail = authData?.user?.email || 'Пользователь';

//   return (
//     <div className="dashboard-container" style={{ padding: '20px', fontFamily: 'Arial, sans-serif' }}>
//       <header style={{ marginBottom: '20px' }}>
//         <h1>Добро пожаловать, {userEmail}!</h1>
//         <nav style={{ display: 'flex', gap: '15px', marginTop: '10px' }}>
//           <Link to="/dashboard">Домой</Link>
//           <Link to="/profile">Мой профиль</Link>
//           <Link to="/recommendations">Рекомендации</Link>
//           <Link to="/chats">Чаты</Link>
//           <Link to="/connections">Подключения</Link>
//         </nav>
//         <button
//           onClick={logout}
//           style={{
//             marginTop: '10px',
//             padding: '8px 12px',
//             backgroundColor: '#f44336',
//             color: '#fff',
//             border: 'none',
//             borderRadius: '4px',
//             cursor: 'pointer'
//           }}
//         >
//           Выйти
//         </button>
//       </header>
      
//       <section>
//         <h2>Общая информация</h2>
//         <p>Это базовая страница дашборда, которая служит центральным узлом для перехода к различным разделам приложения.</p>
//         <ul>
//           <li><Link to="/profile">Просмотреть и редактировать свой профиль</Link></li>
//           <li><Link to="/recommendations">Посмотреть рекомендации для подключений</Link></li>
//           <li><Link to="/chats">Открыть чаты и проверить новые сообщения</Link></li>
//           <li><Link to="/connections">Просмотреть свои существующие подключения</Link></li>
//         </ul>
//       </section>
//     </div>
//   );
// };

// export default Dashboard;
