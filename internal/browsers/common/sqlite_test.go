package common

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestOpenBrowserSQLiteReadsCommittedWALRow(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "history.sqlite")
	writer, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer func() {
		if err := writer.Close(); err != nil {
			t.Errorf("writer.Close() error = %v", err)
		}
	}()

	statements := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA wal_autocheckpoint=0",
		"CREATE TABLE visits (id INTEGER PRIMARY KEY)",
		"INSERT INTO visits(id) VALUES (1)",
	}
	for _, statement := range statements {
		if _, err := writer.Exec(statement); err != nil {
			t.Fatalf("writer.Exec(%q) error = %v", statement, err)
		}
	}

	reader, err := OpenBrowserSQLite(databasePath)
	if err != nil {
		t.Fatalf("OpenBrowserSQLite() error = %v", err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			t.Errorf("reader.Close() error = %v", err)
		}
	}()

	var count int
	if err := reader.QueryRow("SELECT COUNT(*) FROM visits").Scan(&count); err != nil {
		t.Fatalf("reader query error = %v", err)
	}
	if count != 1 {
		t.Errorf("reader count = %d, want 1 committed WAL row", count)
	}
}

func TestReadOnlySQLiteBusyTimeoutIsBounded(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "busy.sqlite")
	writer, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer func() {
		if err := writer.Close(); err != nil {
			t.Errorf("writer.Close() error = %v", err)
		}
	}()
	if _, err := writer.Exec("CREATE TABLE visits (id INTEGER PRIMARY KEY)"); err != nil {
		t.Fatalf("create fixture table error = %v", err)
	}

	connection, err := writer.Conn(context.Background())
	if err != nil {
		t.Fatalf("writer.Conn() error = %v", err)
	}
	defer func() {
		if err := connection.Close(); err != nil {
			t.Errorf("connection.Close() error = %v", err)
		}
	}()
	if _, err := connection.ExecContext(context.Background(), "BEGIN EXCLUSIVE"); err != nil {
		t.Fatalf("BEGIN EXCLUSIVE error = %v", err)
	}
	defer func() {
		if _, err := connection.ExecContext(context.Background(), "ROLLBACK"); err != nil {
			t.Errorf("ROLLBACK error = %v", err)
		}
	}()

	start := time.Now()
	reader, openErr := openReadOnlySQLite(databasePath, 50*time.Millisecond)
	if openErr == nil {
		defer func() {
			if err := reader.Close(); err != nil {
				t.Errorf("reader.Close() error = %v", err)
			}
		}()
		var count int
		openErr = reader.QueryRow("SELECT COUNT(*) FROM visits").Scan(&count)
	}
	if openErr == nil {
		t.Fatal("read from exclusively locked database unexpectedly succeeded")
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("busy read took %v, want a bounded timeout below one second", elapsed)
	}
}

func TestOpenBrowserSQLiteFallsBackFromPersistentExclusiveLock(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "exclusive.sqlite")
	writer, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	writer.SetMaxOpenConns(1)
	defer func() {
		if err := writer.Close(); err != nil {
			t.Errorf("writer.Close() error = %v", err)
		}
	}()

	connection, err := writer.Conn(context.Background())
	if err != nil {
		t.Fatalf("writer.Conn() error = %v", err)
	}
	defer func() {
		if err := connection.Close(); err != nil {
			t.Errorf("connection.Close() error = %v", err)
		}
	}()

	statements := []string{
		"PRAGMA locking_mode=EXCLUSIVE",
		"CREATE TABLE visits (id INTEGER PRIMARY KEY)",
		"INSERT INTO visits(id) VALUES (1)",
	}
	for _, statement := range statements {
		if _, err := connection.ExecContext(context.Background(), statement); err != nil {
			t.Fatalf("writer.Exec(%q) error = %v", statement, err)
		}
	}

	directReader, directErr := openReadOnlySQLite(databasePath, 50*time.Millisecond)
	if directErr == nil {
		directReader.Close()
		t.Fatal("direct read through persistent exclusive lock unexpectedly succeeded")
	}
	if !isSQLiteLockError(directErr) {
		t.Fatalf("direct read error = %v, want SQLite lock error", directErr)
	}

	reader, err := openBrowserSQLite(databasePath, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("openBrowserSQLite() error = %v", err)
	}
	snapshotDirectory := reader.cleanupDirectory
	if snapshotDirectory == "" {
		t.Fatal("openBrowserSQLite() did not use a private snapshot")
	}

	var count int
	if err := reader.QueryRow("SELECT COUNT(*) FROM visits").Scan(&count); err != nil {
		t.Fatalf("snapshot query error = %v", err)
	}
	if count != 1 {
		t.Errorf("snapshot count = %d, want 1", count)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("reader.Close() error = %v", err)
	}
	if _, err := os.Stat(snapshotDirectory); !os.IsNotExist(err) {
		t.Errorf("snapshot directory still exists after Close(): %v", err)
	}
}

func TestSQLiteSnapshotIncludesCommittedWALRows(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "wal-snapshot.sqlite")
	writer, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer func() {
		if err := writer.Close(); err != nil {
			t.Errorf("writer.Close() error = %v", err)
		}
	}()

	statements := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA wal_autocheckpoint=0",
		"CREATE TABLE visits (id INTEGER PRIMARY KEY)",
		"INSERT INTO visits(id) VALUES (1)",
	}
	for _, statement := range statements {
		if _, err := writer.Exec(statement); err != nil {
			t.Fatalf("writer.Exec(%q) error = %v", statement, err)
		}
	}

	snapshotPath, snapshotDirectory, err := createSQLiteSnapshot(databasePath)
	if err != nil {
		t.Fatalf("createSQLiteSnapshot() error = %v", err)
	}
	defer os.RemoveAll(snapshotDirectory)

	reader, err := openReadOnlySQLite(snapshotPath, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("open snapshot error = %v", err)
	}
	defer reader.Close()

	var count int
	if err := reader.QueryRow("SELECT COUNT(*) FROM visits").Scan(&count); err != nil {
		t.Fatalf("snapshot query error = %v", err)
	}
	if count != 1 {
		t.Errorf("snapshot count = %d, want 1 committed WAL row", count)
	}
}

func TestSQLiteSnapshotRejectsHotRollbackJournalHeader(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "active-journal.sqlite")
	writer, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	if _, err := writer.Exec("CREATE TABLE visits (id INTEGER PRIMARY KEY)"); err != nil {
		t.Fatalf("create fixture table error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close() error = %v", err)
	}

	journalMagic := []byte{0xd9, 0xd5, 0x05, 0xf9, 0x20, 0xa1, 0x63, 0xd7}
	if err := os.WriteFile(databasePath+"-journal", journalMagic, 0o600); err != nil {
		t.Fatalf("write hot journal fixture error = %v", err)
	}

	_, _, snapshotErr := createSQLiteSnapshot(databasePath)
	if snapshotErr == nil || !strings.Contains(snapshotErr.Error(), "rollback journal is active") {
		t.Fatalf("createSQLiteSnapshot() error = %v, want active journal error", snapshotErr)
	}
}
