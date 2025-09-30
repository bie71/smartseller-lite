package domain

// BackupOptions defines what should be included in a SQL backup export.
type BackupOptions struct {
	IncludeSchema bool `json:"includeSchema"`
	IncludeData   bool `json:"includeData"`
	IncludeMedia  bool `json:"includeMedia"`
}

// RestoreOptions control how a SQL restore operation is executed.
type RestoreOptions struct {
	DisableForeignKeyChecks bool `json:"disableForeignKeyChecks"`
	UseTransaction          bool `json:"useTransaction"`
}

// RestoreResult summarises the execution of a restore request.
type RestoreResult struct {
	Statements      int    `json:"statements"`
	DurationMillis  int64  `json:"durationMs"`
	ExecutionDriver string `json:"executionDriver"`
}
