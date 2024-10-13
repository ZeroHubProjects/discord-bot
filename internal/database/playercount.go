package database

func (d *Database) InsertPlayerCount(playerCount int) error {
	r, err := d.conn.Exec("INSERT INTO server_metrics(record_time, player_count) VALUES (NOW(), ?)", playerCount)
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
