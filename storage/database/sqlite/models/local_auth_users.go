// Code generated by BobGen sqlite v0.22.0. DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"

	"github.com/RobinThrift/stuff/storage/database/sqlite/types"
	"github.com/aarondl/opt/omit"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/dialect"
	"github.com/stephenafamo/bob/dialect/sqlite/im"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
	"github.com/stephenafamo/bob/dialect/sqlite/um"
)

// LocalAuthUser is an object representing the database table.
type LocalAuthUser struct {
	ID                     int64                `db:"id,pk" `
	Username               string               `db:"username" `
	Algorithm              string               `db:"algorithm" `
	Params                 string               `db:"params" `
	Salt                   []byte               `db:"salt" `
	Password               []byte               `db:"password" `
	RequiresPasswordChange bool                 `db:"requires_password_change" `
	CreatedAt              types.SQLiteDatetime `db:"created_at" `
	UpdatedAt              types.SQLiteDatetime `db:"updated_at" `
}

// LocalAuthUserSlice is an alias for a slice of pointers to LocalAuthUser.
// This should almost always be used instead of []*LocalAuthUser.
type LocalAuthUserSlice []*LocalAuthUser

// LocalAuthUsers contains methods to work with the local_auth_users table
var LocalAuthUsers = sqlite.NewTablex[*LocalAuthUser, LocalAuthUserSlice, *LocalAuthUserSetter]("", "local_auth_users")

// LocalAuthUsersQuery is a query on the local_auth_users table
type LocalAuthUsersQuery = *sqlite.ViewQuery[*LocalAuthUser, LocalAuthUserSlice]

// LocalAuthUsersStmt is a prepared statment on local_auth_users
type LocalAuthUsersStmt = bob.QueryStmt[*LocalAuthUser, LocalAuthUserSlice]

// LocalAuthUserSetter is used for insert/upsert/update operations
// All values are optional, and do not have to be set
// Generated columns are not included
type LocalAuthUserSetter struct {
	ID                     omit.Val[int64]                `db:"id,pk"`
	Username               omit.Val[string]               `db:"username"`
	Algorithm              omit.Val[string]               `db:"algorithm"`
	Params                 omit.Val[string]               `db:"params"`
	Salt                   omit.Val[[]byte]               `db:"salt"`
	Password               omit.Val[[]byte]               `db:"password"`
	RequiresPasswordChange omit.Val[bool]                 `db:"requires_password_change"`
	CreatedAt              omit.Val[types.SQLiteDatetime] `db:"created_at"`
	UpdatedAt              omit.Val[types.SQLiteDatetime] `db:"updated_at"`
}

func (s LocalAuthUserSetter) SetColumns() []string {
	vals := make([]string, 0, 9)
	if !s.ID.IsUnset() {
		vals = append(vals, "id")
	}

	if !s.Username.IsUnset() {
		vals = append(vals, "username")
	}

	if !s.Algorithm.IsUnset() {
		vals = append(vals, "algorithm")
	}

	if !s.Params.IsUnset() {
		vals = append(vals, "params")
	}

	if !s.Salt.IsUnset() {
		vals = append(vals, "salt")
	}

	if !s.Password.IsUnset() {
		vals = append(vals, "password")
	}

	if !s.RequiresPasswordChange.IsUnset() {
		vals = append(vals, "requires_password_change")
	}

	if !s.CreatedAt.IsUnset() {
		vals = append(vals, "created_at")
	}

	if !s.UpdatedAt.IsUnset() {
		vals = append(vals, "updated_at")
	}

	return vals
}

func (s LocalAuthUserSetter) Overwrite(t *LocalAuthUser) {
	if !s.ID.IsUnset() {
		t.ID, _ = s.ID.Get()
	}
	if !s.Username.IsUnset() {
		t.Username, _ = s.Username.Get()
	}
	if !s.Algorithm.IsUnset() {
		t.Algorithm, _ = s.Algorithm.Get()
	}
	if !s.Params.IsUnset() {
		t.Params, _ = s.Params.Get()
	}
	if !s.Salt.IsUnset() {
		t.Salt, _ = s.Salt.Get()
	}
	if !s.Password.IsUnset() {
		t.Password, _ = s.Password.Get()
	}
	if !s.RequiresPasswordChange.IsUnset() {
		t.RequiresPasswordChange, _ = s.RequiresPasswordChange.Get()
	}
	if !s.CreatedAt.IsUnset() {
		t.CreatedAt, _ = s.CreatedAt.Get()
	}
	if !s.UpdatedAt.IsUnset() {
		t.UpdatedAt, _ = s.UpdatedAt.Get()
	}
}

