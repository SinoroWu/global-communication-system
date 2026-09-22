import type { WSMessage } from '../types';

type MessageHandler = (data: any) => void;

class WebSocketService {
  private ws: WebSocket | null = null;
  private url: string;
  private handlers: Map<string, Set<MessageHandler>> = new Map();
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 20;
  private reconnectTimeout: number | null = null;
  private pingInterval: number | null = null;
  public isConnected = false;
  private statusListeners: Set<(connected: boolean) => void> = new Set();

  constructor() {
    const wsProto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const port = window.location.port === '5173' ? ':8080' : (window.location.port ? `:${window.location.port}` : '');
    const host = import.meta.env.VITE_WS_URL || `${wsProto}//${window.location.hostname}${port}/api/ws`;
    this.url = host;
  }

  public connect() {
    if (this.ws && (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING)) {
      return;
    }

    try {
      this.ws = new WebSocket(this.url);

      this.ws.onopen = () => {
        this.isConnected = true;
        this.reconnectAttempts = 0;
        this.notifyStatus(true);
        console.log('[WS] Connected to Global Comm Realtime Hub');
        this.startHeartbeat();
      };

      this.ws.onmessage = (event) => {
        try {
          const lines = event.data.split('\n');
          for (const line of lines) {
            if (!line.trim()) continue;
            const message: WSMessage = JSON.parse(line);
            this.dispatch(message.type, message.data);
          }
        } catch (err) {
          console.error('[WS] Parse message error:', err);
        }
      };

      this.ws.onclose = () => {
        this.isConnected = false;
        this.notifyStatus(false);
        this.stopHeartbeat();
        this.scheduleReconnect();
      };

      this.ws.onerror = (err) => {
        console.warn('[WS] Error:', err);
        this.ws?.close();
      };
    } catch (err) {
      console.error('[WS] Connection exception:', err);
      this.scheduleReconnect();
    }
  }

  public on(type: string, handler: MessageHandler) {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, new Set());
    }
    this.handlers.get(type)!.add(handler);
    return () => {
      this.handlers.get(type)?.delete(handler);
    };
  }

  public onStatusChange(callback: (connected: boolean) => void) {
    this.statusListeners.add(callback);
    callback(this.isConnected);
    return () => {
      this.statusListeners.delete(callback);
    };
  }

  private dispatch(type: string, data: any) {
    const listeners = this.handlers.get(type);
    if (listeners) {
      listeners.forEach((fn) => fn(data));
    }
  }

  private notifyStatus(status: boolean) {
    this.statusListeners.forEach((fn) => fn(status));
  }

  private scheduleReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('[WS] Max reconnect attempts reached');
      return;
    }

    const delay = Math.min(1000 * Math.pow(1.5, this.reconnectAttempts), 10000);
    this.reconnectAttempts++;

    if (this.reconnectTimeout) {
      window.clearTimeout(this.reconnectTimeout);
    }

    this.reconnectTimeout = window.setTimeout(() => {
      console.log(`[WS] Attempting reconnect (${this.reconnectAttempts})...`);
      this.connect();
    }, delay);
  }

  private startHeartbeat() {
    this.stopHeartbeat();
    this.pingInterval = window.setInterval(() => {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify({ type: 'ping' }));
      }
    }, 20000);
  }

  private stopHeartbeat() {
    if (this.pingInterval) {
      window.clearInterval(this.pingInterval);
      this.pingInterval = null;
    }
  }
}

export const wsService = new WebSocketService();
