package postgres

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/hashicorp/go-multierror"
	"github.com/jmoiron/sqlx"

	"github.com/david7482/aws-serverless-service/internal/domain"
)

type PostgresRepository struct {
	db   *sqlx.DB
	pgsq sq.StatementBuilderType
}

func NewPostgresRepository(ctx context.Context, db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
		// set the default placeholder as $ instead of ? because postgres uses $
		pgsq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

// sqlContextGetter is an interface provided both by transaction and standard db connection
type sqlContextGetter interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func (r *PostgresRepository) beginTx() (*sqlx.Tx, domain.Error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, domain.NewExternalError("", nil, err)
	}
	return tx, nil
}

// finishTx close an open transaction
// If error is provided, abort the transaction.
// If err is nil, commit the transaction.
func (r *PostgresRepository) finishTx(err domain.Error, tx *sqlx.Tx) domain.Error {
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			wrapError := multierror.Append(err, rollbackErr)
			return domain.NewExternalError("", nil, wrapError)
		}

		return err
	} else {
		if commitErr := tx.Commit(); commitErr != nil {
			return domain.NewExternalError("", nil, commitErr)
		}

		return nil
	}
}