func (s LocalAuthUserSetter) Apply(q *dialect.UpdateQuery) {
	if !s.ID.IsUnset() {
		um.Set("id").ToArg(s.ID).Apply(q)
	}
	if !s.Username.IsUnset() {
		um.Set("username").ToArg(s.Username).Apply(q)
	}
	if !s.Algorithm.IsUnset() {
		um.Set("algorithm").ToArg(s.Algorithm).Apply(q)
	}
	if !s.Params.IsUnset() {
		um.Set("params").ToArg(s.Params).Apply(q)
	}
	if !s.Salt.IsUnset() {
		um.Set("salt").ToArg(s.Salt).Apply(q)
	}
	if !s.Password.IsUnset() {
		um.Set("password").ToArg(s.Password).Apply(q)
	}
	if !s.RequiresPasswordChange.IsUnset() {
		um.Set("requires_password_change").ToArg(s.RequiresPasswordChange).Apply(q)
	}
	if !s.CreatedAt.IsUnset() {
		um.Set("created_at").ToArg(s.CreatedAt).Apply(q)
	}
	if !s.UpdatedAt.IsUnset() {
		um.Set("updated_at").ToArg(s.UpdatedAt).Apply(q)
	}
}

func (s LocalAuthUserSetter) Insert() bob.Mod[*dialect.InsertQuery] {
	vals := make([]bob.Expression, 0, 9)
	if !s.ID.IsUnset() {
		vals = append(vals, sqlite.Arg(s.ID))
	}

	if !s.Username.IsUnset() {
		vals = append(vals, sqlite.Arg(s.Username))
	}

	if !s.Algorithm.IsUnset() {
		vals = append(vals, sqlite.Arg(s.Algorithm))
	}

	if !s.Params.IsUnset() {
		vals = append(vals, sqlite.Arg(s.Params))
	}

	if !s.Salt.IsUnset() {
		vals = append(vals, sqlite.Arg(s.Salt))
	}

	if !s.Password.IsUnset() {
		vals = append(vals, sqlite.Arg(s.Password))
	}

	if !s.RequiresPasswordChange.IsUnset() {
		vals = append(vals, sqlite.Arg(s.RequiresPasswordChange))
	}

	if !s.CreatedAt.IsUnset() {
		vals = append(vals, sqlite.Arg(s.CreatedAt))
	}

	if !s.UpdatedAt.IsUnset() {
		vals = append(vals, sqlite.Arg(s.UpdatedAt))
	}

	return im.Values(vals...)
}

type localAuthUserColumnNames struct {
	ID                     string
	Username               string
	Algorithm              string
	Params                 string
	Salt                   string
	Password               string
	RequiresPasswordChange string
	CreatedAt              string
	UpdatedAt              string
}

var LocalAuthUserColumns = struct {
	ID                     sqlite.Expression
	Username               sqlite.Expression
	Algorithm              sqlite.Expression
	Params                 sqlite.Expression
	Salt                   sqlite.Expression
	Password               sqlite.Expression
	RequiresPasswordChange sqlite.Expression
	CreatedAt              sqlite.Expression
	UpdatedAt              sqlite.Expression
}{
	ID:                     sqlite.Quote("local_auth_users", "id"),
	Username:               sqlite.Quote("local_auth_users", "username"),
	Algorithm:              sqlite.Quote("local_auth_users", "algorithm"),
	Params:                 sqlite.Quote("local_auth_users", "params"),
	Salt:                   sqlite.Quote("local_auth_users", "salt"),
	Password:               sqlite.Quote("local_auth_users", "password"),
	RequiresPasswordChange: sqlite.Quote("local_auth_users", "requires_password_change"),
	CreatedAt:              sqlite.Quote("local_auth_users", "created_at"),
	UpdatedAt:              sqlite.Quote("local_auth_users", "updated_at"),
}

type localAuthUserWhere[Q sqlite.Filterable] struct {
	ID                     sqlite.WhereMod[Q, int64]
	Username               sqlite.WhereMod[Q, string]
	Algorithm              sqlite.WhereMod[Q, string]
	Params                 sqlite.WhereMod[Q, string]
	Salt                   sqlite.WhereMod[Q, []byte]
	Password               sqlite.WhereMod[Q, []byte]
	RequiresPasswordChange sqlite.WhereMod[Q, bool]
	CreatedAt              sqlite.WhereMod[Q, types.SQLiteDatetime]
	UpdatedAt              sqlite.WhereMod[Q, types.SQLiteDatetime]
}

