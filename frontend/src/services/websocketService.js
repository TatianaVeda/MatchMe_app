/**
 * websocketService.js
 *
 * Singleton service for managing WebSocket connection, message dispatch,
 * reconnect logic, heartbeat, and subscriptions for real-time chat and presence.
 * Handles auto-reconnect, listener management, and message serialization.
 */
class WebSocketService {
  constructor() {
    this.socket = null;
    this.listeners = new Set();
    this.heartbeatInterval = null;
    this.reconnectTimeout = null;
    this.isConnected = false;
    this.userID = null;
    this.reconnectDelay = 5000;
    this.eventQueue = [];
  }
  connect(userID) {
    /**
     * Establishes WebSocket connection for the given user.
     * Handles reconnect logic, heartbeat, and sets up event listeners.
     */
    if (this.isConnected || this.socket) return;
    const wsURL = process.env.REACT_APP_WS_URL || 'ws://localhost:8081/ws';
    this.userID = userID;
    this.socket = new WebSocket(`${wsURL}?userID=${userID}`);
    this.socket.onopen = () => {
      this.isConnected = true;
      if (this.heartbeatInterval) clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = setInterval(() => this.sendHeartbeat(true), 30000);
      // Send all events from queque
      while (this.eventQueue.length > 0) {
        const payload = this.eventQueue.shift();
        this.send(payload);
      }
    };
    this.socket.onmessage = ({ data }) => {
      /**
       * Handles incoming WebSocket messages, parses JSON, dispatches to listeners.
       */
      try {
        const parsed = JSON.parse(data);
        this.listeners.forEach(cb => cb(parsed));
      } catch (err) {
        console.error('Invalid WS message:', err);
      }
    };
    this.socket.onclose = () => {
      /**
       * Handles socket close: sends offline heartbeat, cleans up, schedules reconnect.
       */
      console.warn('WebSocket closed');
      this.sendHeartbeat(false);
      this.cleanupSocket();
      this.scheduleReconnect();
    };
    this.socket.onerror = (err) => {
      console.error('WebSocket error:', err);
    };
  }
  cleanupSocket() {
    if (this.heartbeatInterval) {
           clearInterval(this.heartbeatInterval);
           this.heartbeatInterval = null;
         }
    this.socket = null;
    this.isConnected = false;
  }
  scheduleReconnect() {
    /**
     * Schedules a reconnect attempt after a delay if not already scheduled.
     */
    if (this.reconnectTimeout || !this.userID) return;
    console.log(`Attempting reconnect in ${this.reconnectDelay / 1000}s...`);
    this.reconnectTimeout = setTimeout(() => {
      this.reconnectTimeout = null;
      this.connect(this.userID);
    }, this.reconnectDelay);
  }
  send(payload) {
    /**
     * Sends a JSON-serialized payload over the WebSocket if open.
     * Logs and warns if socket is not open.
     */
    if (this.socket?.readyState === WebSocket.OPEN) {
     console.debug('WS SEND ➔', payload);
      this.socket.send(JSON.stringify(payload));
    } else {
      console.warn('WebSocket is not open. Queuing event:', payload);
      this.eventQueue.push(payload);
    }
  }
  sendMessage(chatId, content) {
    this.send({ action: 'message', chat_id: String(chatId), content });
  }
  sendTyping(chatId, isTyping) {
    this.send({ action: 'typing', chat_id: String(chatId), is_typing: isTyping });
  }
   
  sendRead(chatId) {
    this.send({ action: 'read', chat_id: String(chatId) });
  }
  sendHeartbeat(isOnline) {
    this.send({ action: 'heartbeat', is_online: isOnline });
  }
  subscribe(chatId) {
      this.send({ action: 'subscribe', chat_id: String(chatId) });
    }
  unsubscribe(chatId) {
      this.send({ action: 'unsubscribe', chat_id: String(chatId) });
    }
  addListener(cb) {
    if (!this.listeners.has(cb)) {
      this.listeners.add(cb);
    }
  }
  removeListener(cb) {
    this.listeners.delete(cb);
  }
  disconnect() {
    /**
     * Closes the WebSocket connection and cancels any scheduled reconnects.
     */
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }
    if (this.socket) {
      this.socket.close();
      this.cleanupSocket();
    }
  }
}
export default new WebSocketService();