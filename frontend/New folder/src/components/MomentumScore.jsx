import { useEffect, useState } from "react";

// MomentumScore displays the live engagement score
// it updates every 5 seconds when a momentum message arrives
export default function MomentumScore({ lastMessage, connected, streamerID }) {
  const [score, setScore]     = useState(null);
  const [detail, setDetail]   = useState(null);
  const [lastSeen, setLastSeen] = useState(null);

  useEffect(() => {
      if (!lastMessage || lastMessage.type !== "momentum") return;

      // only update if this momentum message belongs to the selected streamer
      const msgStreamerID = lastMessage._meta?.streamer_id;
      if (msgStreamerID && msgStreamerID !== streamerID) return;

      setScore(lastMessage.payload.score.toFixed(1));
      setDetail(lastMessage.payload);
      setLastSeen(new Date().toLocaleTimeString());
  }, [lastMessage, streamerID]);


  // clear stale data when disconnected so we never show old values as live
  useEffect(() => {
      if (!connected) {
          setScore(null);
          setDetail(null);
          setLastSeen(null);
      }
  }, [connected]);


  // colour the score based on how high it is
  const scoreColour =
    !score      ? "#6c7086" :
    score > 20  ? "#a6e3a1" :  // green = high engagement
    score > 10  ? "#f9e2af" :  // yellow = medium
                  "#f38ba8";   // red = low

  return (
    <div style={styles.container}>
      <h2 style={styles.title}>Stream Momentum</h2>
      <div style={{ ...styles.scoreDisplay, color: scoreColour }}>
        {score ?? "—"}
      </div>
      {detail && (
        <div style={styles.breakdown}>
          <div style={styles.stat}>
            <span style={styles.label}>Subs/min</span>
            <span style={styles.value}>{detail.sub_rate.toFixed(1)}</span>
          </div>
          <div style={styles.stat}>
            <span style={styles.label}>Bits/min</span>
            <span style={styles.value}>{detail.bits_per_min.toFixed(1)}</span>
          </div>
          <div style={styles.stat}>
            <span style={styles.label}>Chat density</span>
            <span style={styles.value}>{detail.chat_density.toFixed(1)}</span>
          </div>
        </div>
      )}
      <p style={styles.hint}>
        {lastSeen ? `Last updated ${lastSeen}` : "Updates every 5 seconds"}
      </p>
    </div>
  );
}

const styles = {
  container: {
    background: "#1e1e2e",
    borderRadius: 12,
    padding: 16,
    textAlign: "center",
  },
  title:        { color: "#cdd6f4", fontSize: 16, fontWeight: 500, marginBottom: 8 },
  scoreDisplay: { fontSize: 64, fontWeight: 500, margin: "16px 0" },
  breakdown: {
    display: "flex",
    justifyContent: "space-around",
    marginBottom: 12,
  },
  stat:  { display: "flex", flexDirection: "column", alignItems: "center", gap: 4 },
  label: { color: "#6c7086", fontSize: 11, textTransform: "uppercase" },
  value: { color: "#cdd6f4", fontSize: 16, fontWeight: 500 },
  hint:  { color: "#6c7086", fontSize: 11 },
};