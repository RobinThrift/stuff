package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/RobinThrift/stuff/storage/database"
	"github.com/RobinThrift/stuff/storage/database/sqlite/models"
	"github.com/RobinThrift/stuff/storage/database/sqlite/types"
	"github.com/aarondl/opt/omit"
	"github.com/alexedwards/scs/v2"
)

const cleanupInterval = time.Minute * 30

var ErrFindingSession = errors.New("error finding sessions")
var ErrCommittingSession = errors.New("error committing sessions")

type SQLiteSessionStore struct {
	db *database.Database
}

var _ scs.Store = (*SQLiteSessionStore)(nil)

func NewSQLiteSessionStore(db *database.Database) *SQLiteSessionStore {
	s := &SQLiteSessionStore{db: db}
	go s.cleanupTask()
	return s
}

func (s *SQLiteSessionStore) Find(token string) ([]byte, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	return s.FindCtx(ctx, token)
}

func (s *SQLiteSessionStore) Commit(token string, b []byte, expiry time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	return s.CommitCtx(ctx, token, b, expiry)
}

func (s *SQLiteSessionStore) Delete(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	return s.DeleteCtx(ctx, token)
}

func (s *SQLiteSessionStore) FindCtx(ctx context.Context, token string) ([]byte, bool, error) {
	sess, err := models.Sessions.Query(ctx, s.db, models.SelectWhere.Sessions.Token.EQ(token)).One()

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}

		slog.Error(ErrFindingSession.Error(), "error", err, "token", token)
		return nil, false, err
	}

	return sess.Data, true, nil
}

func (s *SQLiteSessionStore) CommitCtx(ctx context.Context, token string, b []byte, expiry time.Time) error {
	err := s.db.InTransaction(ctx, func(ctx context.Context, tx database.Executor) error {
		_, err := models.Sessions.Upsert(ctx, tx, true,
			[]string{"token"},
			[]string{"data", "expires_at"},
			&models.SessionSetter{
				Token:     omit.From(token),
				Data:      omit.From(b),
				ExpiresAt: omit.From(types.NewSQLiteDatetime(expiry)),
			},
		)
		return err
	})

	if err != nil {
		slog.Error(ErrCommittingSession.Error(), "error", err, "token", token)
		return fmt.Errorf("%v: %w", ErrCommittingSession, err)
	}

	return nil
}

func (s *SQLiteSessionStore) DeleteCtx(ctx context.Context, token string) error {
	err := s.db.InTransaction(ctx, func(ctx context.Context, tx database.Executor) error {
		_, err := models.Sessions.DeleteQ(ctx, tx, models.DeleteWhere.Sessions.Token.EQ(token)).All()
		return err
	})

	if err != nil {
		slog.Error("error deleting session", "error", err, "token", token)
		return err
	}

	return nil
}

func (s *SQLiteSessionStore) deleteExpired(ctx context.Context) error {
	return s.db.InTransaction(ctx, func(ctx context.Context, tx database.Executor) error {
		_, err := models.Sessions.DeleteQ(ctx, tx, models.DeleteWhere.Sessions.ExpiresAt.LT(types.NewSQLiteDatetime(time.Now()))).All()
		return err
	})
}

func (s *SQLiteSessionStore) cleanupTask() {
	ticker := time.NewTicker(cleanupInterval)
	for range ticker.C {
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			err := s.deleteExpired(ctx)
			if err != nil {
				slog.Error("error while running session cleanup task", "error", err)
			}
		}()
	}
}
