-- +goose Up
CREATE TABLE IF NOT EXISTS link_visits (
  id BIGSERIAL PRIMARY KEY,
  link_id BIGINT NOT NULL REFERENCES links(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  ip TEXT NOT NULL,
  user_agent TEXT NOT NULL,
  referer TEXT NOT NULL DEFAULT '',
  status INT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_link_visits_created_at
  ON link_visits (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_link_visits_link_id_created_at
  ON link_visits (link_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS link_visits;
