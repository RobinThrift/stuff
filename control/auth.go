package control

import (
	"context"
	"errors"
	"log/slog"

	"github.com/RobinThrift/stuff/auth"
	"github.com/RobinThrift/stuff/storage/database"
	"github.com/RobinThrift/stuff/storage/database/sqlite"
	"github.com/stephenafamo/bob"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrUsernameEmpty = errors.New("username must not be empty")
var ErrPasswordEmpty = errors.New("password must not be empty")

type AuthController struct {
	config    AuthConfig
	db        *database.Database
	users     *UserControl
	localAuth LocalAuthRepo
}

type LocalAuthRepo interface {
	Get(ctx context.Context, tx bob.Executor, username string) (*auth.LocalAuthUser, error)
	Create(ctx context.Context, tx bob.Executor, user *auth.LocalAuthUser) error
	Update(ctx context.Context, tx bob.Executor, user *auth.LocalAuthUser) error
	DeleteByUsername(ctx context.Context, tx bob.Executor, username string) error
}

type AuthConfig struct {
	Argon2Params auth.Argon2Params
}

func NewAuthController(config AuthConfig, db *database.Database, users *UserControl, localAuth LocalAuthRepo) *AuthController {
	return &AuthController{config: config, db: db, users: users, localAuth: localAuth}
}

type GetUserForCredentialsQuery struct {
	Username        string
	PlaintextPasswd string
}

func (ac *AuthController) GetUserForCredentials(ctx context.Context, query GetUserForCredentialsQuery) (*auth.User, map[string]string, error) {
	validationErrs := map[string]string{}
	if query.Username == "" {
		validationErrs["username"] = ErrUsernameEmpty.Error()
	}

	if query.PlaintextPasswd == "" {
		validationErrs["password"] = ErrPasswordEmpty.Error()
	}

	if len(validationErrs) != 0 {
		return nil, validationErrs, nil
	}

	user, err := database.InTransaction(ctx, ac.db, func(ctx context.Context, tx database.Executor) (*auth.User, error) {
		localUser, err := ac.localAuth.Get(ctx, tx, query.Username)
		if err != nil {
			if errors.Is(err, sqlite.ErrLocalAuthUserNotFound) {
				return nil, ErrInvalidCredentials
			}

			return nil, err
		}

		passwordMatch, err := auth.CheckPassword([]byte(query.PlaintextPasswd), localUser.Password, localUser.Salt, []byte(localUser.Params))
		if err != nil {
			slog.ErrorContext(ctx, "error comparing user password", "error", err, "username", query.Username)
			return nil, ErrInvalidCredentials
		}

		if !passwordMatch {
			return nil, ErrInvalidCredentials
		}

		user, err := ac.users.GetByRef(ctx, localUser.Username)
		if err != nil {
			if errors.Is(err, auth.ErrUserNotFound) {
				slog.ErrorContext(ctx, "fetching user by auth referenced failed even after passwords match", "error", err, "username", query.Username)
				return nil, ErrInvalidCredentials
			}

			return nil, err
		}

		user.RequiresPasswordChange = localUser.RequiresPasswordChange

		return user, nil
	})

	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			validationErrs["general"] = "Invalid credentials."
			return nil, validationErrs, nil
		}

		return nil, validationErrs, err
	}

	return user, validationErrs, nil
}

type CreateUserCmd struct {
	User            *auth.User
	PlaintextPasswd string
}

func (ac *AuthController) CreateUser(ctx context.Context, cmd CreateUserCmd) error {
	return ac.db.InTransaction(ctx, func(ctx context.Context, tx database.Executor) error {
		if cmd.PlaintextPasswd == "" {
			return errors.New("initial password cannot be empty")
		}

		params, err := ac.config.Argon2Params.ToJSONString()
		if err != nil {
			return err
		}

		hash, salt, err := auth.EncryptPassword([]byte(cmd.PlaintextPasswd), ac.config.Argon2Params)
		if err != nil {
			return err
		}

		err = ac.localAuth.Create(ctx, tx, &auth.LocalAuthUser{
			Username:               cmd.User.Username,
			Algorithm:              "argon2",
			Params:                 params,
			Salt:                   salt,
			Password:               hash,
			RequiresPasswordChange: true,
		})
		if err != nil {
			return err
		}

		return ac.users.Create(ctx, &auth.User{
			Username:    cmd.User.Username,
			DisplayName: cmd.User.DisplayName,
			IsAdmin:     cmd.User.IsAdmin,
			AuthRef:     cmd.User.Username,
		})
	})
}

type ChangeUserCredentialsCmd struct {
	User                     *auth.User
	CurrPasswdPlaintext      string
	NewPasswdPlaintext       string
	NewPasswdRepeatPlaintext string
}

