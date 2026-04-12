import { useEffect, useState } from "react";

export default function MomentumScore({ lastMessage, connected, streamerID }) {
  const [score, setScore]         = useState(null);
  const [detail, setDetail]       = useState(null);
  const [lastSeen, setLastSeen]   = useState(null);
  const [threshold, setThreshold] = useState(10); // configurable warning threshold
  const [alert, setAlert]         = useState(null); // null | "warning" | "critical"

  useEffect(() => {
    if (!lastMessage || lastMessage.type !== "momentum") return;

    const msgStreamerID = lastMessage._meta?.streamer_id;
    if (msgStreamerID && msgStreamerID !== streamerID) return;

    const s = lastMessage.payload.score;
    setScore(s.toFixed(1));
    setDetail(lastMessage.payload);
    setLastSeen(new Date().toLocaleTimeString());

    // check thresholds and set alert level
    if (s <= 0) {
      setAlert("critical");
    } else if (s < threshold) {
      setAlert("warning");
    } else {
      setAlert(null);
    }
  }, [lastMessage, streamerID, threshold]);

  useEffect(() => {
    if (!connected) {
      setScore(null);
      setDetail(null);
      setLastSeen(null);
      setAlert(null);
    }
  }, [connected]);

  // colour the score number based on alert level
  const scoreColour =
    !score          ? "#6c7086" :
    alert === "critical" ? "#f38ba8" :
    alert === "warning"  ? "#f9e2af" :
                           "#a6e3a1";

  return (
    <div style={styles.container}>

      {/* alert banner — only shown when threshold is breached */}
      {alert === "critical" && (
        <div style={{ ...styles.alertBanner, ...styles.critical }}>
          🚨 Stream engagement critically low — momentum near zero!
        </div>
      )}
      {alert === "warning" && (
        <div style={{ ...styles.alertBanner, ...styles.warning }}>
          ⚠ Stream momentum below threshold ({threshold}) — engagement dropping
        </div>
      )}

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

      {/* threshold configuration */}
      <div style={styles.thresholdRow}>
        <span style={styles.thresholdLabel}>
          Warning threshold: <strong style={{ color: "#cdd6f4" }}>{threshold}</strong>
        </span>
        <input
          type="range"
          min="0"
          max="30"
          value={threshold}
          onChange={e => setThreshold(Number(e.target.value))}
          style={styles.slider}
        />
      </div>

    </div>
  );
}

const styles = {
  container: {
    background:   "#1e1e2e",
    borderRadius: 12,
    padding:      16,
    textAlign:    "center",
  },
  alertBanner: {
    borderRadius: 8,
    padding:      "10px 16px",
    marginBottom: 12,
    fontSize:     14,
    fontWeight:   500,
    textAlign:    "left",
  },
  critical: {
    background: "#f38ba8",
    color:      "#1e1e2e",
  },
  warning: {
    background: "#f9e2af",
    color:      "#1e1e2e",
  },
  title:        { color: "#cdd6f4", fontSize: 16, fontWeight: 500, marginBottom: 8 },
  scoreDisplay: { fontSize: 64, fontWeight: 500, margin: "16px 0" },
  breakdown: {
    display:        "flex",
    justifyContent: "space-around",
    marginBottom:   12,
  },
  stat:  { display: "flex", flexDirection: "column", alignItems: "center", gap: 4 },
  label: { color: "#6c7086", fontSize: 11, textTransform: "uppercase" },
  value: { color: "#cdd6f4", fontSize: 16, fontWeight: 500 },
  hint:  { color: "#6c7086", fontSize: 11, marginBottom: 12 },
  thresholdRow: {
    display:        "flex",
    alignItems:     "center",
    justifyContent: "center",
    gap:            12,
    marginTop:      8,
    paddingTop:     12,
    borderTop:      "1px solid #313244",
  },
  thresholdLabel: {
    color:    "#6c7086",
    fontSize: 12,
    minWidth: 180,
  },
  slider: {
    width:  140,
    cursor: "pointer",
  },
};