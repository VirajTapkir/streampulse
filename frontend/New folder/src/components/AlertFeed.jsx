import { useEffect, useState } from "react";

// AlertFeed shows a live scrolling list of incoming events
// each new event pops in at the top of the list
export default function AlertFeed({ lastMessage, streamerID }) {
  const [alerts, setAlerts]   = useState([]);
  const [lastSeen, setLastSeen] = useState(null);

  useEffect(() => {
    if (!lastMessage || lastMessage.type === "momentum") return;
    if (lastMessage._raw?._meta?.streamer_id && lastMessage._raw._meta.streamer_id !== streamerID) return;

    const newAlert = {
      id:       Date.now(), 
      type:     lastMessage.type,
      username: lastMessage.username,
      amount:   lastMessage.amount,
    };

    // add new alert to the top, keep only the last 20
    setAlerts((prev) => [newAlert, ...prev].slice(0, 20));
    setLastSeen(new Date().toLocaleTimeString());

  }, [lastMessage]);

  // pick an emoji based on event type so alerts are easy to scan
  const icon = { sub: "⭐", bits: "💎", donation: "💰" };

  // pick a colour per event type
  const colour = {
    sub:      "#a78bfa",
    bits:     "#38bdf8",
    donation: "#34d399",
  };

  return (
    <div style={styles.container}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 12 }}>
        <h2 style={{ ...styles.title, marginBottom: 0 }}>Live Alerts</h2>
        {lastSeen && <span style={{ color: "#6c7086", fontSize: 11 }}>last event {lastSeen}</span>}
      </div>

      {alerts.length === 0 && (
        <p style={styles.empty}>Waiting for events...</p>
      )}
      {alerts.map((alert) => (
        <div key={alert.id} style={{
          ...styles.alert,
          borderLeft: `4px solid ${colour[alert.type] || "#888"}`,
        }}>
          <span style={styles.icon}>{icon[alert.type]}</span>
          <span style={styles.username}>{alert.username}</span>
          <span style={styles.type}>{alert.type}</span>
          <span style={styles.amount}>${(alert.amount || 0).toFixed(2)}</span>
        </div>
      ))}
    </div>
  );
}

const styles = {
  container: {
    background: "#1e1e2e",
    borderRadius: 12,
    padding: 16,
    height: 400,
    overflowY: "auto",
  },
  title: {
    color: "#cdd6f4",
    fontSize: 16,
    fontWeight: 500,
    marginBottom: 12,
  },
  empty: {
    color: "#6c7086",
    fontSize: 14,
  },
  alert: {
    display: "flex",
    alignItems: "center",
    gap: 10,
    background: "#313244",
    borderRadius: 8,
    padding: "10px 14px",
    marginBottom: 8,
    animation: "fadeIn 0.3s ease",
  },
  icon:     { fontSize: 18 },
  username: { color: "#cdd6f4", fontWeight: 500, flex: 1 },
  type:     { color: "#a6adc8", fontSize: 12, textTransform: "uppercase" },
  amount:   { color: "#a6e3a1", fontWeight: 500 },
};