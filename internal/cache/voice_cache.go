package cache

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/indaco/md2audio/internal/tts"
)

const (
	// DefaultCacheDir is the default directory for cache files
	DefaultCacheDir = ".md2audio"
	// DefaultCacheFile is the default SQLite database filename
	DefaultCacheFile = "voice_cache.db"
	// DefaultCacheDuration is how long cache entries are valid (30 days)
	// Voice lists from TTS providers don't change frequently, so a longer
	// cache duration reduces unnecessary API calls. Users can always use
	// -refresh-cache to get the latest voices when needed.
	DefaultCacheDuration = 30 * 24 * time.Hour
)

// VoiceCache provides caching for TTS provider voices using SQLite.
type VoiceCache struct {
	db            *sql.DB
	cacheDuration time.Duration
}

// NewVoiceCache creates a new voice cache with default settings.
// The cache is stored in ~/.md2audio/voice_cache.db by default.
func NewVoiceCache() (*VoiceCache, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	cacheDir := filepath.Join(homeDir, DefaultCacheDir)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	cachePath := filepath.Join(cacheDir, DefaultCacheFile)
	return NewVoiceCacheWithPath(cachePath, DefaultCacheDuration)
}

// NewVoiceCacheWithPath creates a new voice cache with a custom path.
func NewVoiceCacheWithPath(dbPath string, cacheDuration time.Duration) (*VoiceCache, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create table if not exists
	schema := `
	CREATE TABLE IF NOT EXISTS voices (
		provider TEXT NOT NULL,
		voice_id TEXT NOT NULL,
		name TEXT NOT NULL,
		description TEXT,
		language TEXT,
		gender TEXT,
		cached_at INTEGER NOT NULL,
		PRIMARY KEY (provider, voice_id)
	);
	CREATE INDEX IF NOT EXISTS idx_provider ON voices(provider);
	CREATE INDEX IF NOT EXISTS idx_cached_at ON voices(cached_at);
	`

	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return &VoiceCache{
		db:            db,
		cacheDuration: cacheDuration,
	}, nil
}

// Close closes the database connection.
func (c *VoiceCache) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// Get retrieves cached voices for a provider.
// Returns nil if cache is expired or doesn't exist.
func (c *VoiceCache) Get(ctx context.Context, provider string) ([]tts.Voice, error) {
	cutoff := time.Now().Add(-c.cacheDuration).Unix()

	query := `
	SELECT voice_id, name, description, language, gender
	FROM voices
	WHERE provider = ? AND cached_at > ?
	ORDER BY name
	`

	rows, err := c.db.QueryContext(ctx, query, provider, cutoff)
	if err != nil {
		return nil, fmt.Errorf("failed to query cache: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var voices []tts.Voice
	for rows.Next() {
		var v tts.Voice
		if err := rows.Scan(&v.ID, &v.Name, &v.Description, &v.Language, &v.Gender); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		voices = append(voices, v)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	// Return nil if no voices found (cache miss)
	if len(voices) == 0 {
		return nil, nil
	}

	return voices, nil
}

// Set stores voices for a provider in the cache.
func (c *VoiceCache) Set(ctx context.Context, provider string, voices []tts.Voice) error {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Delete old entries for this provider
	if _, err := tx.ExecContext(ctx, "DELETE FROM voices WHERE provider = ?", provider); err != nil {
		return fmt.Errorf("failed to delete old entries: %w", err)
	}

	// Insert new entries
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO voices (provider, voice_id, name, description, language, gender, cached_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	now := time.Now().Unix()
	for _, voice := range voices {
		if _, err := stmt.ExecContext(ctx, provider, voice.ID, voice.Name, voice.Description, voice.Language, voice.Gender, now); err != nil {
			return fmt.Errorf("failed to insert voice: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Clear removes all cached voices for a provider.
func (c *VoiceCache) Clear(ctx context.Context, provider string) error {
	if _, err := c.db.ExecContext(ctx, "DELETE FROM voices WHERE provider = ?", provider); err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}
	return nil
}

// ClearAll removes all cached voices for all providers.
func (c *VoiceCache) ClearAll(ctx context.Context) error {
	if _, err := c.db.ExecContext(ctx, "DELETE FROM voices"); err != nil {
		return fmt.Errorf("failed to clear all cache: %w", err)
	}
	return nil
}

// GetCacheInfo returns information about cached voices.
func (c *VoiceCache) GetCacheInfo(ctx context.Context, provider string) (*CacheInfo, error) {
	query := `
	SELECT COUNT(*), MIN(cached_at), MAX(cached_at)
	FROM voices
	WHERE provider = ?
	`

	var count int
	var minTime, maxTime sql.NullInt64

	err := c.db.QueryRowContext(ctx, query, provider).Scan(&count, &minTime, &maxTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get cache info: %w", err)
	}

	info := &CacheInfo{
		Provider: provider,
		Count:    count,
	}

	if minTime.Valid {
		info.OldestEntry = time.Unix(minTime.Int64, 0)
	}
	if maxTime.Valid {
		info.NewestEntry = time.Unix(maxTime.Int64, 0)
	}

	return info, nil
}

// CacheInfo contains information about cached voices.
type CacheInfo struct {
	Provider    string
	Count       int
	OldestEntry time.Time
	NewestEntry time.Time
}

// IsExpired returns true if the cache is expired.
func (i *CacheInfo) IsExpired(duration time.Duration) bool {
	if i.Count == 0 {
		return true
	}
	return time.Since(i.NewestEntry) > duration
}

// ExportToJSON exports cached voices to a JSON file.
func (c *VoiceCache) ExportToJSON(ctx context.Context, provider, outputPath string) error {
	voices, err := c.Get(ctx, provider)
	if err != nil {
		return fmt.Errorf("failed to get cached voices: %w", err)
	}

	if len(voices) == 0 {
		return fmt.Errorf("no cached voices found for provider: %s", provider)
	}

	data, err := json.MarshalIndent(voices, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	return nil
}
