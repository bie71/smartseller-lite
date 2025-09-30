package service

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
	"unicode"

	"smartseller-lite-starter/internal/db"
	"smartseller-lite-starter/internal/domain"
	"smartseller-lite-starter/internal/media"
)

// BackupService orchestrates SQL-based backup and restore flows for MySQL/MariaDB.
type BackupService struct {
	store *db.Store
	media *media.Manager
}

type sqlExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

// NewBackupService constructs a BackupService backed by the given store.
func NewBackupService(store *db.Store, mediaManager *media.Manager) *BackupService {
	return &BackupService{store: store, media: mediaManager}
}

// Create builds a base64 encoded SQL dump according to the provided options.
func (s *BackupService) Create(ctx context.Context, opts domain.BackupOptions) (string, error) {
	opts = normaliseBackupOptions(opts)
	sqlDump, err := s.dumpWithClient(ctx, opts)
	if err != nil {
		fallback, fbErr := s.dumpWithDriver(ctx, opts)
		if fbErr != nil {
			return "", fmt.Errorf("mysqldump fallback failed: %v (primary error: %w)", fbErr, err)
		}
		sqlDump = fallback
	}

	payload, err := s.wrapArchive(sqlDump, opts.IncludeMedia)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(payload), nil
}

// Restore executes the provided SQL payload and returns a summary of the operation.
func (s *BackupService) Restore(ctx context.Context, payload string, opts domain.RestoreOptions) (domain.RestoreResult, error) {
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(payload))
	if err != nil {
		return domain.RestoreResult{}, fmt.Errorf("decode payload: %w", err)
	}
	if len(raw) == 0 {
		return domain.RestoreResult{}, errors.New("backup payload is empty")
	}

	if isZipArchive(raw) {
		if s.media != nil {
			dump, err := s.media.RestoreArchive(raw)
			if err != nil {
				return domain.RestoreResult{}, err
			}
			raw = dump
		} else {
			dump, err := extractDumpFromArchive(raw)
			if err != nil {
				return domain.RestoreResult{}, err
			}
			raw = dump
		}
	}

	start := time.Now()
	if err := s.restoreWithClient(ctx, raw, opts); err == nil {
		duration := time.Since(start).Milliseconds()
		return domain.RestoreResult{Statements: 0, DurationMillis: duration, ExecutionDriver: "mysql"}, nil
	}

	statements, fbErr := s.restoreWithDriver(ctx, raw, opts)
	if fbErr != nil {
		return domain.RestoreResult{}, fbErr
	}
	duration := time.Since(start).Milliseconds()
	return domain.RestoreResult{Statements: statements, DurationMillis: duration, ExecutionDriver: "driver"}, nil
}

func normaliseBackupOptions(opts domain.BackupOptions) domain.BackupOptions {
	if !opts.IncludeSchema && !opts.IncludeData {
		opts.IncludeSchema = true
		opts.IncludeData = true
	}
	if !opts.IncludeMedia {
		opts.IncludeMedia = true
	}
	return opts
}

func (s *BackupService) dumpWithClient(ctx context.Context, opts domain.BackupOptions) ([]byte, error) {
	path, err := exec.LookPath("mysqldump")
	if err != nil {
		return nil, err
	}

	cfg := s.store.Config()
	args := make([]string, 0, 16)
	if cfg.Net == "unix" {
		args = append(args, fmt.Sprintf("--socket=%s", cfg.Addr))
	} else {
		host, port, _ := net.SplitHostPort(cfg.Addr)
		if host == "" {
			host = cfg.Addr
		}
		if host != "" {
			args = append(args, fmt.Sprintf("--host=%s", host))
		}
		if port != "" {
			args = append(args, fmt.Sprintf("--port=%s", port))
		}
		args = append(args, "--protocol=tcp")
	}
	if cfg.User != "" {
		args = append(args, fmt.Sprintf("--user=%s", cfg.User))
	}
	args = append(args,
		"--skip-lock-tables",
		"--single-transaction",
		"--default-character-set=utf8mb4",
		"--set-gtid-purged=OFF",
		"--column-statistics=0",
	)
	if !opts.IncludeSchema {
		args = append(args, "--no-create-info")
	}
	if !opts.IncludeData {
		args = append(args, "--no-data")
	}
	args = append(args, cfg.DBName)

	cmd := exec.CommandContext(ctx, path, args...)
	if cfg.Passwd != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("MYSQL_PWD=%s", cfg.Passwd))
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return nil, fmt.Errorf("mysqldump: %w: %s", err, strings.TrimSpace(stderr.String()))
		}
		return nil, fmt.Errorf("mysqldump: %w", err)
	}
	return stdout.Bytes(), nil
}

