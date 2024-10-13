package database

import "database/sql"

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
