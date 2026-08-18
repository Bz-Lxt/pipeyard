package store

const schemaSQL = `
PRAGMA journal_mode=WAL;
PRAGMA foreign_keys=ON;
PRAGMA busy_timeout=5000;

CREATE TABLE IF NOT EXISTS jobs (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  status TEXT NOT NULL,
  spec_json TEXT NOT NULL,
  fingerprint TEXT NOT NULL,
  quota INTEGER NOT NULL DEFAULT 0,
  error TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS nodes (
  job_id TEXT NOT NULL,
  node_id TEXT NOT NULL,
  kind TEXT NOT NULL,
  param TEXT NOT NULL,
  weight INTEGER NOT NULL,
  status TEXT NOT NULL,
  lease_by TEXT NOT NULL DEFAULT '',
  lease_until TEXT NOT NULL DEFAULT '',
  error TEXT NOT NULL DEFAULT '',
  artifact_json TEXT NOT NULL DEFAULT '',
  attempt INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (job_id, node_id),
  FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS edges (
  job_id TEXT NOT NULL,
  src TEXT NOT NULL,
  dst TEXT NOT NULL,
  PRIMARY KEY (job_id, src, dst),
  FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS artifacts (
  digest TEXT PRIMARY KEY,
  job_id TEXT NOT NULL,
  node_id TEXT NOT NULL,
  kind TEXT NOT NULL,
  body TEXT NOT NULL,
  items_json TEXT NOT NULL,
  bytes INTEGER NOT NULL,
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS meta (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
`

const (
	MetaOpenedAt   = "opened_at"
	MetaWALApplied = "wal_applied"
	MetaCheckpoint = "checkpoint_at"
	MetaJobCount   = "job_count"
)
