class WebSocketService {
  constructor() {
    this.socket = null;
    this.listeners = new Set();
    this.heartbeatInterval = null;
    this.reconnectTimeout = null;
    this.isConnected = false;
    this.userID = null;
    this.reconnectDelay = 5000;
  }

  connect(userID) {
    if (this.isConnected || this.socket) return;

    const wsURL = process.env.REACT_APP_WS_URL || 'ws://localhost:8081/ws';
    this.userID = userID;
    this.socket = new WebSocket(`${wsURL}?userID=${userID}`);

    this.socket.onopen = () => {
      console.log('WebSocket connected');
      this.isConnected = true;

      // если вдруг остался старый интервал, почистим его
     if (this.heartbeatInterval) {
         clearInterval(this.heartbeatInterval);
         this.heartbeatInterval = null;
       }

      this.heartbeatInterval = setInterval(() => this.sendHeartbeat(true), 30000);
    };

    this.socket.onmessage = ({ data }) => {
      try {
        const parsed = JSON.parse(data);
        this.listeners.forEach(cb => cb(parsed));
      } catch (err) {
        console.error('Invalid WS message:', err);
      }
    };

    this.socket.onclose = () => {
      console.warn('WebSocket closed');
      this.cleanupSocket();
      this.scheduleReconnect();
    };

    this.socket.onerror = (err) => {
      console.error('WebSocket error:', err);
    };
  }

  cleanupSocket() {
    // clearInterval(this.heartbeatInterval);
    // this.heartbeatInterval = null;

    if (this.heartbeatInterval) {
           clearInterval(this.heartbeatInterval);
           this.heartbeatInterval = null;
         }

    this.socket = null;
    this.isConnected = false;
  }

  scheduleReconnect() {
    if (this.reconnectTimeout || !this.userID) return;

    console.log(`Attempting reconnect in ${this.reconnectDelay / 1000}s...`);
    this.reconnectTimeout = setTimeout(() => {
      this.reconnectTimeout = null;
      this.connect(this.userID);
    }, this.reconnectDelay);
  }

  send(payload) {
    if (this.socket?.readyState === WebSocket.OPEN) {
       // логируем, что именно шлем
     console.debug('WS SEND ➔', payload);
      this.socket.send(JSON.stringify(payload));
    } else {
      console.warn('WebSocket is not open. Failed to send:', payload);
    }
  }

  sendMessage(chatId, content) {
    // Убедись, что сервер ожидает 'message', а не 'new_message'
    this.send({ action: 'message', chat_id: String(chatId), content });
  }

  sendTyping(chatId, isTyping) {
    this.send({ action: 'typing', chat_id: String(chatId), is_typing: isTyping });
  }

  sendHeartbeat(isOnline) {
    this.send({ action: 'heartbeat', is_online: isOnline });
  }

  // новые методы подписки на чат
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
