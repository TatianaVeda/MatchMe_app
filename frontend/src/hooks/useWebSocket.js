// /src/hooks/useWebSocket.js
import { useEffect, useCallback } from 'react';
import WebSocketService from '../services/websocketService';
import { useAuthState } from '../contexts/AuthContext';

const useWebSocket = (onMessage) => {
  const { user } = useAuthState();

  useEffect(() => {
    if (!user) return;

    // При монтировании — подключаемся и регистрируем обработчик
    WebSocketService.connect(user.id);
    if (onMessage) WebSocketService.addListener(onMessage);

    return () => {
      // При размонтировании — убираем слушатель и отключаемся
      if (onMessage) WebSocketService.removeListener(onMessage);
      WebSocketService.disconnect();
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

  return { sendMessage, sendTyping, sendHeartbeat };
};

export default useWebSocket;
