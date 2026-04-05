import { useEffect, useRef, useState, useCallback } from "react";

export function useWebSocket(url) {
  const [lastMessage, setLastMessage] = useState(null);
  const [connected, setConnected]   = useState(false);
  const [connecting, setConnecting] = useState(true);
  const [retryDelay, setRetryDelay] = useState(null);

  const wsRef       = useRef(null);
  const retryCount  = useRef(0);
  const retryTimer  = useRef(null);

  const connect = useCallback(() => {
    // clean up any existing connection before creating a new one
    if (wsRef.current) {
      wsRef.current.close();
    }

    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onopen = () => {
      console.log("WebSocket connected");
      setConnected(true);
      setConnecting(false);
      setRetryDelay(null);
      retryCount.current = 0;
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        setLastMessage(data);
      } catch (e) {
        console.error("failed to parse message", e);
      }
    };

    ws.onclose = () => {
      setConnected(false);
      setConnecting(false);
      const delay = Math.min(1000 * 2 ** retryCount.current, 30000);
      retryCount.current += 1;
      setRetryDelay(delay / 1000);
      console.log(`retrying in ${delay / 1000}s (attempt ${retryCount.current})`);
      retryTimer.current = setTimeout(connect, delay);
    };


    ws.onerror = (err) => {
      console.error("WebSocket error", err);
      // onclose fires automatically after onerror
      // so reconnection is handled there
    };
  }, [url]);

  useEffect(() => {
    connect();

    // cleanup — cancel any pending retry and close the connection
    return () => {
      clearTimeout(retryTimer.current);
      if (wsRef.current) wsRef.current.close();
    };
  }, [connect]);

  return { lastMessage, connected, connecting, retryDelay };
}