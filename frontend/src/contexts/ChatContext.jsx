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
  presence: {},
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
      case 'SET_PRESENCE':
     return {
       ...state,
       presence: {
         ...state.presence,
         [action.userId]: action.isOnline
       }
     };
    default:
      return state;
  }
}
export const ChatProvider = ({ children }) => {
  const { user } = useAuthState(); 
  const [state, dispatch] = useReducer(chatReducer, initialState);
  const handleIncoming = useCallback(
    (data) => {
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
          if (data.user_id && data.user_id !== user.id) {
            dispatch({
              type: 'SET_TYPING_STATUS',
              chatId: data.chat_id,
              status: data.is_typing,
            });
          }
          break;
          case 'presence':
       dispatch({
         type: 'SET_PRESENCE',
         userId: data.user_id,
         isOnline: data.is_online
       });
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