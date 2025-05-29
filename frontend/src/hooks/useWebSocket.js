// /src/hooks/useWebSocket.js
import { useEffect, useCallback } from 'react';
import WebSocketService from '../services/websocketService';
import { useAuthState } from '../contexts/AuthContext';

/**
 * useWebSocket.js
 *
 * Custom React hook for integrating WebSocketService with React components.
 * Handles connection, subscription, and exposes memoized send/subscribe helpers.
 * Automatically manages listener lifecycle and user context.
 */
const useWebSocket = (onMessage) => {
  const { user } = useAuthState();

  useEffect(() => {
    /**
     * On mount: connects WebSocket for current user and subscribes to onMessage.
     * On unmount: removes listener. Re-subscribes if user or callback changes.
     */
    if (!user || !onMessage) return;
  
    WebSocketService.connect(user.id);
    WebSocketService.addListener(onMessage);
  
    return () => {
      WebSocketService.removeListener(onMessage);
    };
  }, [user, onMessage]);

  /**
   * Memoized function to send a chat message via WebSocket.
   */
  const sendMessage = useCallback((chatId, content) => {
    WebSocketService.sendMessage(chatId, content);
  }, []);

  /**
   * Memoized function to send typing status via WebSocket.
   */
  const sendTyping = useCallback((chatId, isTyping) => {
    WebSocketService.sendTyping(chatId, isTyping);
  }, []);

  /**
   * Memoized function to send heartbeat (online status) via WebSocket.
   */
  const sendHeartbeat = useCallback((isOnline) => {
    WebSocketService.sendHeartbeat(isOnline);
  }, []);

  /**
   * Memoized function to subscribe to a chat via WebSocket.
   */
  const subscribe      = useCallback((chatId) => {
        WebSocketService.subscribe(chatId);
      }, []);
  /**
   * Memoized function to unsubscribe from a chat via WebSocket.
   */
  const unsubscribe    = useCallback((chatId) => {
        WebSocketService.unsubscribe(chatId);
      }, []);

  return { sendMessage, sendTyping, sendHeartbeat, subscribe, unsubscribe };
};

export default useWebSocket;
