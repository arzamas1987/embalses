CREATE TABLE IF NOT EXISTS readings (
    id SERIAL PRIMARY KEY,
    reservoir_id INTEGER NOT NULL REFERENCES reservoirs(id) ON DELETE CASCADE,
    source_id INTEGER REFERENCES sources(id),
    observed_at DATE NOT NULL,
    volume_hm3 NUMERIC,
    capacity_hm3 NUMERIC,
    fill_pct NUMERIC,
    weekly_variation_hm3 NUMERIC,
    is_provisional BOOLEAN DEFAULT FALSE,
    is_official BOOLEAN DEFAULT TRUE,
    published_at TIMESTAMPTZ,
    fetched_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(reservoir_id, observed_at, source_id)
);

CREATE INDEX IF NOT EXISTS idx_readings_reservoir ON readings(reservoir_id);
CREATE INDEX IF NOT EXISTS idx_readings_observed ON readings(observed_at);
CREATE INDEX IF NOT EXISTS idx_readings_source ON readings(source_id);

CREATE TABLE IF NOT EXISTS api_keys (
    id SERIAL PRIMARY KEY,
    key_hash TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    tier TEXT NOT NULL DEFAULT 'free',
    daily_quota INTEGER NOT NULL DEFAULT 100,
    rate_limit_per_minute INTEGER NOT NULL DEFAULT 60,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS metering (
    id SERIAL PRIMARY KEY,
    api_key_id INTEGER REFERENCES api_keys(id),
    endpoint TEXT NOT NULL,
    method TEXT NOT NULL,
    status_code INTEGER,
    response_time_ms INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_metering_key ON metering(api_key_id);
CREATE INDEX IF NOT EXISTS idx_metering_created ON metering(created_at);

-- Seed a test API key (hash of 'test-key-123')
INSERT INTO api_keys (key_hash, name, tier, daily_quota, rate_limit_per_minute)
VALUES ('test-key-123', 'Test Key', 'free', 1000, 120)
ON CONFLICT (key_hash) DO NOTHING;
