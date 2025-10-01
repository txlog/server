package database

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/txlog/server/models"
)

// GetAllAvailableMigrations returns all migration files from the embedded filesystem
func GetAllAvailableMigrations() ([]models.Migration, error) {
	var migrations []models.Migration

	// Read migration files from embedded filesystem
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	migrationMap := make(map[int]models.Migration)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if !strings.HasSuffix(filename, ".up.sql") {
			continue
		}

		// Parse migration version and description from filename
		// Format: YYYYMMDD_description.up.sql
		parts := strings.SplitN(filename, "_", 2)
		if len(parts) < 2 {
			continue
		}

		versionStr := parts[0]
		version, err := strconv.Atoi(versionStr)
		if err != nil {
			continue
		}

		descPart := strings.TrimSuffix(parts[1], ".up.sql")
		description := strings.ReplaceAll(descPart, "_", " ")
		description = toTitle(description)

		migrationMap[version] = models.Migration{
			Version:     version,
			Filename:    filename,
			Description: description,
			Applied:     false,
		}
	}

	// Convert map to sorted slice
	var versions []int
	for version := range migrationMap {
		versions = append(versions, version)
	}
	sort.Ints(versions)

	for _, version := range versions {
		migrations = append(migrations, migrationMap[version])
	}

	return migrations, nil
}

// RunAllMigrations applies all pending migrations using the same mechanism as ConnectDatabase
func RunAllMigrations() error {
	// Create postgres driver instance
	driver, err := postgres.WithInstance(Db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create database driver: %w", err)
	}

	// Create migration source from embedded filesystem
	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	// Create migration instance
	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}
	defer m.Close()

	// Apply all pending migrations
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

// toTitle converts a string to title case (first letter of each word capitalized)
// This is a simple replacement for the deprecated strings.Title function
func toTitle(s string) string {
	words := strings.Fields(strings.ToLower(s))
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}
