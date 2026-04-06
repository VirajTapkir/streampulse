import AlertFeed        from "./components/AlertFeed";
import RevenueChart     from "./components/RevenueChart";
import EmoteLeaderboard from "./components/EmoteLeaderboard";
import MomentumScore    from "./components/MomentumScore";
import { useWebSocket } from "./hooks/useWebSocket";

export default function App() {
  // single WebSocket connection shared across all four panels
  const { lastMessage, connected, connecting, retryDelay, retryCount } = useWebSocket("ws://localhost:8080/ws");

  return (
    <div style={styles.page}>

      {/* connecting spinner — shown on first load before WS handshake */}
      {connecting && (
        <div style={styles.banner}>
          <span style={styles.spinner}>⟳</span>
          Connecting to StreamPulse backend...
        </div>
      )}

      {/* disconnected banner — shown when backend goes down */}
      {!connecting && !connected && (
        <div style={{ ...styles.banner, ...styles.errorBanner }}>
            ⚠ Backend unavailable — retrying in {retryDelay}s (attempt {retryCount})
        </div>
      )}

      {/* header */}
      <div style={styles.header}>
        <h1 style={styles.logo}>StreamPulse</h1>
        {connected && (
          <div style={{ ...styles.status, background: "#a6e3a1" }}>
              Live
          </div>
        )}
      </div>

      {/* top row — momentum score takes full width */}
      <div style={styles.topRow}>
        <MomentumScore lastMessage={lastMessage} connected={connected}/>

      </div>

      {/* middle row — revenue chart full width */}
      <div style={styles.midRow}>
        <RevenueChart lastMessage={lastMessage}/>
      </div>

      {/* bottom row — alert feed and leaderboard side by side */}
      <div style={styles.bottomRow}>
        <div style={styles.half}>
          <AlertFeed lastMessage={lastMessage}/>
        </div>
        <div style={styles.half}>
          <EmoteLeaderboard lastMessage={lastMessage}/>
        </div>
      </div>

    </div>
  );
}

const styles = {
  page: {
    background:  "#181825",
    minHeight:   "100vh",
    padding:     24,
    fontFamily:  "system-ui, sans-serif",
    boxSizing:   "border-box",
  },
  header: {
    display:        "flex",
    alignItems:     "center",
    justifyContent: "space-between",
    marginBottom:   24,
  },
  logo: {
    color:      "#cdd6f4",
    fontSize:   28,
    fontWeight: 500,
    margin:     0,
  },
  status: {
    borderRadius: 20,
    padding:      "4px 14px",
    fontSize:     13,
    fontWeight:   500,
    color:        "#1e1e2e",
  },
  topRow:    { marginBottom: 16 },
  midRow:    { marginBottom: 16 },
  bottomRow: { display: "flex", gap: 16 },
  half: { flex: 1, minWidth: 0 },
  banner: {
    background:   "#f9e2af",
    color:        "#1e1e2e",
    borderRadius: 8,
    padding:      "10px 16px",
    marginBottom: 16,
    fontSize:     14,
    fontWeight:   500,
    display:      "flex",
    alignItems:   "center",
    gap:          8,
  },
  errorBanner: {
    background: "#f38ba8",
  },
  spinner: {
    display:         "inline-block",
    animation:       "spin 1s linear infinite",
    fontSize:        18,
  },
};