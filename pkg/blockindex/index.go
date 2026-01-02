package blockindex

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

// Index stores the mapping from block hash to storage location
type Index struct {
	db *sql.DB
	mu sync.RWMutex
}

// Entry represents a block index entry
type Entry struct {
	Hash     string
	CellID   string
	BucketID string
	Checksum string
}

// NewIndex creates a new block index
func NewIndex(dbPath string) (*Index, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	index := &Index{db: db}

	if err := index.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return index, nil
}

// initSchema creates the database schema
func (i *Index) initSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS blocks (
		hash TEXT PRIMARY KEY,
		cell_id TEXT NOT NULL,
		bucket_id TEXT NOT NULL,
		checksum TEXT NOT NULL,
		created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
	);
	
	CREATE INDEX IF NOT EXISTS idx_cell_bucket ON blocks(cell_id, bucket_id);
	`

	_, err := i.db.Exec(query)
	return err
}

// PutEntry adds or updates a block index entry
func (i *Index) PutEntry(entry *Entry) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	query := `
	INSERT OR REPLACE INTO blocks (hash, cell_id, bucket_id, checksum)
	VALUES (?, ?, ?, ?)
	`

	_, err := i.db.Exec(query, entry.Hash, entry.CellID, entry.BucketID, entry.Checksum)
	if err != nil {
		return fmt.Errorf("failed to put entry: %w", err)
	}

	return nil
}

// GetEntry retrieves a block index entry by hash
func (i *Index) GetEntry(hash string) (*Entry, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	query := `
	SELECT hash, cell_id, bucket_id, checksum
	FROM blocks
	WHERE hash = ?
	`

	var entry Entry
	err := i.db.QueryRow(query, hash).Scan(
		&entry.Hash,
		&entry.CellID,
		&entry.BucketID,
		&entry.Checksum,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get entry: %w", err)
	}

	return &entry, nil
}

// Exists checks if a block exists in the index
func (i *Index) Exists(hash string) (bool, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	query := `SELECT 1 FROM blocks WHERE hash = ? LIMIT 1`
	var exists int
	err := i.db.QueryRow(query, hash).Scan(&exists)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}

	return true, nil
}

// Close closes the database connection
func (i *Index) Close() error {
	return i.db.Close()
}

