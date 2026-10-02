-- Collection-wide source inventory: one row per Internet Archive item, with
-- normalized show fields and track titles for calendar, date and search views.
-- Raw metadata snapshots stay on disk; only their hashes are recorded here.
CREATE TABLE IF NOT EXISTS source_items (
  id text PRIMARY KEY,
  show_date date,
  date_text text NOT NULL DEFAULT '',
  artist text NOT NULL DEFAULT '',
  venue text NOT NULL DEFAULT '',
  location text NOT NULL DEFAULT '',
  title text NOT NULL DEFAULT '',
  kind text NOT NULL DEFAULT 'unknown',
  source text NOT NULL DEFAULT '',
  lineage text NOT NULL DEFAULT '',
  taper text NOT NULL DEFAULT '',
  transferer text NOT NULL DEFAULT '',
  downloads integer NOT NULL DEFAULT 0,
  avg_rating real NOT NULL DEFAULT 0,
  num_reviews integer NOT NULL DEFAULT 0,
  item_size bigint NOT NULL DEFAULT 0,
  added_at timestamptz,
  tracks jsonb NOT NULL DEFAULT '[]',
  songs text NOT NULL DEFAULT '',
  duration double precision NOT NULL DEFAULT 0,
  warnings jsonb NOT NULL DEFAULT '[]',
  summary jsonb NOT NULL DEFAULT '{}',
  snapshot_sha256 text NOT NULL DEFAULT '',
  observed_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  search tsvector GENERATED ALWAYS AS (
    to_tsvector('simple', title || ' ' || venue || ' ' || location || ' ' || date_text || ' ' || taper || ' ' || transferer || ' ' || source || ' ' || songs)
  ) STORED
);
CREATE INDEX IF NOT EXISTS source_items_show_date ON source_items (show_date);
CREATE INDEX IF NOT EXISTS source_items_month_day ON source_items ((EXTRACT(MONTH FROM show_date)), (EXTRACT(DAY FROM show_date)));
CREATE INDEX IF NOT EXISTS source_items_search ON source_items USING GIN (search);
