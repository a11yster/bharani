package replication

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

// Table stores the mapping from volumes to OSDs
type Table struct {
	db *sql.DB
	mu sync.RWMutex
}

// VolumeInfo represents volume information
type VolumeInfo struct {
	VolumeID    string
	OSDAddresses []string
	Generation  int64
	CellID      string
	State       string
}

// NewTable creates a new replication table
func NewTable(dbPath string) (*Table, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	table := &Table{db: db}

	if err := table.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return table, nil
}

// initSchema creates the database schema
func (t *Table) initSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS volumes (
		volume_id TEXT PRIMARY KEY,
		cell_id TEXT NOT NULL,
		generation INTEGER NOT NULL DEFAULT 1,
		state TEXT NOT NULL DEFAULT 'open',
		created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
		updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
	);
	
	CREATE TABLE IF NOT EXISTS volume_osds (
		volume_id TEXT NOT NULL,
		osd_address TEXT NOT NULL,
		PRIMARY KEY (volume_id, osd_address),
		FOREIGN KEY (volume_id) REFERENCES volumes(volume_id) ON DELETE CASCADE
	);
	
	CREATE INDEX IF NOT EXISTS idx_cell ON volumes(cell_id);
	`

	_, err := t.db.Exec(query)
	return err
}

// CreateVolume creates a new volume with OSD addresses
func (t *Table) CreateVolume(volumeID string, osdAddresses []string, cellID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	tx, err := t.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO volumes (volume_id, cell_id, generation, state)
		VALUES (?, ?, 1, 'open')
	`, volumeID, cellID)
	if err != nil {
		return fmt.Errorf("failed to create volume: %w", err)
	}

	for _, osdAddr := range osdAddresses {
		_, err = tx.Exec(`
			INSERT INTO volume_osds (volume_id, osd_address)
			VALUES (?, ?)
		`, volumeID, osdAddr)
		if err != nil {
			return fmt.Errorf("failed to add OSD address: %w", err)
		}
	}

	return tx.Commit()
}

// GetVolume retrieves volume information
func (t *Table) GetVolume(volumeID string) (*VolumeInfo, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var info VolumeInfo
	var cellID, state string
	var generation int64

	err := t.db.QueryRow(`
		SELECT volume_id, cell_id, generation, state
		FROM volumes
		WHERE volume_id = ?
	`, volumeID).Scan(&info.VolumeID, &cellID, &generation, &state)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get volume: %w", err)
	}

	info.CellID = cellID
	info.Generation = generation
	info.State = state

	rows, err := t.db.Query(`
		SELECT osd_address
		FROM volume_osds
		WHERE volume_id = ?
	`, volumeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get OSD addresses: %w", err)
	}
	defer rows.Close()

	osdAddresses := make([]string, 0)
	for rows.Next() {
		var addr string
		if err := rows.Scan(&addr); err != nil {
			return nil, fmt.Errorf("failed to scan OSD address: %w", err)
		}
		osdAddresses = append(osdAddresses, addr)
	}

	info.OSDAddresses = osdAddresses
	return &info, nil
}

// UpdateVolume updates volume information
func (t *Table) UpdateVolume(volumeID string, osdAddresses []string, generation int64, state string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	tx, err := t.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if state != "" {
		_, err = tx.Exec(`
			UPDATE volumes
			SET generation = ?, state = ?, updated_at = strftime('%s', 'now')
			WHERE volume_id = ?
		`, generation, state, volumeID)
	} else {
		_, err = tx.Exec(`
			UPDATE volumes
			SET generation = ?, updated_at = strftime('%s', 'now')
			WHERE volume_id = ?
		`, generation, volumeID)
	}
	if err != nil {
		return fmt.Errorf("failed to update volume: %w", err)
	}

	_, err = tx.Exec(`DELETE FROM volume_osds WHERE volume_id = ?`, volumeID)
	if err != nil {
		return fmt.Errorf("failed to delete old OSD addresses: %w", err)
	}

	for _, osdAddr := range osdAddresses {
		_, err = tx.Exec(`
			INSERT INTO volume_osds (volume_id, osd_address)
			VALUES (?, ?)
		`, volumeID, osdAddr)
		if err != nil {
			return fmt.Errorf("failed to add OSD address: %w", err)
		}
	}

	return tx.Commit()
}

// ListVolumes returns all volume IDs, optionally filtered by cell
func (t *Table) ListVolumes(cellID string) ([]string, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var rows *sql.Rows
	var err error

	if cellID != "" {
		rows, err = t.db.Query(`
			SELECT volume_id
			FROM volumes
			WHERE cell_id = ?
		`, cellID)
	} else {
		rows, err = t.db.Query(`SELECT volume_id FROM volumes`)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list volumes: %w", err)
	}
	defer rows.Close()

	volumeIDs := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan volume ID: %w", err)
		}
		volumeIDs = append(volumeIDs, id)
	}

	return volumeIDs, nil
}

// Close closes the database connection
func (t *Table) Close() error {
	return t.db.Close()
}


