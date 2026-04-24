import { useEffect, useRef, useState, useCallback } from 'react';
import type { WSMessage } from './types';

export function useWebSocket(url: string | null) {
  const ws = useRef<WebSocket | null>(null);
  const [connected, setConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState<WSMessage | null>(null);
  const listeners = useRef<Map<string, Set<(msg: WSMessage) => void>>>(new Map());

  useEffect(() => {
    if (!url) return;
    const socket = new WebSocket(url);
    socket.onopen = () => setConnected(true);
    socket.onclose = () => setConnected(false);
    socket.onmessage = (e) => {
      try {
        const msg: WSMessage = JSON.parse(e.data);
        setLastMessage(msg);
        listeners.current.get(msg.type)?.forEach(cb => cb(msg));
      } catch {}
    };
    ws.current = socket;
    return () => { socket.close(); setConnected(false); };
  }, [url]);

  const send = useCallback((msg: WSMessage) => {
    if (ws.current?.readyState === WebSocket.OPEN) ws.current.send(JSON.stringify(msg));
  }, []);

  const on = useCallback((type: string, cb: (msg: WSMessage) => void) => {
    if (!listeners.current.has(type)) listeners.current.set(type, new Set());
    listeners.current.get(type)!.add(cb);
    return () => { listeners.current.get(type)?.delete(cb); };
  }, []);

  return { connected, send, on, lastMessage };
}
