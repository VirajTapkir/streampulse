import { useEffect, useRef, useState } from "react";

export function useWebSocket(url) {
  const [lastMessage, setLastMessage] = useState(null);
  const [connected, setConnected]     = useState(false);
  const [connecting, setConnecting]   = useState(true);
  const [retryDelay, setRetryDelay]   = useState(null);

  const wsRef        = useRef(null);
  const retryCount   = useRef(0);
  const retryTimer   = useRef(null);
  const currentUrl   = useRef(url);
  const isMounted    = useRef(true);

  useEffect(() => {
    isMounted.current = true;
    return () => { isMounted.current = false; };
  }, []);

  useEffect(() => {
    // update the current URL ref immediately when it changes
    currentUrl.current = url;
    retryCount.current = 0;

    function cleanup() {
      clearTimeout(retryTimer.current);
      if (wsRef.current) {
        const ws = wsRef.current;
        wsRef.current = null;
        ws.onopen    = null;
        ws.onmessage = null;
        ws.onclose   = null;
        ws.onerror   = null;
        ws.close();
      }
    }

    function openConnection(connectUrl) {
      cleanup();

      if (!isMounted.current) return;

      setConnecting(true);
      setConnected(false);

      const ws = new WebSocket(connectUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        if (wsRef.current !== ws) return; // stale connection
        console.log("WebSocket connected:", connectUrl);
        setConnected(true);
        setConnecting(false);
        setRetryDelay(null);
        retryCount.current = 0;
      };

      ws.onmessage = (event) => {
        if (wsRef.current !== ws) return; // stale connection
        try {
          const raw = JSON.parse(event.data);
          let normalised = raw;

          if (raw.type === "momentum") {
            normalised = raw;
          } else if (raw._meta) {
            normalised = {
              type:      raw._meta.type,
              username:  raw.event?.user_login || "anonymous",
              amount:    raw._meta.amount,
              timestamp: raw._meta.timestamp,
              _raw:      raw,
            };
          }

          setLastMessage(normalised);
        } catch (e) {
          console.error("failed to parse message", e);
        }
      };

      ws.onclose = () => {
        if (wsRef.current !== ws) return; // stale — don't retry
        setConnected(false);
        setConnecting(false);

        const delay = Math.min(1000 * 2 ** retryCount.current, 30000);
        retryCount.current += 1;
        setRetryDelay(delay / 1000);

        console.log(`retrying in ${delay / 1000}s (attempt ${retryCount.current})`);
        retryTimer.current = setTimeout(() => {
          openConnection(currentUrl.current);
        }, delay);
      };

      ws.onerror = () => {
        // onclose fires automatically after onerror
      };
    }

    openConnection(url);

    return cleanup;
  }, [url]); // re-runs every time streamerID changes the URL

  return {
    lastMessage,
    connected,
    connecting,
    retryDelay,
    retryCount: retryCount.current,
  };
}