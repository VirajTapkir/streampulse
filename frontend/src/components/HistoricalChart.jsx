import { useEffect, useState } from "react";
import {
  BarChart, Bar, XAxis, YAxis,
  CartesianGrid, Tooltip, Legend,
  ResponsiveContainer
} from "recharts";


export default function HistoricalChart({ streamerID }) {
  const [data, setData]       = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError]     = useState(null);

  useEffect(() => {
    setLoading(true);
    setError(null);

    fetch(`http://localhost:8080/api/analytics?streamer_id=${streamerID}&days=7`)
      .then(r => r.json())
      .then(raw => {

        const byDay = {};

        raw.forEach(row => {
          if (!byDay[row.day]) {
            byDay[row.day] = { day: row.day, sub: 0, bits: 0, donation: 0 };
          }
          byDay[row.day][row.event_type] = parseFloat(row.total.toFixed(2));
        });

        setData(Object.values(byDay));
        setLoading(false);
      })
      .catch(err => {
        setError("Failed to load historical data");
        setLoading(false);
      });
  }, [streamerID]); 

  return (
    <div style={styles.container}>
      <div style={styles.header}>
        <h2 style={styles.title}>7-day Revenue Breakdown</h2>
        <span style={styles.subtitle}>by event type</span>
      </div>

      {loading && <p style={styles.status}>Loading...</p>}
      {error   && <p style={styles.error}>{error}</p>}

      {!loading && !error && data.length === 0 && (
        <p style={styles.status}>No earnings data yet — check back soon!</p>
      )}

      {!loading && !error && data.length > 0 && (
        <ResponsiveContainer width="100%" height={260}>
          <BarChart data={data} margin={{ top: 8, right: 16, left: 0, bottom: 0 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#313244"/>
            <XAxis
              dataKey="day"
              tick={{ fill: "#a6adc8", fontSize: 11 }}
            />
            <YAxis
              tick={{ fill: "#a6adc8", fontSize: 11 }}
              label={{
                value: "USD $",
                angle: -90,
                position: "insideLeft",
                fill: "#6c7086",
                fontSize: 11,
              }}
            />
            <Tooltip
              contentStyle={{
                background:   "#313244",
                border:       "none",
                borderRadius: 8,
              }}
              labelStyle={{ color: "#cdd6f4" }}
              formatter={(value) => [`$${value.toFixed(2)}`, undefined]}
            />
            <Legend
              wrapperStyle={{ color: "#a6adc8", fontSize: 12 }}
            />
            <Bar dataKey="sub"      name="Subscriptions" fill="#a78bfa" radius={[4,4,0,0]}/>
            <Bar dataKey="bits"     name="Bits"          fill="#38bdf8" radius={[4,4,0,0]}/>
            <Bar dataKey="donation" name="Donations"     fill="#34d399" radius={[4,4,0,0]}/>
          </BarChart>
        </ResponsiveContainer>
      )}
    </div>
  );
}

const styles = {
  container: {
    background:   "#1e1e2e",
    borderRadius: 12,
    padding:      16,
    marginBottom: 16,
  },
  header: {
    display:        "flex",
    alignItems:     "baseline",
    gap:            10,
    marginBottom:   12,
  },
  title:    { color: "#cdd6f4", fontSize: 16, fontWeight: 500, margin: 0 },
  subtitle: { color: "#6c7086", fontSize: 12 },
  status:   { color: "#6c7086", fontSize: 14 },
  error:    { color: "#f38ba8", fontSize: 14 },
};