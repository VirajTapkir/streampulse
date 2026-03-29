import { useEffect, useRef, useState } from "react";

// useWebSocket connects to the Go backend WebSocket server
// and returns a stream of messages that components can react to
export function useWebSocket(url) {
  const [lastMessage, setLastMessage] = useState(null);
  const [connected, setConnected]     = useState(false);
  const wsRef = useRef(null);

  useEffect(() => {
    // create the WebSocket connection when the component mounts
    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onopen = () => {
      console.log("WebSocket connected");
      setConnected(true);
    };

    ws.onmessage = (event) => {
      // parse the JSON and pass it to whoever is listening
      try {
        const data = JSON.parse(event.data);
        setLastMessage(data);
      } catch (e) {
        console.error("failed to parse message", e);
      }
    };

    ws.onclose = () => {
      console.log("WebSocket disconnected");
      setConnected(false);
    };

    ws.onerror = (err) => {
      console.error("WebSocket error", err);
    };

    // cleanup — close the connection when the component unmounts
    return () => ws.close();
  }, [url]);

  return { lastMessage, connected };
}