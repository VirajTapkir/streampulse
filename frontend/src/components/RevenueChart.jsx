import { useEffect, useState } from "react";
import {
  LineChart, Line, XAxis, YAxis,
  CartesianGrid, Tooltip, ResponsiveContainer
} from "recharts";

// RevenueChart shows a real-time line graph of earnings over time
// each new event adds a data point to the chart
export default function RevenueChart({ lastMessage }) {
  const [data, setData]         = useState([]);
  const [total, setTotal]       = useState(0);

  useEffect(() => {
    if (!lastMessage || lastMessage.type === "momentum") return;

    const amount = parseFloat(lastMessage.amount) || 0;

    setTotal((prev) => {
      const newTotal = prev + amount;

      // add a new point to the chart — time on X axis, running total on Y
      setData((points) => [
        ...points,
        {
          time:    new Date().toLocaleTimeString(),
          revenue: parseFloat(newTotal.toFixed(2)),
        },
      ].slice(-30)); // keep last 30 points so the chart stays readable

      return newTotal;
    });
  }, [lastMessage]);

  return (
    <div style={styles.container}>
      <div style={styles.header}>
        <h2 style={styles.title}>Revenue</h2>
        <span style={styles.total}>${total.toFixed(2)}</span>
      </div>
      <ResponsiveContainer width="100%" height={300}>
        <LineChart data={data}>
          <CartesianGrid strokeDasharray="3 3" stroke="#313244"/>
          <XAxis
            dataKey="time"
            tick={{ fill: "#a6adc8", fontSize: 11 }}
            interval="preserveStartEnd"
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
            contentStyle={{ background: "#313244", border: "none", borderRadius: 8 }}
            labelStyle={{ color: "#cdd6f4" }}
            itemStyle={{ color: "#a6e3a1" }}
          />
          <Line
            type="monotone"
            dataKey="revenue"
            stroke="#a6e3a1"
            strokeWidth={2}
            dot={false}
            isAnimationActive={false}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}

const styles = {
  container: {
    background: "#1e1e2e",
    borderRadius: 12,
    padding: 16,
  },
  header: {
    display: "flex",
    justifyContent: "space-between",
    alignItems: "center",
    marginBottom: 12,
  },
  title: { color: "#cdd6f4", fontSize: 16, fontWeight: 500 },
  total: { color: "#a6e3a1", fontSize: 24, fontWeight: 500 },
};