func (ac *AuthController) ChangeUserCredentials(ctx context.Context, cmd ChangeUserCredentialsCmd) (map[string]string, error) {
	validationErrs := map[string]string{}

	if cmd.CurrPasswdPlaintext == "" {
		validationErrs["current_password"] = "Current password must not be empty."
	}

	if cmd.NewPasswdPlaintext == "" {
		validationErrs["new_password"] = "New password must not be empty."
	} else if cmd.NewPasswdRepeatPlaintext != cmd.NewPasswdPlaintext {
		validationErrs["new_password_repeat"] = "New passwords don't match."
		validationErrs["new_password"] = validationErrs["new_password_repeat"]
	}

	if len(validationErrs) != 0 {
		return validationErrs, nil
	}

	localUser, err := ac.localAuth.Get(ctx, ac.db, cmd.User.Username)
	if err != nil {
		slog.ErrorContext(ctx, "error finding user for changing password", "error", err, "username", cmd.User.Username)
		if errors.Is(err, sqlite.ErrLocalAuthUserNotFound) {
			return nil, ErrInvalidCredentials
		}

		return nil, err
	}

	passwordMatch, err := auth.CheckPassword([]byte(cmd.CurrPasswdPlaintext), localUser.Password, localUser.Salt, []byte(localUser.Params))
	if err != nil {
		slog.ErrorContext(ctx, "error comparing user password", "error", err, "username", cmd.User.Username)
		return nil, ErrInvalidCredentials
	}

	if !passwordMatch {
		validationErrs["current_password"] = "Incorrect current password."
		return validationErrs, nil
	}

	if cmd.NewPasswdPlaintext == cmd.CurrPasswdPlaintext {
		validationErrs["new_password_repeat"] = "New password cannot be the same as the old password."
		validationErrs["new_password"] = validationErrs["new_password_repeat"]
	}

	params, err := ac.config.Argon2Params.ToJSONString()
	if err != nil {
		return nil, err
	}

	hash, salt, err := auth.EncryptPassword([]byte(cmd.NewPasswdRepeatPlaintext), ac.config.Argon2Params)
	if err != nil {
		return nil, err
	}

	localUser.Params = params
	localUser.Password = hash
	localUser.Salt = salt
	localUser.RequiresPasswordChange = false

	err = ac.localAuth.Update(ctx, ac.db, localUser)
	if err != nil {
		slog.ErrorContext(ctx, "error updating local user in DB", "error", err, "username", cmd.User.Username)
		return nil, err
	}

	return validationErrs, nil
}

func (ac *AuthController) ToggleAdmin(ctx context.Context, id int64) error {
	return ac.db.InTransaction(ctx, func(ctx context.Context, tx database.Executor) error {
		user, err := ac.users.Get(ctx, id)
		if err != nil {
			return err
		}

		user.IsAdmin = !user.IsAdmin

		err = ac.users.Update(ctx, user)
		if err != nil {
			return err
		}

		count, err := ac.users.CountAdmins(ctx)
		if err != nil {
			return err
		}

		if count == 0 {
			return errors.New("can't demote user, must always have at leas one admin")
		}

		return nil
	})
}

type ResetPasswordCmd struct {
	UserID          int64
	PlaintextPasswd string
}

func (ac *AuthController) ResetPassword(ctx context.Context, cmd ResetPasswordCmd) error {
	if cmd.PlaintextPasswd == "" {
		return ErrPasswordEmpty
	}

	return ac.db.InTransaction(ctx, func(ctx context.Context, tx database.Executor) error {
		user, err := ac.users.Get(ctx, cmd.UserID)
		if err != nil {
			return err
		}

		localUser, err := ac.localAuth.Get(ctx, tx, user.AuthRef)
		if err != nil {
			slog.ErrorContext(ctx, "error finding user for password reset", "error", err, "username", user.Username)
			return err
		}

		params, err := ac.config.Argon2Params.ToJSONString()
		if err != nil {
			return err
		}

		hash, salt, err := auth.EncryptPassword([]byte(cmd.PlaintextPasswd), ac.config.Argon2Params)
		if err != nil {
			return err
		}

		localUser.Params = params
		localUser.Password = hash
		localUser.Salt = salt
		localUser.RequiresPasswordChange = true

		err = ac.localAuth.Update(ctx, tx, localUser)
		if err != nil {
			slog.ErrorContext(ctx, "error updating local auth user", "error", err, "username", user.Username)
			return err
		}

		user.RequiresPasswordChange = true
		return nil
	})
}

func (ac *AuthController) DeleteUser(ctx context.Context, userID int64) error {
	return ac.db.InTransaction(ctx, func(ctx context.Context, tx database.Executor) error {
		user, err := ac.users.Get(ctx, userID)
		if err != nil {
			return err
		}

		err = ac.localAuth.DeleteByUsername(ctx, tx, user.AuthRef)
		if err != nil {
			return err
		}

		err = ac.users.Delete(ctx, user.ID)
		if err != nil {
			return err
		}

		return nil
	})
}
