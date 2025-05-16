// // frontend/src/contexts/ChatContext.jsx
// import React, { createContext, useContext, useReducer, useCallback } from 'react';
// import { useAuthState } from './AuthContext';
// import useWebSocket from '../hooks/useWebSocket';


// const ChatStateContext = createContext();
// const ChatDispatchContext = createContext();

// const initialState = {
//   chats: [],           // [{ id, otherUserID, otherUser, lastMessage, unreadCount, otherUserOnline, ... }]
//   activeChat: null,
//   messages: {},        // { [chatId]: [msg, ...] }
//   typingStatuses: {},  // { [chatId]: true/false }
// };

// function chatReducer(state, action) {
//   switch (action.type) {
//     case 'SET_CHATS':
//       return { ...state, chats: action.payload };

//     case 'SET_ACTIVE_CHAT':
//       return {
//         ...state,
//         activeChat: action.payload,
//         // при открытии чата сбросим счётчик непрочитанных
//         chats: state.chats.map(c =>
//           c.id === action.payload
//             ? { ...c, unreadCount: 0 }
//             : c
//         )
//       };

//     case 'SET_MESSAGES':
//       return {
//         ...state,
//         messages: {
//           ...state.messages,
//           [action.chatId]: action.payload,
//         },
//       };

//     case 'ADD_MESSAGE':
//       // для прямой вставки (неиспользуется, но оставляем для совместимости)
//       return {
//         ...state,
//         messages: {
//           ...state.messages,
//           [action.chatId]: state.messages[action.chatId]
//             ? [...state.messages[action.chatId], action.message]
//             : [action.message],
//         },
//       };

//     case 'RECEIVE_MESSAGE': {
//       const { chatId, message } = action;
//       // 1) вставляем сообщение
//       const updatedMessages = state.messages[chatId]
//         ? [...state.messages[chatId], message]
//         : [message];

//       // 2) обновляем чат-лист: lastMessage, unreadCount (если чат не открыт), и поднимаем наверх
//       const updatedChats = state.chats
//         .map(c => {
//           if (c.id !== chatId) return c;
//           return {
//             ...c,
//             lastMessage: message,
//             unreadCount: c.id === state.activeChat ? 0 : c.unreadCount + 1,
//           };
//         })
//         // .sort((a, b) => {
//         //   // самый свежий чат первым
//         //   const ta = new Date(a.lastMessage.timestamp).getTime();
//         //   const tb = new Date(b.lastMessage.timestamp).getTime();
//         //   return tb - ta;
//         // });

//         .sort((a, b) => {
//           const ta = a.lastMessage && a.lastMessage.timestamp
//             ? new Date(a.lastMessage.timestamp).getTime()
//             : 0;
//           const tb = b.lastMessage && b.lastMessage.timestamp
//             ? new Date(b.lastMessage.timestamp).getTime()
//             : 0;
//           return tb - ta;
//         })

//       return {
//         ...state,
//         messages: {
//           ...state.messages,
//           [chatId]: updatedMessages,
//         },
//         chats: updatedChats,
//       };
//     }

//     case 'SET_TYPING_STATUS':
//       return {
//         ...state,
//         typingStatuses: {
//           ...state.typingStatuses,
//           [action.chatId]: action.status,
//         },
//       };

//     case 'HEARTBEAT': {
//       // action.payload = { chatId, otherUserOnline }
//       const { chatId, otherUserOnline } = action.payload;
//       return {
//         ...state,
//         chats: state.chats.map(c =>
//           c.id === chatId ? { ...c, otherUserOnline } : c
//         ),
//       };
//     }

//     default:
//       return state;
//   }
// }

// export const ChatProvider = ({ children }) => {
//   const { user } = useAuthState();
//   const [state, dispatch] = useReducer(chatReducer, initialState);

//   // Обработчик WebSocket-сообщений
//   const handleIncoming = useCallback((data) => {
//     switch (data.type) {
//       case 'message':
//         // data.chat_id, data.timestamp, data.content, data.sender_id, data.read…
//         dispatch({
//           type: 'RECEIVE_MESSAGE',
//           chatId: data.chat_id,
//           message: data,
//         });
//         break;
//       case 'typing':
//         dispatch({
//           type: 'SET_TYPING_STATUS',
//           chatId: data.chat_id,
//           status: data.is_typing,
//         });
//         break;
//       case 'heartbeat':
//         // при получении heartbeat от другого участника
//         dispatch({
//           type: 'HEARTBEAT',
//           payload: {
//             chatId: data.chat_id,
//             otherUserOnline: data.is_online,
//           },
//         });
//         break;
//       default:
//         break;
//     }
//   }, [dispatch]);

