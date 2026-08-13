package common

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/mattn/go-sqlite3"
)

const (
	DefaultSQLiteBusyTimeout = 5 * time.Second
	snapshotAttempts         = 10
	snapshotRetryDelay       = 20 * time.Millisecond
)

var snapshotSuffixes = []string{"", "-wal", "-shm", "-journal"}

// BrowserSQLite is a browser database connection and any private snapshot it owns.
type BrowserSQLite struct {
	*sql.DB

	cleanupDirectory string
	cleanupOnce      sync.Once
	cleanupError     error
}

// Close closes the database and removes a private lock-fallback snapshot, if any.
func (database *BrowserSQLite) Close() error {
	closeError := database.DB.Close()
	database.cleanupOnce.Do(func() {
		if database.cleanupDirectory != "" {
			database.cleanupError = os.RemoveAll(database.cleanupDirectory)
		}
	})
	return errors.Join(closeError, database.cleanupError)
}

// OpenBrowserSQLite opens a live browser database without modifying the source files.
func OpenBrowserSQLite(databasePath string) (*BrowserSQLite, error) {
	return openBrowserSQLite(databasePath, DefaultSQLiteBusyTimeout)
}

func openBrowserSQLite(
	databasePath string,
	busyTimeout time.Duration,
) (*BrowserSQLite, error) {
	database, err := openReadOnlySQLite(databasePath, busyTimeout)
	if err == nil {
		return &BrowserSQLite{DB: database}, nil
	}
	if !isSQLiteLockError(err) {
		return nil, err
	}

	snapshotPath, snapshotDirectory, snapshotErr := createSQLiteSnapshot(databasePath)
	if snapshotErr != nil {
		return nil, fmt.Errorf(
			"open SQLite database %q: source is locked and snapshot failed: %w",
			databasePath,
			snapshotErr,
		)
	}

	snapshotDatabase, snapshotOpenErr := openReadOnlySQLite(snapshotPath, busyTimeout)
	if snapshotOpenErr != nil {
		cleanupErr := os.RemoveAll(snapshotDirectory)
		return nil, fmt.Errorf(
			"open SQLite database %q from private snapshot: %w",
			databasePath,
			errors.Join(snapshotOpenErr, cleanupErr),
		)
	}

	return &BrowserSQLite{
		DB:               snapshotDatabase,
		cleanupDirectory: snapshotDirectory,
	}, nil
}

func openReadOnlySQLite(databasePath string, busyTimeout time.Duration) (*sql.DB, error) {
	absolutePath, err := filepath.Abs(databasePath)
	if err != nil {
		return nil, fmt.Errorf("resolve SQLite path %q: %w", databasePath, err)
	}

	if busyTimeout < 0 {
		return nil, fmt.Errorf("open SQLite database %q: busy timeout must not be negative", absolutePath)
	}

	query := url.Values{}
	query.Set("mode", "ro")
	query.Set("_busy_timeout", strconv.FormatInt(busyTimeout.Milliseconds(), 10))
	dsn := (&url.URL{
		Scheme:   "file",
		Path:     filepath.ToSlash(absolutePath),
		RawQuery: query.Encode(),
	}).String()

	database, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open SQLite database %q: %w", absolutePath, err)
	}
	database.SetMaxOpenConns(1)

	if err := database.Ping(); err != nil {
		if closeErr := database.Close(); closeErr != nil {
			return nil, fmt.Errorf(
				"open SQLite database %q: %w (close after failure: %v)",
				absolutePath,
				err,
				closeErr,
			)
		}
		return nil, fmt.Errorf("open SQLite database %q: %w", absolutePath, err)
	}

	return database, nil
}

func isSQLiteLockError(err error) bool {
	var sqliteError sqlite3.Error
	if !errors.As(err, &sqliteError) {
		return false
	}
	return sqliteError.Code == sqlite3.ErrBusy || sqliteError.Code == sqlite3.ErrLocked
}

type snapshotFileState struct {
	exists bool
	info   os.FileInfo
}