func (s *BackupService) dumpWithDriver(ctx context.Context, opts domain.BackupOptions) ([]byte, error) {
	dbConn := s.store.DB()
	cfg := s.store.Config()

	tables, err := s.listTables(ctx)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	now := time.Now().UTC().Format(time.RFC3339)
	buf.WriteString("-- SmartSeller Lite SQL Backup\n")
	buf.WriteString(fmt.Sprintf("-- Generated at %s UTC\n\n", now))
	buf.WriteString("SET NAMES utf8mb4;\n")
	buf.WriteString("SET time_zone = '+00:00';\n\n")
	if cfg.DBName != "" {
		buf.WriteString(fmt.Sprintf("USE `%s`;\n\n", cfg.DBName))
	}

	if opts.IncludeSchema {
		for _, table := range tables {
			createStmt, err := s.showCreateTable(ctx, table)
			if err != nil {
				return nil, err
			}
			buf.WriteString(fmt.Sprintf("DROP TABLE IF EXISTS `%s`;\n", table))
			buf.WriteString(createStmt)
			if !strings.HasSuffix(createStmt, ";\n") {
				buf.WriteString(";\n")
			}
			buf.WriteString("\n")
		}
	}

	if opts.IncludeData {
		buf.WriteString("SET FOREIGN_KEY_CHECKS=0;\n")
		for _, table := range tables {
			insertSQL, err := s.dumpTableRows(ctx, dbConn, table)
			if err != nil {
				return nil, err
			}
			buf.Write(insertSQL)
		}
		buf.WriteString("SET FOREIGN_KEY_CHECKS=1;\n")
	}

	return buf.Bytes(), nil
}

func (s *BackupService) listTables(ctx context.Context) ([]string, error) {
	cfg := s.store.Config()
	query := fmt.Sprintf("SHOW FULL TABLES IN `%s` WHERE Table_type = 'BASE TABLE';", cfg.DBName)
	rows, err := s.store.DB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name, tableType string
		if err := rows.Scan(&name, &tableType); err != nil {
			return nil, fmt.Errorf("scan table name: %w", err)
		}
		tables = append(tables, name)
	}
	sort.Strings(tables)
	return tables, rows.Err()
}

func (s *BackupService) wrapArchive(sqlDump []byte, includeMedia bool) ([]byte, error) {
	if s.media != nil {
		return s.media.CreateArchive("dump.sql", sqlDump, includeMedia)
	}

	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	file, err := writer.Create("dump.sql")
	if err != nil {
		return nil, err
	}
	if _, err := file.Write(sqlDump); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func extractDumpFromArchive(payload []byte) ([]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
	if err != nil {
		return nil, err
	}
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		name := strings.ToLower(file.Name)
		if strings.HasSuffix(name, ".sql") {
			rc, err := file.Open()
			if err != nil {
				return nil, err
			}
			data, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, err
			}
			if len(data) == 0 {
				continue
			}
			return data, nil
		}
	}
	return nil, errors.New("archive missing SQL dump")
}

func isZipArchive(data []byte) bool {
	return len(data) > 3 && data[0] == 'P' && data[1] == 'K'
}

func (s *BackupService) showCreateTable(ctx context.Context, table string) (string, error) {
	query := fmt.Sprintf("SHOW CREATE TABLE `%s`", table)
	row := s.store.DB().QueryRowContext(ctx, query)
	var name, createSQL string
	if err := row.Scan(&name, &createSQL); err != nil {
		return "", fmt.Errorf("show create table %s: %w", table, err)
	}
	if !strings.HasSuffix(createSQL, ";") {
		createSQL += ";"
	}
	return createSQL + "\n", nil
}

