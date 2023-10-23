package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/RobinThrift/stuff/storage/database/sqlite/models"
	"github.com/RobinThrift/stuff/storage/database/sqlite/types"
	"github.com/aarondl/opt/omit"
	"github.com/alexedwards/scs/v2"
	"github.com/stephenafamo/bob"
)

const cleanupInterval = time.Minute * 30

type SQLiteSessionStore struct {
	db bob.Executor
}

var _ scs.Store = (*SQLiteSessionStore)(nil)

func NewSQLiteSessionStore(db bob.Executor) *SQLiteSessionStore {
	s := &SQLiteSessionStore{db: db}
	go s.cleanupTask()
	return s
}

func (s *SQLiteSessionStore) Find(token string) ([]byte, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	sess, err := models.Sessions.Query(ctx, s.db, models.SelectWhere.Sessions.Token.EQ(token)).One()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}

		slog.Error("error finding session", "error", err, "token", token)
		return nil, false, err
	}

	return sess.Data, true, nil
}

func (s *SQLiteSessionStore) Commit(token string, b []byte, expiry time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	_, err := models.Sessions.Upsert(ctx, s.db, true,
		[]string{"token"},
		[]string{"data", "expires_at"},
		&models.SessionSetter{
			Token:     omit.From(token),
			Data:      omit.From(b),
			ExpiresAt: omit.From(types.NewSQLiteDatetime(expiry)),
		},
	)

	if err != nil {
		slog.Error("error committing session", "error", err, "token", token)
		return err
	}

	return nil
}

func (s *SQLiteSessionStore) Delete(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	_, err := models.Sessions.DeleteQ(ctx, s.db, models.DeleteWhere.Sessions.Token.EQ(token)).All()
	if err != nil {
		slog.Error("error deleting session", "error", err, "token", token)
		return err
	}

	return nil
}

func (s *SQLiteSessionStore) deleteExpired(ctx context.Context) error {
	_, err := models.Sessions.DeleteQ(ctx, s.db, models.DeleteWhere.Sessions.ExpiresAt.LT(types.NewSQLiteDatetime(time.Now()))).All()
	return err
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
