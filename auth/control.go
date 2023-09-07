package auth

import (
	"context"
	"errors"
	"log/slog"

	"github.com/kodeshack/stuff/storage/database"
	"github.com/kodeshack/stuff/users"
	"github.com/stephenafamo/bob"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type Control struct {
	DB            *database.Database
	Users         *users.Control
	LocalAuthRepo LocalAuthRepo
	Argon2Params  Argon2Params
}

type LocalAuthRepo interface {
	GetLocalUser(ctx context.Context, tx bob.Executor, username string) (*LocalAuthUser, error)
	CreateLocalUser(ctx context.Context, tx bob.Executor, user *LocalAuthUser) (*LocalAuthUser, error)
	UpdateLocalUser(ctx context.Context, tx bob.Executor, user *LocalAuthUser) error
}

func (c *Control) RunInitSetup(ctx context.Context, username string, plaintextPasswd string) error {
	return c.DB.InTransaction(ctx, func(ctx context.Context, tx bob.Tx) error {
		localUser, err := c.LocalAuthRepo.GetLocalUser(ctx, tx, username)
		if err != nil {
			if !errors.Is(err, ErrLocalAuthUserNotFound) {
				return err
			}
		}

		if localUser != nil {
			return nil
		}

		if plaintextPasswd == "" {
			return errors.New("initial password cannot be empty")
		}

		params, err := c.Argon2Params.toJSONString()
		if err != nil {
			return err
		}

		hash, salt, err := encryptPassword([]byte(plaintextPasswd), c.Argon2Params)
		if err != nil {
			return err
		}

		_, err = c.LocalAuthRepo.CreateLocalUser(ctx, tx, &LocalAuthUser{
			Username:               username,
			Algorithm:              "argon2",
			Params:                 params,
			Salt:                   salt,
			Password:               hash,
			RequiresPasswordChange: true,
		})
		if err != nil {
			return err
		}

		_, err = c.Users.CreateUser(ctx, &users.User{
			Username:    username,
			DisplayName: "Admin",
			IsAdmin:     true,
		}, username)
		if err != nil {
			return err
		}

		return nil
	})
}

func (c *Control) getUserForCredentials(ctx context.Context, username string, plaintextPasswd string) (*users.User, map[string]string, error) {
	validationErrs := map[string]string{}
	if username == "" {
		validationErrs["username"] = "Username must not be empty"
	}

	if plaintextPasswd == "" {
		validationErrs["password"] = "Password must not be empty"
	}

	if len(validationErrs) != 0 {
		return nil, validationErrs, nil
	}

	user, err := database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) (*users.User, error) {
		localUser, err := c.LocalAuthRepo.GetLocalUser(ctx, tx, username)
		if err != nil {
			if errors.Is(err, ErrLocalAuthUserNotFound) {
				return nil, ErrInvalidCredentials
			}

			return nil, err
		}

		passwordMatch, err := checkPassword([]byte(plaintextPasswd), localUser.Password, localUser.Salt, []byte(localUser.Params))
		if err != nil {
			slog.ErrorContext(ctx, "error comparing user password", "error", err, "username", username)
			return nil, ErrInvalidCredentials
		}

		if !passwordMatch {
			return nil, ErrInvalidCredentials
		}

		user, err := c.Users.GetUserByRef(ctx, localUser.Username)
		if err != nil {
			if errors.Is(err, users.ErrUserNotFound) {
				slog.ErrorContext(ctx, "fetching user by auth referenced failed even after passwords match", "error", err, "username", username)
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

type changeUserCredentialsCmd struct {
	user                     *users.User
	currPasswdPlaintext      string
	newPasswdPlaintext       string
	newPasswdRepeatPlaintext string
}

func (c *Control) changeUserCredentials(ctx context.Context, cmd changeUserCredentialsCmd) (map[string]string, error) {
	validationErrs := map[string]string{}

	if cmd.currPasswdPlaintext == "" {
		validationErrs["current_password"] = "Current password must not be empty."
	}

	if cmd.newPasswdPlaintext == "" {
		validationErrs["new_password"] = "New password must not be empty."
	} else if cmd.newPasswdRepeatPlaintext != cmd.newPasswdPlaintext {
		validationErrs["new_password_repeat"] = "New passwords don't match."
		validationErrs["new_password"] = validationErrs["new_password_repeat"]
	}

	if len(validationErrs) != 0 {
		return validationErrs, nil
	}

	localUser, err := c.LocalAuthRepo.GetLocalUser(ctx, c.DB, cmd.user.Username)
	if err != nil {
		slog.ErrorContext(ctx, "error finding user for changing password", "error", err, "username", cmd.user.Username)
		if errors.Is(err, ErrLocalAuthUserNotFound) {
			return nil, ErrInvalidCredentials
		}

		return nil, err
	}

	passwordMatch, err := checkPassword([]byte(cmd.currPasswdPlaintext), localUser.Password, localUser.Salt, []byte(localUser.Params))
	if err != nil {
		slog.ErrorContext(ctx, "error comparing user password", "error", err, "username", cmd.user.Username)
		return nil, ErrInvalidCredentials
	}

	if !passwordMatch {
		validationErrs["current_password"] = "Incorrect current password."
		return validationErrs, nil
	}

	if cmd.newPasswdPlaintext == cmd.currPasswdPlaintext {
		validationErrs["new_password_repeat"] = "New password cannot be the same as the old password."
		validationErrs["new_password"] = validationErrs["new_password_repeat"]
	}

	params, err := c.Argon2Params.toJSONString()
	if err != nil {
		return nil, err
	}

	hash, salt, err := encryptPassword([]byte(cmd.newPasswdRepeatPlaintext), c.Argon2Params)
	if err != nil {
		return nil, err
	}

	localUser.Params = params
	localUser.Password = hash
	localUser.Salt = salt
	localUser.RequiresPasswordChange = false

	err = c.LocalAuthRepo.UpdateLocalUser(ctx, c.DB, localUser)
	if err != nil {
		slog.ErrorContext(ctx, "error updating local user in DB", "error", err, "username", cmd.user.Username)
		return nil, err
	}

	return validationErrs, nil
}