// deliberate: guideline is 7 vars/branches; one function keeps each snapshot attempt atomic.
func createSQLiteSnapshot(databasePath string) (string, string, error) {
	absolutePath, err := filepath.Abs(databasePath)
	if err != nil {
		return "", "", fmt.Errorf("resolve snapshot source %q: %w", databasePath, err)
	}

	var lastReason string
	for attempt := 0; attempt < snapshotAttempts; attempt++ {
		before, err := readSnapshotFileStates(absolutePath)
		if err != nil {
			return "", "", err
		}
		journalActive, err := rollbackJournalActive(absolutePath+"-journal", before["-journal"])
		if err != nil {
			return "", "", err
		}
		if journalActive {
			lastReason = "rollback journal is active"
			time.Sleep(snapshotRetryDelay)
			continue
		}

		snapshotDirectory, err := os.MkdirTemp("", "browser-sqlite-snapshot-")
		if err != nil {
			return "", "", fmt.Errorf("create private SQLite snapshot directory: %w", err)
		}

		copyErr := copySQLiteSnapshotFiles(absolutePath, snapshotDirectory, before)
		after, stateErr := readSnapshotFileStates(absolutePath)
		stable := stateErr == nil && snapshotFileStatesEqual(before, after)
		if copyErr == nil && stable {
			return filepath.Join(snapshotDirectory, filepath.Base(absolutePath)),
				snapshotDirectory,
				nil
		}

		if cleanupErr := os.RemoveAll(snapshotDirectory); cleanupErr != nil {
			return "", "", fmt.Errorf("remove unstable SQLite snapshot: %w", cleanupErr)
		}
		if stateErr != nil {
			return "", "", stateErr
		}
		if copyErr != nil && stable {
			return "", "", copyErr
		}
		lastReason = "source files changed during copy"
		time.Sleep(snapshotRetryDelay)
	}

	return "", "", fmt.Errorf(
		"capture stable SQLite snapshot after %d attempts: %s",
		snapshotAttempts,
		lastReason,
	)
}

func rollbackJournalActive(path string, state snapshotFileState) (active bool, err error) {
	if !state.exists || state.info.Size() == 0 {
		return false, nil
	}

	journal, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("inspect SQLite rollback journal %q: %w", path, err)
	}
	defer func() {
		if closeErr := journal.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close SQLite rollback journal %q: %w", path, closeErr)
		}
	}()

	header := make([]byte, 8)
	bytesRead, readErr := io.ReadFull(journal, header)
	if readErr != nil &&
		!errors.Is(readErr, io.EOF) &&
		!errors.Is(readErr, io.ErrUnexpectedEOF) {
		return false, fmt.Errorf("read SQLite rollback journal %q: %w", path, readErr)
	}
	if bytesRead < len(header) {
		return true, nil
	}
	for _, value := range header {
		if value != 0 {
			return true, nil
		}
	}
	return false, nil
}

func readSnapshotFileStates(databasePath string) (map[string]snapshotFileState, error) {
	states := make(map[string]snapshotFileState, len(snapshotSuffixes))
	for _, suffix := range snapshotSuffixes {
		path := databasePath + suffix
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			states[suffix] = snapshotFileState{}
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("inspect SQLite snapshot source %q: %w", path, err)
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("inspect SQLite snapshot source %q: not a regular file", path)
		}
		states[suffix] = snapshotFileState{exists: true, info: info}
	}
	if !states[""].exists {
		return nil, fmt.Errorf("inspect SQLite snapshot source %q: file does not exist", databasePath)
	}
	return states, nil
}

func copySQLiteSnapshotFiles(
	databasePath string,
	snapshotDirectory string,
	states map[string]snapshotFileState,
) error {
	for _, suffix := range snapshotSuffixes {
		state := states[suffix]
		if !state.exists || suffix == "-journal" {
			continue
		}
		sourcePath := databasePath + suffix
		destinationPath := filepath.Join(
			snapshotDirectory,
			filepath.Base(databasePath)+suffix,
		)
		if err := copyPrivateFile(sourcePath, destinationPath); err != nil {
			return fmt.Errorf("copy SQLite snapshot file %q: %w", sourcePath, err)
		}
	}
	return nil
}

func copyPrivateFile(sourcePath, destinationPath string) (err error) {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := source.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	destination, err := os.OpenFile(
		destinationPath,
		os.O_CREATE|os.O_EXCL|os.O_WRONLY,
		0o600,
	)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := destination.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	if _, err := io.Copy(destination, source); err != nil {
		return err
	}
	return nil
}

func snapshotFileStatesEqual(
	first map[string]snapshotFileState,
	second map[string]snapshotFileState,
) bool {
	for _, suffix := range snapshotSuffixes {
		firstState := first[suffix]
		secondState := second[suffix]
		if firstState.exists != secondState.exists {
			return false
		}
		if !firstState.exists {
			continue
		}
		if !os.SameFile(firstState.info, secondState.info) ||
			firstState.info.Size() != secondState.info.Size() ||
			!firstState.info.ModTime().Equal(secondState.info.ModTime()) {
			return false
		}
	}
	return true
}