func (s *BackupService) dumpTableRows(ctx context.Context, dbConn *sql.DB, table string) ([]byte, error) {
	rows, err := dbConn.QueryContext(ctx, fmt.Sprintf("SELECT * FROM `%s`", table))
	if err != nil {
		return nil, fmt.Errorf("select %s: %w", table, err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("columns %s: %w", table, err)
	}
	if len(columns) == 0 {
		return nil, nil
	}
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("column types %s: %w", table, err)
	}

	var buf bytes.Buffer
	placeholders := make([]string, len(columns))
	for i, col := range columns {
		placeholders[i] = fmt.Sprintf("`%s`", col)
	}
	columnList := strings.Join(placeholders, ", ")

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]any, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("scan %s row: %w", table, err)
		}
		buf.WriteString(fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (", table, columnList))
		for i, raw := range values {
			if i > 0 {
				buf.WriteString(", ")
			}
			if raw == nil {
				buf.WriteString("NULL")
				continue
			}
			dbType := ""
			if i < len(colTypes) && colTypes[i] != nil {
				dbType = strings.ToUpper(colTypes[i].DatabaseTypeName())
			}
			switch {
			case isNumericType(dbType):
				buf.WriteString(string(raw))
			default:
				buf.WriteByte('\'')
				buf.WriteString(escapeSQLString(string(raw)))
				buf.WriteByte('\'')
			}
		}
		buf.WriteString(");\n")
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate %s rows: %w", table, err)
	}
	if buf.Len() > 0 {
		buf.WriteByte('\n')
	}
	return buf.Bytes(), nil
}

func isNumericType(dbType string) bool {
	switch strings.ToUpper(dbType) {
	case "INT", "INTEGER", "BIGINT", "SMALLINT", "TINYINT", "MEDIUMINT", "DECIMAL", "DOUBLE", "FLOAT", "REAL", "NUMERIC", "BIT":
		return true
	}
	return false
}

func escapeSQLString(value string) string {
	replaced := strings.ReplaceAll(value, "\\", "\\\\")
	replaced = strings.ReplaceAll(replaced, "'", "''")
	return replaced
}

func (s *BackupService) restoreWithClient(ctx context.Context, payload []byte, opts domain.RestoreOptions) error {
	path, err := exec.LookPath("mysql")
	if err != nil {
		return err
	}
	cfg := s.store.Config()

	args := make([]string, 0, 12)
	if cfg.Net == "unix" {
		args = append(args, fmt.Sprintf("--socket=%s", cfg.Addr))
	} else {
		host, port, _ := net.SplitHostPort(cfg.Addr)
		if host == "" {
			host = cfg.Addr
		}
		if host != "" {
			args = append(args, fmt.Sprintf("--host=%s", host))
		}
		if port != "" {
			args = append(args, fmt.Sprintf("--port=%s", port))
		}
		args = append(args, "--protocol=tcp")
	}
	if cfg.User != "" {
		args = append(args, fmt.Sprintf("--user=%s", cfg.User))
	}
	args = append(args, "--default-character-set=utf8mb4", cfg.DBName)

	var input bytes.Buffer
	if opts.DisableForeignKeyChecks {
		input.WriteString("SET FOREIGN_KEY_CHECKS=0;\n")
	}
	if opts.UseTransaction {
		input.WriteString("START TRANSACTION;\n")
	}
	input.Write(payload)
	input.WriteByte('\n')
	if opts.UseTransaction {
		input.WriteString("COMMIT;\n")
	}
	if opts.DisableForeignKeyChecks {
		input.WriteString("SET FOREIGN_KEY_CHECKS=1;\n")
	}

	cmd := exec.CommandContext(ctx, path, args...)
	if cfg.Passwd != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("MYSQL_PWD=%s", cfg.Passwd))
	}
	cmd.Stdin = &input
	cmd.Stdout = io.Discard
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return fmt.Errorf("mysql restore: %w: %s", err, strings.TrimSpace(stderr.String()))
		}
		return fmt.Errorf("mysql restore: %w", err)
	}
	return nil
}

func (s *BackupService) restoreWithDriver(ctx context.Context, payload []byte, opts domain.RestoreOptions) (int, error) {
	statements, err := splitSQLStatements(string(payload))
	if err != nil {
		return 0, err
	}
	dbConn := s.store.DB()
	var execTarget sqlExecutor = dbConn
	var tx *sql.Tx
	if opts.UseTransaction {
		tx, err = dbConn.BeginTx(ctx, nil)
		if err != nil {
			return 0, fmt.Errorf("begin restore transaction: %w", err)
		}
		execTarget = tx
	}

	fkExecutor := execTarget
	if opts.DisableForeignKeyChecks {
		if _, err := fkExecutor.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS=0"); err != nil {
			if tx != nil {
				_ = tx.Rollback()
			}
			return 0, fmt.Errorf("disable fk checks: %w", err)
		}
	}

	executed := 0
	for _, stmt := range statements {
		trimmed := strings.TrimSpace(stmt)
		if trimmed == "" || shouldSkipStatement(trimmed) {
			continue
		}
		if _, err := execTarget.ExecContext(ctx, trimmed); err != nil {
			if opts.DisableForeignKeyChecks {
				_, _ = fkExecutor.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS=1")
			}
			if tx != nil {
				_ = tx.Rollback()
			}
			return executed, fmt.Errorf("execute statement %d: %w", executed+1, err)
		}
		executed++
	}

	if opts.DisableForeignKeyChecks {
		if _, err := fkExecutor.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS=1"); err != nil {
			if tx != nil {
				_ = tx.Rollback()
			}
			return executed, fmt.Errorf("enable fk checks: %w", err)
		}
	}
	if tx != nil {
		if err := tx.Commit(); err != nil {
			return executed, fmt.Errorf("commit restore: %w", err)
		}
	}
	return executed, nil
}

