package auth

import (
	"context"
	"errors"
	"log/slog"

	"github.com/RobinThrift/stuff/storage/database"
	"github.com/stephenafamo/bob"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrUsernameEmpty = errors.New("username must not be empty")
var ErrPasswordEmpty = errors.New("password must not be empty")

type Control struct {
	DB            *database.Database
	UserRepo      UserRepo
	LocalAuthRepo LocalAuthRepo
	Argon2Params  Argon2Params
}

type LocalAuthRepo interface {
	GetLocalUser(ctx context.Context, tx bob.Executor, username string) (*LocalAuthUser, error)
	CreateLocalUser(ctx context.Context, tx bob.Executor, user *LocalAuthUser) (*LocalAuthUser, error)
	UpdateLocalUser(ctx context.Context, tx bob.Executor, user *LocalAuthUser) error
	DeleteByUsername(ctx context.Context, tx bob.Executor, username string) error
}

type UserRepo interface {
	List(ctx context.Context, exec bob.Executor, query ListUsersQuery) (*UserListPage, error)
	Create(ctx context.Context, tx bob.Executor, toCreate *User) (*User, error)
	Update(ctx context.Context, tx bob.Executor, toUpdate *User) (*User, error)
	Get(ctx context.Context, tx bob.Executor, id int64) (*User, error)
	GetByRef(ctx context.Context, tx bob.Executor, ref string) (*User, error)
	CountAdmins(ctx context.Context, tx bob.Executor) (int64, error)
	Delete(ctx context.Context, tx bob.Executor, id int64) error
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

		_, err = c.createLocalAuthUser(ctx, &User{Username: username, DisplayName: "Admin", IsAdmin: true}, plaintextPasswd)
		return err
	})
}

func (c *Control) createLocalAuthUser(ctx context.Context, toCreate *User, plaintextPasswd string) (*User, error) {
	return database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) (*User, error) {
		if plaintextPasswd == "" {
			return nil, errors.New("initial password cannot be empty")
		}

		params, err := c.Argon2Params.toJSONString()
		if err != nil {
			return nil, err
		}

		hash, salt, err := encryptPassword([]byte(plaintextPasswd), c.Argon2Params)
		if err != nil {
			return nil, err
		}

		_, err = c.LocalAuthRepo.CreateLocalUser(ctx, tx, &LocalAuthUser{
			Username:               toCreate.Username,
			Algorithm:              "argon2",
			Params:                 params,
			Salt:                   salt,
			Password:               hash,
			RequiresPasswordChange: true,
		})
		if err != nil {
			return nil, err
		}

		created, err := c.createUser(ctx, &User{
			Username:    toCreate.Username,
			DisplayName: toCreate.DisplayName,
			IsAdmin:     toCreate.IsAdmin,
			AuthRef:     toCreate.Username,
		}, toCreate.Username)
		if err != nil {
			return nil, err
		}

		return created, nil
	})
}

func (c *Control) updateUser(ctx context.Context, toUpdate *User) (*User, error) {
	return database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) (*User, error) {
		return c.UserRepo.Update(ctx, tx, toUpdate)
	})
}

func (c *Control) createUser(ctx context.Context, toCreate *User, authRef string) (*User, error) {
	return database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) (*User, error) {
		user, err := c.UserRepo.Create(ctx, tx, &User{
			Username:    toCreate.Username,
			DisplayName: toCreate.DisplayName,
			IsAdmin:     toCreate.IsAdmin,
			AuthRef:     authRef,
		})
		if err != nil {
			return nil, err
		}

		return &User{
			ID:          user.ID,
			Username:    user.Username,
			DisplayName: user.DisplayName,
			IsAdmin:     user.IsAdmin,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		}, nil
	})
}

func (c *Control) getUser(ctx context.Context, id int64) (*User, error) {
	return database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) (*User, error) {
		user, err := c.UserRepo.Get(ctx, tx, id)
		if err != nil {
			return nil, err
		}

		return &User{
			ID:          user.ID,
			Username:    user.Username,
			DisplayName: user.DisplayName,
			IsAdmin:     user.IsAdmin,
			AuthRef:     user.AuthRef,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		}, nil
	})
}

