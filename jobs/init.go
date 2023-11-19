package jobs

import (
	"context"
	"errors"

	"github.com/RobinThrift/stuff/auth"
	"github.com/RobinThrift/stuff/control"
	"github.com/RobinThrift/stuff/storage/database"
)

type InitJob struct {
	config InitJobConfig
	db     *database.Database
	auth   *control.AuthController
	users  *control.UserControl
}

type InitJobConfig struct {
	Username string
	Password string
}

func NewInitJob(config InitJobConfig, db *database.Database, auth *control.AuthController, users *control.UserControl) *InitJob {
	return &InitJob{config: config, db: db, auth: auth, users: users}
}

func (ij *InitJob) Run(ctx context.Context) error {
	return ij.db.InTransaction(ctx, func(ctx context.Context, _ database.Executor) error {
		user, err := ij.users.GetByUsername(ctx, ij.config.Username)
		if err != nil {
			if !errors.Is(err, control.ErrUserNotFound) {
				return err
			}
		}

		if user != nil {
			return nil
		}

		err = ij.auth.CreateUser(ctx, control.CreateUserCmd{User: &auth.User{Username: ij.config.Username, DisplayName: "Admin", IsAdmin: true}, PlaintextPasswd: ij.config.Password})
		return err
	})

}