func splitSQLStatements(input string) ([]string, error) {
	cleaned := strings.ReplaceAll(input, "\r\n", "\n")
	cleaned = strings.TrimLeftFunc(cleaned, func(r rune) bool { return r == '\ufeff' || unicode.IsSpace(r) })
	cleaned = expandVersionedComments(cleaned)

	var statements []string
	var buf strings.Builder
	inSingle := false
	inDouble := false
	inBacktick := false
	inLineComment := false
	inBlockComment := false

	for i := 0; i < len(cleaned); i++ {
		ch := cleaned[i]
		next := byte(0)
		if i+1 < len(cleaned) {
			next = cleaned[i+1]
		}

		if inLineComment {
			if ch == '\n' {
				inLineComment = false
				buf.WriteByte(ch)
			}
			continue
		}
		if inBlockComment {
			if ch == '*' && next == '/' {
				inBlockComment = false
				i++
			}
			continue
		}

		switch ch {
		case '-':
			if !inSingle && !inDouble && !inBacktick && next == '-' {
				after := byte(' ')
				if i+2 < len(cleaned) {
					after = cleaned[i+2]
				}
				if after == ' ' || after == '\t' || after == '\n' || after == '\r' {
					inLineComment = true
					i++
					continue
				}
			}
		case '#':
			if !inSingle && !inDouble && !inBacktick {
				inLineComment = true
				continue
			}
		case '/':
			if !inSingle && !inDouble && !inBacktick && next == '*' {
				inBlockComment = true
				i++
				continue
			}
		case '\n':
			if !inSingle && !inDouble && !inBacktick {
				buf.WriteByte(ch)
				continue
			}
		case '\'':
			if !inDouble && !inBacktick {
				buf.WriteByte(ch)
				if inSingle {
					if next == '\'' {
						buf.WriteByte(next)
						i++
						continue
					}
					inSingle = false
				} else {
					inSingle = true
				}
				continue
			}
		case '"':
			if !inSingle && !inBacktick {
				buf.WriteByte(ch)
				if inDouble {
					if next == '"' {
						buf.WriteByte(next)
						i++
						continue
					}
					inDouble = false
				} else {
					inDouble = true
				}
				continue
			}
		case '`':
			if !inSingle && !inDouble {
				buf.WriteByte(ch)
				inBacktick = !inBacktick
				continue
			}
		case ';':
			if !inSingle && !inDouble && !inBacktick {
				statement := strings.TrimSpace(buf.String())
				if statement != "" {
					statements = append(statements, statement)
				}
				buf.Reset()
				continue
			}
		}

		buf.WriteByte(ch)
	}

	if tail := strings.TrimSpace(buf.String()); tail != "" {
		statements = append(statements, tail)
	}
	return statements, nil
}

func expandVersionedComments(input string) string {
	var out strings.Builder
	for i := 0; i < len(input); i++ {
		if i+2 < len(input) && input[i] == '/' && input[i+1] == '*' && input[i+2] == '!' {
			i += 3
			for i < len(input) && input[i] >= '0' && input[i] <= '9' {
				i++
			}
			if i < len(input) && input[i] == ' ' {
				i++
			}
			for i+1 < len(input) && !(input[i] == '*' && input[i+1] == '/') {
				out.WriteByte(input[i])
				i++
			}
			i++
			continue
		}
		out.WriteByte(input[i])
	}
	return out.String()
}

func shouldSkipStatement(stmt string) bool {
	upper := strings.ToUpper(strings.TrimSpace(stmt))
	switch {
	case upper == "", upper == "COMMIT", upper == "ROLLBACK":
		return true
	case strings.HasPrefix(upper, "DELIMITER "):
		return true
	case strings.HasPrefix(upper, "LOCK TABLES"):
		return true
	case strings.HasPrefix(upper, "UNLOCK TABLES"):
		return true
	case strings.HasPrefix(upper, "USE "):
		return true
	}
	return false
}
