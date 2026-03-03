package store

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS schema_version (version INTEGER NOT NULL);
INSERT INTO schema_version (version) SELECT 0 WHERE NOT EXISTS (SELECT 1 FROM schema_version);`,

	`CREATE TABLE IF NOT EXISTS tasks (
		id               TEXT PRIMARY KEY,
		name             TEXT NOT NULL,
		prompt           TEXT NOT NULL,
		schedule_type    TEXT NOT NULL,
		schedule_expr    TEXT NOT NULL,
		enabled          INTEGER NOT NULL DEFAULT 1,
		work_dir         TEXT NOT NULL DEFAULT '',
		skip_permissions INTEGER NOT NULL DEFAULT 1,
		created_at       INTEGER NOT NULL,
		updated_at       INTEGER NOT NULL,
		last_run_at      INTEGER,
		last_exit_code   INTEGER
	);

	CREATE TABLE IF NOT EXISTS run_logs (
		id         TEXT PRIMARY KEY,
		task_id    TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
		started_at INTEGER NOT NULL,
		ended_at   INTEGER,
		exit_code  INTEGER,
		output     TEXT NOT NULL DEFAULT '',
		triggered  TEXT NOT NULL DEFAULT 'scheduled'
	);

	CREATE INDEX IF NOT EXISTS idx_run_logs_task_id ON run_logs(task_id, started_at DESC);`,
}
