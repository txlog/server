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

func (am *AssetManager) UpsertAsset(tx *sql.Tx, hostname string, machineID string, timestamp time.Time, needsRestarting sql.NullBool, restartingReason sql.NullString, os string, agentVersion string) error {
	var existingAssetID int
	var existingIsActive bool

	err := tx.QueryRow(`
		SELECT asset_id, is_active
		FROM assets
		WHERE hostname = $1 AND machine_id = $2
	`, hostname, machineID).Scan(&existingAssetID, &existingIsActive)

	if err == sql.ErrNoRows {
		err = am.deactivateAssetsByMachineID(tx, machineID)
		if err != nil {
			return err
		}

		_, err = tx.Exec(`
			INSERT INTO assets (hostname, machine_id, first_seen, last_seen, is_active, created_at, needs_restarting, restarting_reason, os, agent_version)
			VALUES ($1, $2, $3, $3, TRUE, CURRENT_TIMESTAMP, $4, $5, $6, $7)
		`, hostname, machineID, timestamp, needsRestarting, restartingReason, os, agentVersion)

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
		SET last_seen = $1, needs_restarting = $2, restarting_reason = $3, os = $4, agent_version = $5
		WHERE asset_id = $6
	`, timestamp, needsRestarting, restartingReason, os, agentVersion, existingAssetID)

	if err != nil {
		logger.Error("Error updating asset last_seen: " + err.Error())
		return err
	}

	if !existingIsActive {
		err = am.deactivateAssetsByMachineID(tx, machineID)
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

func (am *AssetManager) deactivateAssetsByMachineID(tx *sql.Tx, machineID string) error {
	_, err := tx.Exec(`
		UPDATE assets
		SET is_active = FALSE, deactivated_at = CURRENT_TIMESTAMP
		WHERE machine_id = $1
		AND is_active = TRUE
	`, machineID)

	if err != nil {
		logger.Error("Error deactivating old assets by machine_id: " + err.Error())
		return err
	}

	return nil
}

func (am *AssetManager) GetActiveAsset(hostname string) (*Asset, error) {
	var asset Asset
	var deactivatedAt sql.NullTime
	var needsRestarting sql.NullBool
	var restartingReason sql.NullString

	err := am.db.QueryRow(`
		SELECT asset_id, hostname, machine_id, first_seen, last_seen, is_active, created_at, deactivated_at, needs_restarting, restarting_reason
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
		&needsRestarting,
		&restartingReason,
	)

	if err != nil {
		return nil, err
	}

	if deactivatedAt.Valid {
		asset.DeactivatedAt = &deactivatedAt.Time
	}

	if needsRestarting.Valid {
		asset.NeedsRestarting = &needsRestarting.Bool
	}

	if restartingReason.Valid {
		asset.RestartingReason = &restartingReason.String
	}

	return &asset, nil
}

func (am *AssetManager) GetAssetByMachineID(machineID string) (*Asset, error) {
	var asset Asset
	var deactivatedAt sql.NullTime
	var needsRestarting sql.NullBool
	var restartingReason sql.NullString

	err := am.db.QueryRow(`
		SELECT asset_id, hostname, machine_id, first_seen, last_seen, is_active, created_at, deactivated_at, needs_restarting, restarting_reason
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
		&needsRestarting,
		&restartingReason,
	)

	if err != nil {
		return nil, err
	}

	if deactivatedAt.Valid {
		asset.DeactivatedAt = &deactivatedAt.Time
	}

	if needsRestarting.Valid {
		asset.NeedsRestarting = &needsRestarting.Bool
	}

	if restartingReason.Valid {
		asset.RestartingReason = &restartingReason.String
	}

	return &asset, nil
}

type Asset struct {
	AssetID          int
	Hostname         string
	MachineID        string
	FirstSeen        time.Time
	LastSeen         time.Time
	IsActive         bool
	CreatedAt        time.Time
	DeactivatedAt    *time.Time
	NeedsRestarting  *bool
	RestartingReason *string
}