//   // WS-сервис «вешает» handleIncoming на входящие события
//   const { sendMessage, sendTyping } = useWebSocket(handleIncoming);

//   const setChats = chats => dispatch({ type: 'SET_CHATS', payload: chats });
//   const setActiveChat = chatId => dispatch({ type: 'SET_ACTIVE_CHAT', payload: chatId });
//   const setMessages = (chatId, msgs) => dispatch({ type: 'SET_MESSAGES', chatId, payload: msgs });

//   return (
//     <ChatStateContext.Provider value={state}>
//       <ChatDispatchContext.Provider value={{
//         setChats,
//         setActiveChat,
//         setMessages,
//         sendMessage,
//         sendTyping
//       }}>
//         {children}
//       </ChatDispatchContext.Provider>
//     </ChatStateContext.Provider>
//   );
// };

// export const useChatState = () => useContext(ChatStateContext);
// export const useChatDispatch = () => useContext(ChatDispatchContext);
// src/contexts/ChatContext.jsx
import React, { createContext, useContext, useReducer, useCallback } from 'react';
import { useAuthState } from './AuthContext';
import useWebSocket from '../hooks/useWebSocket';

const ChatStateContext = createContext();
const ChatDispatchContext = createContext();

const initialState = {
  chats: [],
  activeChat: null,
  messages: {},
  typingStatuses: {},
};

function chatReducer(state, action) {
  switch (action.type) {
    case 'SET_CHATS':
      return { ...state, chats: action.payload };

    case 'SET_ACTIVE_CHAT':
      return { ...state, activeChat: action.payload };

    case 'SET_MESSAGES':
      return {
        ...state,
        messages: {
          ...state.messages,
          [action.chatId]: action.payload,
        },
      };

    case 'RECEIVE_MESSAGE':
      const prevMessages = state.messages[action.chatId] || [];
      return {
        ...state,
        messages: {
          ...state.messages,
          [action.chatId]: [...prevMessages, action.message],
        },
      };

    case 'SET_TYPING_STATUS':
      return {
        ...state,
        typingStatuses: {
          ...state.typingStatuses,
          [action.chatId]: action.status,
        },
      };

    default:
      return state;
  }
}

export const ChatProvider = ({ children }) => {
  const { user } = useAuthState(); // <-- Получаем текущего пользователя
  const [state, dispatch] = useReducer(chatReducer, initialState);

  const handleIncoming = useCallback(
    (data) => {
      // 1. Ждём, пока user будет инициализирован
      if (!user || !user.id) return;

      switch (data.type) {
        case 'message':
          dispatch({
            type: 'RECEIVE_MESSAGE',
            chatId: data.chat_id,
            message: data,
          });
          break;

        case 'typing':
          // 2. Показываем индикатор "печатает" только если это не сам пользователь
          if (data.user_id && data.user_id !== user.id) {
            dispatch({
              type: 'SET_TYPING_STATUS',
              chatId: data.chat_id,
              status: data.is_typing,
            });
          }
          break;

        default:
          break;
      }
    },
    [user]
  );

  const { sendMessage, sendTyping } = useWebSocket(handleIncoming);

  const setChats = (chats) => {
    dispatch({ type: 'SET_CHATS', payload: chats });
  };

  const setActiveChat = (chatId) => {
    dispatch({ type: 'SET_ACTIVE_CHAT', payload: chatId });
  };

  const setMessages = (chatId, messages) => {
    dispatch({ type: 'SET_MESSAGES', chatId, payload: messages });
  };

  return (
    <ChatStateContext.Provider value={state}>
      <ChatDispatchContext.Provider
        value={{
          setChats,
          setActiveChat,
          setMessages,
          sendMessage,
          sendTyping,
        }}
      >
        {children}
      </ChatDispatchContext.Provider>
    </ChatStateContext.Provider>
  );
};

export const useChatState = () => useContext(ChatStateContext);
export const useChatDispatch = () => useContext(ChatDispatchContext);
