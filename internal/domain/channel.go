package domain

import "time"

type Channel struct {
	ID                    int
	Name                  string
	ExternalChannelID     string
	ExternalChannelSecret string
	AccessToken           string
	AccessTokenExpiredAt  time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
