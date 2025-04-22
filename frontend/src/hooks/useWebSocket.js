// /src/hooks/useWebSocket.js
import { useEffect, useCallback } from 'react';
import WebSocketService from '../services/websocketService';
import { useAuthState } from '../contexts/AuthContext';

const useWebSocket = (onMessage) => {
  const { user } = useAuthState();

  useEffect(() => {
      // Подключаемся и подписываемся ТОЛЬКО когда есть обработчик onMessage
      if (!user || !onMessage) return;
    
      WebSocketService.connect(user.id);
      WebSocketService.addListener(onMessage);
    
      return () => {
        WebSocketService.removeListener(onMessage);
       // WebSocketService.disconnect();
      };
  }, [user, onMessage]);

  // Обёртки для методов отправки
  const sendMessage = useCallback((chatId, content) => {
    WebSocketService.sendMessage(chatId, content);
  }, []);

  const sendTyping = useCallback((chatId, isTyping) => {
    WebSocketService.sendTyping(chatId, isTyping);
  }, []);

  const sendHeartbeat = useCallback((isOnline) => {
    WebSocketService.sendHeartbeat(isOnline);
  }, []);

  const subscribe      = useCallback((chatId) => {
        WebSocketService.subscribe(chatId);
      }, []);
      const unsubscribe    = useCallback((chatId) => {
        WebSocketService.unsubscribe(chatId);
      }, []);

  return { sendMessage, sendTyping, sendHeartbeat, subscribe, unsubscribe };
};

export default useWebSocket;
