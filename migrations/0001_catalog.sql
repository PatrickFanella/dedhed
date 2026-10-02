CREATE TABLE IF NOT EXISTS recordings (
  id text PRIMARY KEY,
  source_hash text NOT NULL,
  source jsonb NOT NULL,
  review jsonb NOT NULL DEFAULT '{}',
  playback jsonb,
  media text NOT NULL DEFAULT '',
  media_sha256 text NOT NULL DEFAULT '',
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS source_snapshots (
  sha256 text PRIMARY KEY,
  item_id text NOT NULL,
  raw jsonb NOT NULL,
  observed_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS station_schedule (
  id integer PRIMARY KEY CHECK (id = 1),
  schedule jsonb NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now()
);
