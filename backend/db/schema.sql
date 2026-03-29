CREATE TABLE IF NOT EXISTS streamers (
    id           SERIAL PRIMARY KEY,
    username     VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(100),
    created_at   TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS earnings (
    id          SERIAL PRIMARY KEY,
    streamer_id INTEGER REFERENCES streamers(id),
    event_type  VARCHAR(50) NOT NULL,
    amount      NUMERIC(10, 2) NOT NULL,
    occurred_at TIMESTAMP DEFAULT NOW()
);

INSERT INTO streamers (username, display_name)
VALUES ('teststreamer', 'Test Streamer')
ON CONFLICT DO NOTHING;