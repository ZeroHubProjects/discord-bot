package database

import "time"

type Verification struct {
	Ckey       string    `db:"ckey"`
	DisplayKey string    `db:"display_key"`
	CreatedAt  time.Time `db:"created_at"`
}

type VerifiedPlayer struct {
	Ckey          string    `db:"ckey"`
	DisplayKey    string    `db:"display_key"`
	DiscordUserID string    `db:"discord_user_id"`
	CreatedAt     time.Time `db:"created_at"`
}
