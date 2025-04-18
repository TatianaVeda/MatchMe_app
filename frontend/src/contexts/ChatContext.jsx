// /src/contexts/ChatContext.jsx
import React, { createContext, useContext, useReducer, useCallback } from 'react';
import { useAuthState } from './AuthContext';
import useWebSocket from '../hooks/useWebSocket';

const ChatStateContext = createContext();
const ChatDispatchContext = createContext();

const initialState = {
  chats: [],
  activeChat: null,
  messages: {},       // { [chatId]: [msg1, msg2, ...] }
  typingStatuses: {}, // { [chatId]: true/false }
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
        messages: { ...state.messages, [action.chatId]: action.payload },
      };
    case 'ADD_MESSAGE':
      return {
        ...state,
        messages: {
          ...state.messages,
          [action.chatId]: state.messages[action.chatId]
            ? [...state.messages[action.chatId], action.message]
            : [action.message],
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
  const { user } = useAuthState();
  const [state, dispatch] = useReducer(chatReducer, initialState);

  // Обработчик входящих WS-сообщений
  const handleIncoming = useCallback((data) => {
    switch (data.type) {
      case 'message':
        dispatch({ type: 'ADD_MESSAGE', chatId: data.chat_id, message: data });
        break;
      case 'typing':
        dispatch({ type: 'SET_TYPING_STATUS', chatId: data.chat_id, status: data.is_typing });
        break;
      default:
        break;
    }
  }, [dispatch]);
  

  // Хук управляет подключением и подпиской
  const { sendMessage, sendTyping } = useWebSocket(handleIncoming);

  const setChats = (chats) => dispatch({ type: 'SET_CHATS', payload: chats });
  const setActiveChat = (chatId) => dispatch({ type: 'SET_ACTIVE_CHAT', payload: chatId });
  const setMessages = (chatId, msgs) => dispatch({ type: 'SET_MESSAGES', chatId, payload: msgs });

  return (
    <ChatStateContext.Provider value={state}>
      <ChatDispatchContext.Provider
        value={{ setChats, setActiveChat, setMessages, sendMessage, sendTyping }}
      >
        {children}
      </ChatDispatchContext.Provider>
    </ChatStateContext.Provider>
  );
};

export const useChatState = () => useContext(ChatStateContext);
export const useChatDispatch = () => useContext(ChatDispatchContext);
