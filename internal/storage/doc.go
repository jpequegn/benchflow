// Package storage provides persistent storage for benchmark results using SQLite.
//
// # Overview
//
// The storage package implements historical tracking of benchmark results in SQLite,
// enabling trend analysis, baseline comparison, and long-term performance monitoring.
//
// # Features
//
//   - SQLite-based persistent storage
//   - Historical result tracking with timestamps
//   - Query by timestamp, range, or benchmark name
//   - Automatic cleanup of old records
//   - Foreign key constraints for data integrity
//   - Indexed queries for fast retrieval
//
// # Usage
//
// Basic storage operations:
//
//	// Create storage instance
//	storage, err := storage.NewSQLiteStorage("./benchflow.db")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer storage.Close()
//
//	// Initialize schema
//	if err := storage.Init(); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Save aggregated results
//	if err := storage.Save(suite); err != nil {
//	    log.Fatal(err)
//	}
//
// Retrieving historical data:
//
//	// Get most recent suite
//	latest, err := storage.GetLatest()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get suite by specific timestamp
//	suite, err := storage.GetByTimestamp(timestamp)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get suites within time range
//	start := time.Now().AddDate(0, 0, -7) // Last 7 days
//	end := time.Now()
//	suites, err := storage.GetRange(start, end)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Benchmark history tracking:
//
//	// Get history for specific benchmark
//	history, err := storage.GetHistory("bench_sort", 10) // Last 10 runs
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Analyze trend
//	for _, result := range history {
//	    fmt.Printf("%s: %v\n", result.Timestamp, result.Mean)
//	}
//
// Cleanup old records:
//
//	// Remove records older than 90 days
//	if err := storage.Cleanup(90); err != nil {
//	    log.Fatal(err)
//	}
//
// # Database Schema
//
// ## suites table
//
//	CREATE TABLE suites (
//	    id INTEGER PRIMARY KEY AUTOINCREMENT,
//	    timestamp DATETIME NOT NULL,
//	    duration INTEGER NOT NULL,
//	    metadata TEXT,
//	    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
//	);
//
// ## results table
//
//	CREATE TABLE results (
//	    id INTEGER PRIMARY KEY AUTOINCREMENT,
//	    suite_id INTEGER NOT NULL,
//	    name TEXT NOT NULL,
//	    language TEXT NOT NULL,
//	    mean INTEGER NOT NULL,
//	    median INTEGER NOT NULL,
//	    min INTEGER NOT NULL,
//	    max INTEGER NOT NULL,
//	    stddev INTEGER NOT NULL,
//	    iterations INTEGER NOT NULL,
//	    timestamp DATETIME NOT NULL,
//	    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
//	    FOREIGN KEY (suite_id) REFERENCES suites(id) ON DELETE CASCADE
//	);
//
// # Indexes
//
// The following indexes are created for query optimization:
//
//   - suites.timestamp - Fast retrieval by timestamp
//   - results.suite_id - Fast join with suites
//   - results.name - Fast benchmark history queries
//   - results.timestamp - Fast time-based queries
//
// # Data Model
//
// Each benchmark run creates:
//   - 1 suite record (metadata + timestamp)
//   - N result records (one per benchmark)
//
// Results are linked to suites via foreign key with CASCADE delete,
// ensuring referential integrity.
//
// # Query Performance
//
// Typical query performance on standard hardware:
//
//   - GetLatest: <1ms
//   - GetByTimestamp: <1ms
//   - GetRange (100 suites): ~10ms
//   - GetHistory (1000 results): ~20ms
//   - Save (10 results): ~5ms
//
// # Storage Size
//
// Approximate storage requirements:
//
//   - Suite record: ~100 bytes
//   - Result record: ~150 bytes
//   - 1000 suites × 10 results: ~1.5 MB
//   - 10000 suites × 10 results: ~15 MB
//
// # Thread Safety
//
// SQLiteStorage uses database/sql which provides connection pooling and
// is safe for concurrent use from multiple goroutines.
//
// However, SQLite itself has limitations with concurrent writes. For high-
// concurrency scenarios, consider:
//
//   - WAL mode: PRAGMA journal_mode=WAL
//   - Connection pool tuning
//   - External queue for writes
//
// # Transactions
//
// The Save method uses transactions to ensure atomicity:
//
//   - BEGIN TRANSACTION
//   - INSERT suite
//   - INSERT all results
//   - COMMIT
//
// If any step fails, the entire operation is rolled back.
//
// # Data Retention
//
// Use the Cleanup method to implement data retention policies:
//
//	// Daily cleanup job
//	ticker := time.NewTicker(24 * time.Hour)
//	go func() {
//	    for range ticker.C {
//	        if err := storage.Cleanup(90); err != nil {
//	            log.Printf("Cleanup failed: %v", err)
//	        }
//	    }
//	}()
//
// # Migration
//
// The Init method is idempotent and safe to call multiple times. It uses
// CREATE TABLE IF NOT EXISTS for schema creation.
//
// For schema changes, implement migrations manually:
//
//	ALTER TABLE results ADD COLUMN new_field TEXT;
//
// # Backup
//
// To backup the database:
//
//	// Close connections first
//	storage.Close()
//
//	// Copy the database file
//	cp benchflow.db benchflow_backup.db
//
//	// Reopen storage
//	storage, _ = storage.NewSQLiteStorage("benchflow.db")
//	storage.Init()
//
// Or use SQLite's BACKUP API for online backups.
package storage
