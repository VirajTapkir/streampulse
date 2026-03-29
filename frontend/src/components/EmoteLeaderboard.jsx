import { useEffect, useState } from "react";

// EmoteLeaderboard tracks how many events each user has triggered
// and displays them sorted by count — highest first
export default function EmoteLeaderboard({ lastMessage }) {
  const [leaderboard, setLeaderboard] = useState({});

  useEffect(() => {
    if (!lastMessage || lastMessage.type === "momentum") return;

    const username = lastMessage.username;

    // increment this user's count by 1
    setLeaderboard((prev) => ({
      ...prev,
      [username]: (prev[username] || 0) + 1,
    }));
  }, [lastMessage]);

  // convert the object to an array and sort by count descending
  const sorted = Object.entries(leaderboard)
    .sort((a, b) => b[1] - a[1])
    .slice(0, 10); // top 10 only

  const maxCount = sorted[0]?.[1] || 1; // for progress bar scaling

  return (
    <div style={styles.container}>
      <h2 style={styles.title}>Top Chatters</h2>
      {sorted.length === 0 && (
        <p style={styles.empty}>Waiting for events...</p>
      )}
      {sorted.map(([username, count], index) => (
        <div key={username} style={styles.row}>
          <span style={styles.rank}>#{index + 1}</span>
          <span style={styles.username}>{username}</span>
          <div style={styles.barContainer}>
            <div style={{
              ...styles.bar,
              width: `${(count / maxCount) * 100}%`,
            }}/>
          </div>
          <span style={styles.count}>{count}</span>
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
  },
  title:    { color: "#cdd6f4", fontSize: 16, fontWeight: 500, marginBottom: 12 },
  empty:    { color: "#6c7086", fontSize: 14 },
  row: {
    display: "flex",
    alignItems: "center",
    gap: 10,
    marginBottom: 10,
  },
  rank:     { color: "#6c7086", fontSize: 12, width: 24 },
  username: { color: "#cdd6f4", fontSize: 14, width: 120 },
  barContainer: {
    flex: 1,
    background: "#313244",
    borderRadius: 4,
    height: 8,
    overflow: "hidden",
  },
  bar: {
    height: "100%",
    background: "#cba6f7",
    borderRadius: 4,
    transition: "width 0.3s ease",
  },
  count: { color: "#a6adc8", fontSize: 12, width: 24, textAlign: "right" },
};