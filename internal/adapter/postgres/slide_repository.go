package postgres

import (
	"context"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"

	"github.com/david7482/aws-serverless-service/internal/domain"
)

type repoSlide struct {
	ID        int    `db:"id"`
	ChannelID int    `db:"channel_id"`
	URL       string `db:"url"`
	Page      int    `db:"page"`
	Current   bool   `db:"current"`
}

type repoColumnPatternSlide struct {
	ID        string
	ChannelID string
	URL       string
	Page      string
	Current   string
}

const repoTableSlide = "slide"

var repoColumnSlide = repoColumnPatternSlide{
	ID:        "id",
	ChannelID: "channel_id",
	URL:       "url",
	Page:      "page",
	Current:   "current",
}

func (c *repoColumnPatternSlide) columns() string {
	return strings.Join([]string{
		c.ID,
		c.ChannelID,
		c.URL,
		c.Page,
		c.Current,
	}, ", ")
}

func (r *PostgresRepository) GetSlideURLByPage(ctx context.Context, channelID, page int) (url string, p int, err domain.Error) {
	slide, err := r.getSlideByPage(ctx, channelID, page)
	if err != nil {
		return "", 0, err
	}

	return slide.URL, slide.Page, nil
}

func (r *PostgresRepository) getSlideByPage(ctx context.Context, channelID, page int) (*repoSlide, domain.Error) {
	query, args, err := r.pgsq.Select(repoColumnSlide.columns()).
		From(repoTableSlide).
		Where(sq.Eq{
			repoColumnSlide.ChannelID: channelID,
			repoColumnSlide.Page:      page,
		}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, domain.NewInternalError("", err)
	}

	row := repoSlide{}
	// get one row from result
	if err = r.db.GetContext(ctx, &row, query, args...); err != nil {
		return nil, domain.NewExternalError("", nil, err)
	}
	return &row, nil
}

func (r *PostgresRepository) GetLastPageNumber(ctx context.Context, channelID int) (int, domain.Error) {
	query, args, err := r.pgsq.Select(repoColumnSlide.Page).
		From(repoTableSlide).
		Where(sq.Eq{
			repoColumnSlide.ChannelID: channelID,
		}).
		OrderBy(fmt.Sprintf("%v desc", repoColumnSlide.Page)).
		Limit(1).
		ToSql()
	if err != nil {
		return 0, domain.NewInternalError("", err)
	}

	var last int
	if err = r.db.GetContext(ctx, &last, query, args...); err != nil {
		return 0, domain.NewExternalError("", nil, err)
	}
	return last, nil
}

func (r *PostgresRepository) UpdateCurrentPage(ctx context.Context, channelID, page int) domain.Error {
	tx, err := r.beginTx()
	if err != nil {
		return err
	}
	defer func() {
		err = r.finishTx(err, tx)
	}()

	err = r.updateCurrentPage(ctx, tx, channelID, page)
	return err
}

func (r *PostgresRepository) updateCurrentPage(ctx context.Context, db sqlContextGetter, channelID, page int) domain.Error {
	// Set the current page to be enabled
	query, args, err := r.pgsq.Update(repoTableSlide).
		Set(repoColumnSlide.Current, true).
		Where(sq.Eq{
			repoColumnSlide.ChannelID: channelID,
			repoColumnSlide.Page:      page,
		}).
		ToSql()
	if err != nil {
		return domain.NewInternalError("", err)
	}
	if _, err = db.ExecContext(ctx, query, args...); err != nil {
		return domain.NewExternalError("", nil, err)
	}

	// Set all other pages are disabled
	query, args, err = r.pgsq.Update(repoTableSlide).
		Set(repoColumnSlide.Current, false).
		Where(sq.Eq{repoColumnSlide.ChannelID: channelID}).
		Where(sq.NotEq{repoColumnSlide.Page: page}).
		ToSql()
	if err != nil {
		return domain.NewInternalError("", err)
	}
	if _, err = db.ExecContext(ctx, query, args...); err != nil {
		return domain.NewExternalError("", nil, err)
	}

	return nil
}

func (r *PostgresRepository) GetEnabledSlideURL(ctx context.Context, channelID int) (string, domain.Error) {
	query, args, err := r.pgsq.Select(repoColumnSlide.columns()).
		From(repoTableSlide).
		Where(sq.Eq{
			repoColumnSlide.ChannelID: channelID,
			repoColumnSlide.Current:   true,
		}).
		Limit(1).
		ToSql()
	if err != nil {
		return "", domain.NewInternalError("", err)
	}

	row := repoSlide{}
	// get one row from result
	if err = r.db.GetContext(ctx, &row, query, args...); err != nil {
		return "", domain.NewExternalError("", nil, err)
	}
	return row.URL, nil
}
