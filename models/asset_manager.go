package models

import (
	"database/sql"
	"time"

	logger "github.com/txlog/server/logger"
)

type AssetManager struct {
	db *sql.DB
}

func NewAssetManager(db *sql.DB) *AssetManager {
	return &AssetManager{db: db}
}

func (am *AssetManager) UpsertAsset(tx *sql.Tx, hostname string, machineID string, timestamp time.Time) error {
	var existingAssetID int
	var existingIsActive bool

	err := tx.QueryRow(`
		SELECT asset_id, is_active
		FROM assets
		WHERE hostname = $1 AND machine_id = $2
	`, hostname, machineID).Scan(&existingAssetID, &existingIsActive)

	if err == sql.ErrNoRows {
		err = am.deactivateOldAssets(tx, hostname, machineID)
		if err != nil {
			return err
		}

		_, err = tx.Exec(`
			INSERT INTO assets (hostname, machine_id, first_seen, last_seen, is_active, created_at)
			VALUES ($1, $2, $3, $3, TRUE, CURRENT_TIMESTAMP)
		`, hostname, machineID, timestamp)

		if err != nil {
			logger.Error("Error inserting asset: " + err.Error())
			return err
		}

		logger.Debug("Created new asset: hostname=" + hostname + " machine_id=" + machineID)
		return nil
	} else if err != nil {
		logger.Error("Error checking existing asset: " + err.Error())
		return err
	}

	_, err = tx.Exec(`
		UPDATE assets
		SET last_seen = $1
		WHERE asset_id = $2
	`, timestamp, existingAssetID)

	if err != nil {
		logger.Error("Error updating asset last_seen: " + err.Error())
		return err
	}

	if !existingIsActive {
		err = am.deactivateOldAssets(tx, hostname, machineID)
		if err != nil {
			return err
		}

		_, err = tx.Exec(`
			UPDATE assets
			SET is_active = TRUE, deactivated_at = NULL
			WHERE asset_id = $1
		`, existingAssetID)

		if err != nil {
			logger.Error("Error reactivating asset: " + err.Error())
			return err
		}

		logger.Info("Reactivated asset: hostname=" + hostname + " machine_id=" + machineID)
	}

	return nil
}

func (am *AssetManager) deactivateOldAssets(tx *sql.Tx, hostname string, newMachineID string) error {
	_, err := tx.Exec(`
		UPDATE assets
		SET is_active = FALSE, deactivated_at = CURRENT_TIMESTAMP
		WHERE hostname = $1
		AND machine_id != $2
		AND is_active = TRUE
	`, hostname, newMachineID)

	if err != nil {
		logger.Error("Error deactivating old assets: " + err.Error())
		return err
	}

	return nil
}

func (am *AssetManager) GetActiveAsset(hostname string) (*Asset, error) {
	var asset Asset
	var deactivatedAt sql.NullTime

	err := am.db.QueryRow(`
		SELECT asset_id, hostname, machine_id, first_seen, last_seen, is_active, created_at, deactivated_at
		FROM assets
		WHERE hostname = $1 AND is_active = TRUE
		LIMIT 1
	`, hostname).Scan(
		&asset.AssetID,
		&asset.Hostname,
		&asset.MachineID,
		&asset.FirstSeen,
		&asset.LastSeen,
		&asset.IsActive,
		&asset.CreatedAt,
		&deactivatedAt,
	)

	if err != nil {
		return nil, err
	}

	if deactivatedAt.Valid {
		asset.DeactivatedAt = &deactivatedAt.Time
	}

	return &asset, nil
}

func (am *AssetManager) GetAssetByMachineID(machineID string) (*Asset, error) {
	var asset Asset
	var deactivatedAt sql.NullTime

	err := am.db.QueryRow(`
		SELECT asset_id, hostname, machine_id, first_seen, last_seen, is_active, created_at, deactivated_at
		FROM assets
		WHERE machine_id = $1
		LIMIT 1
	`, machineID).Scan(
		&asset.AssetID,
		&asset.Hostname,
		&asset.MachineID,
		&asset.FirstSeen,
		&asset.LastSeen,
		&asset.IsActive,
		&asset.CreatedAt,
		&deactivatedAt,
	)

	if err != nil {
		return nil, err
	}

	if deactivatedAt.Valid {
		asset.DeactivatedAt = &deactivatedAt.Time
	}

	return &asset, nil
}

type Asset struct {
	AssetID       int
	Hostname      string
	MachineID     string
	FirstSeen     time.Time
	LastSeen      time.Time
	IsActive      bool
	CreatedAt     time.Time
	DeactivatedAt *time.Time
}