func LocalAuthUserWhere[Q sqlite.Filterable]() localAuthUserWhere[Q] {
	return localAuthUserWhere[Q]{
		ID:                     sqlite.Where[Q, int64](LocalAuthUserColumns.ID),
		Username:               sqlite.Where[Q, string](LocalAuthUserColumns.Username),
		Algorithm:              sqlite.Where[Q, string](LocalAuthUserColumns.Algorithm),
		Params:                 sqlite.Where[Q, string](LocalAuthUserColumns.Params),
		Salt:                   sqlite.Where[Q, []byte](LocalAuthUserColumns.Salt),
		Password:               sqlite.Where[Q, []byte](LocalAuthUserColumns.Password),
		RequiresPasswordChange: sqlite.Where[Q, bool](LocalAuthUserColumns.RequiresPasswordChange),
		CreatedAt:              sqlite.Where[Q, types.SQLiteDatetime](LocalAuthUserColumns.CreatedAt),
		UpdatedAt:              sqlite.Where[Q, types.SQLiteDatetime](LocalAuthUserColumns.UpdatedAt),
	}
}

// FindLocalAuthUser retrieves a single record by primary key
// If cols is empty Find will return all columns.
func FindLocalAuthUser(ctx context.Context, exec bob.Executor, IDPK int64, cols ...string) (*LocalAuthUser, error) {
	if len(cols) == 0 {
		return LocalAuthUsers.Query(
			ctx, exec,
			SelectWhere.LocalAuthUsers.ID.EQ(IDPK),
		).One()
	}

	return LocalAuthUsers.Query(
		ctx, exec,
		SelectWhere.LocalAuthUsers.ID.EQ(IDPK),
		sm.Columns(LocalAuthUsers.Columns().Only(cols...)),
	).One()
}

// LocalAuthUserExists checks the presence of a single record by primary key
func LocalAuthUserExists(ctx context.Context, exec bob.Executor, IDPK int64) (bool, error) {
	return LocalAuthUsers.Query(
		ctx, exec,
		SelectWhere.LocalAuthUsers.ID.EQ(IDPK),
	).Exists()
}

// PrimaryKeyVals returns the primary key values of the LocalAuthUser
func (o *LocalAuthUser) PrimaryKeyVals() bob.Expression {
	return sqlite.Arg(o.ID)
}

// Update uses an executor to update the LocalAuthUser
func (o *LocalAuthUser) Update(ctx context.Context, exec bob.Executor, s *LocalAuthUserSetter) error {
	return LocalAuthUsers.Update(ctx, exec, s, o)
}

// Delete deletes a single LocalAuthUser record with an executor
func (o *LocalAuthUser) Delete(ctx context.Context, exec bob.Executor) error {
	return LocalAuthUsers.Delete(ctx, exec, o)
}

// Reload refreshes the LocalAuthUser using the executor
func (o *LocalAuthUser) Reload(ctx context.Context, exec bob.Executor) error {
	o2, err := LocalAuthUsers.Query(
		ctx, exec,
		SelectWhere.LocalAuthUsers.ID.EQ(o.ID),
	).One()
	if err != nil {
		return err
	}

	*o = *o2

	return nil
}

func (o LocalAuthUserSlice) UpdateAll(ctx context.Context, exec bob.Executor, vals LocalAuthUserSetter) error {
	return LocalAuthUsers.Update(ctx, exec, &vals, o...)
}

func (o LocalAuthUserSlice) DeleteAll(ctx context.Context, exec bob.Executor) error {
	return LocalAuthUsers.Delete(ctx, exec, o...)
}

func (o LocalAuthUserSlice) ReloadAll(ctx context.Context, exec bob.Executor) error {
	var mods []bob.Mod[*dialect.SelectQuery]

	IDPK := make([]int64, len(o))

	for i, o := range o {
		IDPK[i] = o.ID
	}

	mods = append(mods,
		SelectWhere.LocalAuthUsers.ID.In(IDPK...),
	)

	o2, err := LocalAuthUsers.Query(ctx, exec, mods...).All()
	if err != nil {
		return err
	}

	for _, old := range o {
		for _, new := range o2 {
			if new.ID != old.ID {
				continue
			}

			*old = *new
			break
		}
	}

	return nil
}