func (c *Control) getUserByRef(ctx context.Context, ref string) (*User, error) {
	return database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) (*User, error) {
		user, err := c.UserRepo.GetByRef(ctx, tx, ref)
		if err != nil {
			return nil, err
		}

		return &User{
			ID:          user.ID,
			Username:    user.Username,
			DisplayName: user.DisplayName,
			IsAdmin:     user.IsAdmin,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		}, nil
	})
}

func (c *Control) listUsers(ctx context.Context, query ListUsersQuery) (*UserListPage, error) {
	return database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) (*UserListPage, error) {
		return c.UserRepo.List(ctx, tx, query)
	})
}

func (c *Control) getUserForCredentials(ctx context.Context, username string, plaintextPasswd string) (*User, map[string]string, error) {
	validationErrs := map[string]string{}
	if username == "" {
		validationErrs["username"] = ErrUsernameEmpty.Error()
	}

	if plaintextPasswd == "" {
		validationErrs["password"] = ErrPasswordEmpty.Error()
	}

	if len(validationErrs) != 0 {
		return nil, validationErrs, nil
	}

	user, err := database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) (*User, error) {
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

		user, err := c.getUserByRef(ctx, localUser.Username)
		if err != nil {
			if errors.Is(err, ErrUserNotFound) {
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
	user                     *User
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

func (c *Control) toggleAdmin(ctx context.Context, id int64) (*User, error) {
	return database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) (*User, error) {
		user, err := c.getUser(ctx, id)
		if err != nil {
			return nil, err
		}

		user.IsAdmin = !user.IsAdmin

		updated, err := c.updateUser(ctx, user)
		if err != nil {
			return nil, err
		}

		count, err := c.UserRepo.CountAdmins(ctx, tx)
		if err != nil {
			return nil, err
		}

		if count == 0 {
			return nil, errors.New("can't demote user, must always have at leas one admin")
		}

		return updated, nil
	})
}

func (c *Control) resetPassword(ctx context.Context, id int64, plaintextPasswd string) (*User, error) {
	if plaintextPasswd == "" {
		return nil, ErrPasswordEmpty
	}

	return database.InTransaction(ctx, c.DB, func(ctx context.Context, tx bob.Tx) (*User, error) {
		user, err := c.getUser(ctx, id)
		if err != nil {
			return nil, err
		}

		localUser, err := c.LocalAuthRepo.GetLocalUser(ctx, tx, user.AuthRef)
		if err != nil {
			slog.ErrorContext(ctx, "error finding user for password reset", "error", err, "username", user.Username)
			return nil, err
		}

		params, err := c.Argon2Params.toJSONString()
		if err != nil {
			return nil, err
		}

		hash, salt, err := encryptPassword([]byte(plaintextPasswd), c.Argon2Params)
		if err != nil {
			return nil, err
		}

		localUser.Params = params
		localUser.Password = hash
		localUser.Salt = salt
		localUser.RequiresPasswordChange = true

		err = c.LocalAuthRepo.UpdateLocalUser(ctx, tx, localUser)
		if err != nil {
			slog.ErrorContext(ctx, "error updating local auth user", "error", err, "username", user.Username)
			return nil, err
		}

		user.RequiresPasswordChange = true
		return user, nil
	})
}

func (c *Control) deleteLocalAuthUser(ctx context.Context, userID int64) error {
	return c.DB.InTransaction(ctx, func(ctx context.Context, tx bob.Tx) error {
		user, err := c.UserRepo.Get(ctx, tx, userID)
		if err != nil {
			return err
		}

		err = c.LocalAuthRepo.DeleteByUsername(ctx, tx, user.AuthRef)
		if err != nil {
			return err
		}

		err = c.UserRepo.Delete(ctx, tx, user.ID)
		if err != nil {
			return err
		}

		return nil
	})
}
