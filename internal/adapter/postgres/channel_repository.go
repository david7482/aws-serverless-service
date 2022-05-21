package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/david7482/aws-serverless-service/internal/domain"
)

type repoChannel struct {
	ID                    int       `db:"id"`
	Name                  string    `db:"name"`
	ExternalChannelID     string    `db:"external_channel_id"`
	ExternalChannelSecret string    `db:"external_channel_secret"`
	AccessToken           string    `db:"access_token"`
	AccessTokenExpiredAt  time.Time `db:"access_token_expired_at"`
	CreatedAt             time.Time `db:"created_at"`
	UpdatedAt             time.Time `db:"updated_at"`
}

type repoColumnPatternChannel struct {
	ID                    string
	Name                  string
	ExternalChannelID     string
	ExternalChannelSecret string
	AccessToken           string
	AccessTokenExpiredAt  string
	CreatedAt             string
	UpdatedAt             string
}

const repoTableChannel = "channel"

var repoColumnChannel = repoColumnPatternChannel{
	ID:                    "id",
	Name:                  "name",
	ExternalChannelID:     "external_channel_id",
	ExternalChannelSecret: "external_channel_secret",
	AccessToken:           "access_token",
	AccessTokenExpiredAt:  "access_token_expired_at",
	CreatedAt:             "created_at",
	UpdatedAt:             "updated_at",
}

func (c *repoColumnPatternChannel) columns() string {
	return strings.Join([]string{
		c.ID,
		c.Name,
		c.ExternalChannelID,
		c.ExternalChannelSecret,
		c.AccessToken,
		c.AccessTokenExpiredAt,
		c.CreatedAt,
		c.UpdatedAt,
	}, ", ")
}

func (r *PostgresRepository) CreateChannel(ctx context.Context, channel domain.Channel) (*domain.Channel, domain.Error) {
	update := map[string]interface{}{
		repoColumnChannel.Name:                  channel.Name,
		repoColumnChannel.ExternalChannelID:     channel.ExternalChannelID,
		repoColumnChannel.ExternalChannelSecret: channel.ExternalChannelSecret,
		repoColumnChannel.AccessToken:           channel.AccessToken,
		repoColumnChannel.AccessTokenExpiredAt:  channel.AccessTokenExpiredAt,
	}
	// build SQL query
	query, args, err := r.pgsq.Insert(repoTableChannel).
		SetMap(update).
		Suffix(fmt.Sprintf("returning %s", repoColumnChannel.columns())).
		ToSql()
	if err != nil {
		return nil, domain.NewInternalError("", err)
	}

	// execute SQL query
	row := repoChannel{}
	if err = r.db.GetContext(ctx, &row, query, args...); err != nil {
		return nil, domain.NewExternalError("", nil, err)
	}

	// map the query result back to domain model
	d := domain.Channel(row)
	return &d, nil
}

func (r *PostgresRepository) GetChannelByID(ctx context.Context, channelID int) (*domain.Channel, domain.Error) {
	query, args, err := r.pgsq.Select(repoColumnChannel.columns()).
		From(repoTableChannel).
		Where(sq.Eq{repoColumnChannel.ID: channelID}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, domain.NewInternalError("", err)
	}

	// get one row from result
	row := repoChannel{}
	if err = r.db.GetContext(ctx, &row, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.NewResourceNotFoundError("channel is not found", err)
		}
		return nil, domain.NewExternalError("", nil, err)
	}

	// map the query result back to domain model
	channel := domain.Channel(row)
	return &channel, nil
}

func (r *PostgresRepository) GetChannelByExternalID(ctx context.Context, externalChannelID string) (*domain.Channel, domain.Error) {
	query, args, err := r.pgsq.Select(repoColumnChannel.columns()).
		From(repoTableChannel).
		Where(sq.Eq{repoColumnChannel.ExternalChannelID: externalChannelID}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, domain.NewInternalError("", err)
	}

	// get one row from result
	row := repoChannel{}
	if err = r.db.GetContext(ctx, &row, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.NewResourceNotFoundError("channel is not found", err)
		}
		return nil, domain.NewExternalError("", nil, err)
	}

	// map the query result back to domain model
	channel := domain.Channel(row)
	return &channel, nil
}
