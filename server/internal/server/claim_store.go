package server

import (
	"database/sql"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

// ClaimStore is an in-memory store for IP address claims
// It can optionally use SQLite as a backend store
type ClaimStore struct {
	mutex    sync.RWMutex
	claims   map[string]string // map[ipAddress]claimantName
	ipTree   *IPTree           // Hierarchical tree for subnet-based queries
	db       *sql.DB           // Optional SQLite database for persistence
	dbPath   string            // Path to SQLite database file
}

// Verify ClaimStore implements Store interface
var _ Store = (*ClaimStore)(nil)

// NewClaimStore creates a new in-memory claim store without SQLite
func NewClaimStore() *ClaimStore {
	return &ClaimStore{
		claims: make(map[string]string),
		ipTree: NewIPTree(),
	}
}

// NewClaimStoreWithSQLite creates a claim store with SQLite backend
func NewClaimStoreWithSQLite(dbPath string) (*ClaimStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	store := &ClaimStore{
		claims: make(map[string]string),
		ipTree: NewIPTree(),
		db:     db,
		dbPath: dbPath,
	}

	// Initialize database schema
	if err := store.initSchema(); err != nil {
		return nil, err
	}

	// Load existing claims from SQLite
	if err := store.loadFromSQLite(); err != nil {
		return nil, err
	}

	return store, nil
}

// initSchema creates the database schema if it doesn't exist
func (cs *ClaimStore) initSchema() error {
	schema := `
		CREATE TABLE IF NOT EXISTS claims (
			ip_address TEXT PRIMARY KEY,
			claimant TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_claimant ON claims(claimant);
	`
	_, err := cs.db.Exec(schema)
	return err
}

// loadFromSQLite loads all claims from SQLite into memory
func (cs *ClaimStore) loadFromSQLite() error {
	rows, err := cs.db.Query("SELECT ip_address, claimant FROM claims")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var ipAddr, claimant string
		if err := rows.Scan(&ipAddr, &claimant); err != nil {
			return err
		}

		// Store in memory
		cs.claims[ipAddr] = claimant
		// Update the tree
		cs.ipTree.processClaim(ipAddr, claimant, "")
	}

	return rows.Err()
}

// ProcessClaim processes a claim request and updates the store
// Note: Updated to overwrite existing claims as per new requirements
func (cs *ClaimStore) ProcessClaim(ipAddr string, claimant string) error {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	// Get existing claimant if any
	oldClaimant, exists := cs.claims[ipAddr]

	// Store new claim in memory
	cs.claims[ipAddr] = claimant

	// If SQLite is enabled, write through to SQLite
	if cs.db != nil {
		var err error
		if exists {
			// Update existing claim
			_, err = cs.db.Exec(
				"UPDATE claims SET claimant = ?, updated_at = CURRENT_TIMESTAMP WHERE ip_address = ?",
				claimant, ipAddr,
			)
		} else {
			// Insert new claim
			_, err = cs.db.Exec(
				"INSERT INTO claims (ip_address, claimant) VALUES (?, ?)",
				ipAddr, claimant,
			)
		}

		if err != nil {
			// If SQLite fails, revert the in-memory change and propagate error
			if exists {
				cs.claims[ipAddr] = oldClaimant
			} else {
				delete(cs.claims, ipAddr)
			}
			return err
		}
	}

	// Update tree with hierarchical information
	if exists {
		// We're updating an existing claim
		cs.ipTree.processClaim(ipAddr, claimant, oldClaimant)
	} else {
		// New claim
		cs.ipTree.processClaim(ipAddr, claimant, "")
	}

	return nil
}

// GetClaim retrieves the claimant for an IP address
func (cs *ClaimStore) GetClaim(ipAddr string) (string, bool) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	claimant, exists := cs.claims[ipAddr]
	return claimant, exists
}

// GetSubnetStats retrieves statistics for a specific subnet
func (cs *ClaimStore) GetSubnetStats(subnet string) (*SubnetStats, bool) {
	return cs.ipTree.GetSubnetStats(subnet)
}

// GetAllClaims returns all claims in the store
func (cs *ClaimStore) GetAllClaims() map[string]string {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	// Create a copy to avoid concurrent access issues
	claims := make(map[string]string)
	for ip, claimant := range cs.claims {
		claims[ip] = claimant
	}

	return claims
}

// Close releases any resources held by the store
func (cs *ClaimStore) Close() error {
	if cs.db != nil {
		return cs.db.Close()
	}
	return nil
}
