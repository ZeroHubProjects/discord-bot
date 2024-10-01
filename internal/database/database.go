package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ZeroHubProjects/discord-bot/internal/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Database struct {
	cfg    config.DatabaseConfig
	conn   *sqlx.DB
	logger *zap.SugaredLogger
}

func NewDatabase(cfg config.DatabaseConfig, logger *zap.SugaredLogger) (*Database, error) {
	db := Database{cfg: cfg, logger: logger}

	err := db.connect()
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	db.verifyConnection()
	err = db.verifyConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to verify database connection: %w", err)
	}

	logger.Debug("connection established")
	return &db, nil
}

func (d *Database) connect() error {
	// For more information on the parameters see https://github.com/go-sql-driver/mysql#parameters
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", d.cfg.Username, d.cfg.Password, d.cfg.Address, d.cfg.Port, d.cfg.DatabaseName)

	conn, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return err
	}
	// MariaDB default server timeout is 8 hours, we must handle reconnection before that happens
	// otherwise "unexpected EOF" and "invalid connection" errors will occur.
	// For more info see https://mariadb.com/docs/server/ref/mdb/system-variables/wait_timeout/
	// The lifetime is set to "under 5 minutes" as per go-sql-driver recommendations
	// For more info see https://github.com/go-sql-driver/mysql#important-settings
	conn.SetConnMaxLifetime(3 * time.Minute)
	conn.SetMaxIdleConns(10)
	conn.SetMaxOpenConns(10)
	d.conn = conn
	return nil
}

func (d *Database) verifyConnection() error {
	err := d.conn.Ping()
	if err != nil {
		d.logger.Debugf("trying to reconnect a disconnected database, reason: %v", err)
		err := d.connect()
		if err != nil {
			return fmt.Errorf("failed to reconnect the database: %v", err)
		}
	}
	return nil
}

func (d *Database) GetVerifiedPlayer(userID string) (*VerifiedPlayer, error) {
	var v VerifiedPlayer
	err := d.conn.Get(&v, "SELECT * FROM player_discord WHERE discord_user_id = ?", userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (d *Database) FetchVerification(code string) (*Verification, error) {
	var v Verification
	err := d.conn.Get(&v, "SELECT ckey, display_key, created_at FROM verification WHERE code = ?", code)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (d *Database) DeleteVerification(code string) error {
	r, err := d.conn.Exec("DELETE FROM verification WHERE code = ?", code)
	if err != nil {
		return err
	}
	affected, err := r.RowsAffected()
	if err != nil {
		d.logger.Errorf("failed to check number of affected rows: %v", err)
		return nil
	}
	if affected <= 0 {
		d.logger.Warn("expected at least one verification to be deleted, got: %d", affected)
		return nil
	}
	return nil
}

func (d *Database) CreateVerifiedAccountEntry(v Verification, userID string) error {
	r, err := d.conn.Exec("INSERT INTO player_discord(ckey, display_key, discord_user_id) VALUES (?, ?, ?)", v.Ckey, v.DisplayKey, userID)
	if err != nil {
		return err
	}
	affected, err := r.RowsAffected()
	if err != nil {
		d.logger.Errorf("failed to check number of affected rows: %v", err)
		return nil
	}
	if affected <= 0 {
		d.logger.Warn("expected an entry to be inserted, got %d affected rows", affected)
		return nil
	}
	return nil
}

func (d *Database) Close() {
	if d.conn != nil {
		_ = d.conn.Close()
	}
}
