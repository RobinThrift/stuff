// Code generated by BobGen sqlite v0.22.0. DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/aarondl/opt/omit"
	"github.com/kodeshack/stuff/storage/database/sqlite/types"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/clause"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/dialect"
	"github.com/stephenafamo/bob/dialect/sqlite/im"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
	"github.com/stephenafamo/bob/dialect/sqlite/um"
	"github.com/stephenafamo/bob/mods"
	"github.com/stephenafamo/bob/orm"
)

// AssetFile is an object representing the database table.
type AssetFile struct {
	ID        int64                `db:"id,pk" `
	AssetID   int64                `db:"asset_id" `
	Name      string               `db:"name" `
	Filetype  string               `db:"filetype" `
	Sha256    []byte               `db:"sha256" `
	SizeBytes int64                `db:"size_bytes" `
	CreatedBy int64                `db:"created_by" `
	CreatedAt types.SQLiteDatetime `db:"created_at" `
	UpdatedAt types.SQLiteDatetime `db:"updated_at" `

	R assetFileR `db:"-" `
}

// AssetFileSlice is an alias for a slice of pointers to AssetFile.
// This should almost always be used instead of []*AssetFile.
type AssetFileSlice []*AssetFile

// AssetFiles contains methods to work with the asset_files table
var AssetFiles = sqlite.NewTablex[*AssetFile, AssetFileSlice, *AssetFileSetter]("", "asset_files")

// AssetFilesQuery is a query on the asset_files table
type AssetFilesQuery = *sqlite.ViewQuery[*AssetFile, AssetFileSlice]

// AssetFilesStmt is a prepared statment on asset_files
type AssetFilesStmt = bob.QueryStmt[*AssetFile, AssetFileSlice]

// assetFileR is where relationships are stored.
type assetFileR struct {
	CreatedByUser *User // fk_asset_files_0
}

// AssetFileSetter is used for insert/upsert/update operations
// All values are optional, and do not have to be set
// Generated columns are not included
type AssetFileSetter struct {
	ID        omit.Val[int64]                `db:"id,pk"`
	AssetID   omit.Val[int64]                `db:"asset_id"`
	Name      omit.Val[string]               `db:"name"`
	Filetype  omit.Val[string]               `db:"filetype"`
	Sha256    omit.Val[[]byte]               `db:"sha256"`
	SizeBytes omit.Val[int64]                `db:"size_bytes"`
	CreatedBy omit.Val[int64]                `db:"created_by"`
	CreatedAt omit.Val[types.SQLiteDatetime] `db:"created_at"`
	UpdatedAt omit.Val[types.SQLiteDatetime] `db:"updated_at"`
}

func (s AssetFileSetter) SetColumns() []string {
	vals := make([]string, 0, 9)
	if !s.ID.IsUnset() {
		vals = append(vals, "id")
	}

	if !s.AssetID.IsUnset() {
		vals = append(vals, "asset_id")
	}

	if !s.Name.IsUnset() {
		vals = append(vals, "name")
	}

	if !s.Filetype.IsUnset() {
		vals = append(vals, "filetype")
	}

	if !s.Sha256.IsUnset() {
		vals = append(vals, "sha256")
	}

	if !s.SizeBytes.IsUnset() {
		vals = append(vals, "size_bytes")
	}

	if !s.CreatedBy.IsUnset() {
		vals = append(vals, "created_by")
	}

	if !s.CreatedAt.IsUnset() {
		vals = append(vals, "created_at")
	}

	if !s.UpdatedAt.IsUnset() {
		vals = append(vals, "updated_at")
	}

	return vals
}

func (s AssetFileSetter) Overwrite(t *AssetFile) {
	if !s.ID.IsUnset() {
		t.ID, _ = s.ID.Get()
	}
	if !s.AssetID.IsUnset() {
		t.AssetID, _ = s.AssetID.Get()
	}
	if !s.Name.IsUnset() {
		t.Name, _ = s.Name.Get()
	}
	if !s.Filetype.IsUnset() {
		t.Filetype, _ = s.Filetype.Get()
	}
	if !s.Sha256.IsUnset() {
		t.Sha256, _ = s.Sha256.Get()
	}
	if !s.SizeBytes.IsUnset() {
		t.SizeBytes, _ = s.SizeBytes.Get()
	}
	if !s.CreatedBy.IsUnset() {
		t.CreatedBy, _ = s.CreatedBy.Get()
	}
	if !s.CreatedAt.IsUnset() {
		t.CreatedAt, _ = s.CreatedAt.Get()
	}
	if !s.UpdatedAt.IsUnset() {
		t.UpdatedAt, _ = s.UpdatedAt.Get()
	}
}

func (s AssetFileSetter) Apply(q *dialect.UpdateQuery) {
	if !s.ID.IsUnset() {
		um.Set("id").ToArg(s.ID).Apply(q)
	}
	if !s.AssetID.IsUnset() {
		um.Set("asset_id").ToArg(s.AssetID).Apply(q)
	}
	if !s.Name.IsUnset() {
		um.Set("name").ToArg(s.Name).Apply(q)
	}
	if !s.Filetype.IsUnset() {
		um.Set("filetype").ToArg(s.Filetype).Apply(q)
	}
	if !s.Sha256.IsUnset() {
		um.Set("sha256").ToArg(s.Sha256).Apply(q)
	}
	if !s.SizeBytes.IsUnset() {
		um.Set("size_bytes").ToArg(s.SizeBytes).Apply(q)
	}
	if !s.CreatedBy.IsUnset() {
		um.Set("created_by").ToArg(s.CreatedBy).Apply(q)
	}
	if !s.CreatedAt.IsUnset() {
		um.Set("created_at").ToArg(s.CreatedAt).Apply(q)
	}
	if !s.UpdatedAt.IsUnset() {
		um.Set("updated_at").ToArg(s.UpdatedAt).Apply(q)
	}
}

func (s AssetFileSetter) Insert() bob.Mod[*dialect.InsertQuery] {
	vals := make([]bob.Expression, 0, 9)
	if !s.ID.IsUnset() {
		vals = append(vals, sqlite.Arg(s.ID))
	}

	if !s.AssetID.IsUnset() {
		vals = append(vals, sqlite.Arg(s.AssetID))
	}

	if !s.Name.IsUnset() {
		vals = append(vals, sqlite.Arg(s.Name))
	}

	if !s.Filetype.IsUnset() {
		vals = append(vals, sqlite.Arg(s.Filetype))
	}

	if !s.Sha256.IsUnset() {
		vals = append(vals, sqlite.Arg(s.Sha256))
	}

	if !s.SizeBytes.IsUnset() {
		vals = append(vals, sqlite.Arg(s.SizeBytes))
	}

	if !s.CreatedBy.IsUnset() {
		vals = append(vals, sqlite.Arg(s.CreatedBy))
	}

	if !s.CreatedAt.IsUnset() {
		vals = append(vals, sqlite.Arg(s.CreatedAt))
	}

	if !s.UpdatedAt.IsUnset() {
		vals = append(vals, sqlite.Arg(s.UpdatedAt))
	}

	return im.Values(vals...)
}

type assetFileColumnNames struct {
	ID        string
	AssetID   string
	Name      string
	Filetype  string
	Sha256    string
	SizeBytes string
	CreatedBy string
	CreatedAt string
	UpdatedAt string
}

type assetFileRelationshipJoins[Q dialect.Joinable] struct {
	CreatedByUser bob.Mod[Q]
}

func buildassetFileRelationshipJoins[Q dialect.Joinable](ctx context.Context, typ string) assetFileRelationshipJoins[Q] {
	return assetFileRelationshipJoins[Q]{
		CreatedByUser: assetFilesJoinCreatedByUser[Q](ctx, typ),
	}
}

func assetFilesJoin[Q dialect.Joinable](ctx context.Context) joinSet[assetFileRelationshipJoins[Q]] {
	return joinSet[assetFileRelationshipJoins[Q]]{
		InnerJoin: buildassetFileRelationshipJoins[Q](ctx, clause.InnerJoin),
		LeftJoin:  buildassetFileRelationshipJoins[Q](ctx, clause.LeftJoin),
		RightJoin: buildassetFileRelationshipJoins[Q](ctx, clause.RightJoin),
	}
}

var AssetFileColumns = struct {
	ID        sqlite.Expression
	AssetID   sqlite.Expression
	Name      sqlite.Expression
	Filetype  sqlite.Expression
	Sha256    sqlite.Expression
	SizeBytes sqlite.Expression
	CreatedBy sqlite.Expression
	CreatedAt sqlite.Expression
	UpdatedAt sqlite.Expression
}{
	ID:        sqlite.Quote("asset_files", "id"),
	AssetID:   sqlite.Quote("asset_files", "asset_id"),
	Name:      sqlite.Quote("asset_files", "name"),
	Filetype:  sqlite.Quote("asset_files", "filetype"),
	Sha256:    sqlite.Quote("asset_files", "sha256"),
	SizeBytes: sqlite.Quote("asset_files", "size_bytes"),
	CreatedBy: sqlite.Quote("asset_files", "created_by"),
	CreatedAt: sqlite.Quote("asset_files", "created_at"),
	UpdatedAt: sqlite.Quote("asset_files", "updated_at"),
}

type assetFileWhere[Q sqlite.Filterable] struct {
	ID        sqlite.WhereMod[Q, int64]
	AssetID   sqlite.WhereMod[Q, int64]
	Name      sqlite.WhereMod[Q, string]
	Filetype  sqlite.WhereMod[Q, string]
	Sha256    sqlite.WhereMod[Q, []byte]
	SizeBytes sqlite.WhereMod[Q, int64]
	CreatedBy sqlite.WhereMod[Q, int64]
	CreatedAt sqlite.WhereMod[Q, types.SQLiteDatetime]
	UpdatedAt sqlite.WhereMod[Q, types.SQLiteDatetime]
}

func AssetFileWhere[Q sqlite.Filterable]() assetFileWhere[Q] {
	return assetFileWhere[Q]{
		ID:        sqlite.Where[Q, int64](AssetFileColumns.ID),
		AssetID:   sqlite.Where[Q, int64](AssetFileColumns.AssetID),
		Name:      sqlite.Where[Q, string](AssetFileColumns.Name),
		Filetype:  sqlite.Where[Q, string](AssetFileColumns.Filetype),
		Sha256:    sqlite.Where[Q, []byte](AssetFileColumns.Sha256),
		SizeBytes: sqlite.Where[Q, int64](AssetFileColumns.SizeBytes),
		CreatedBy: sqlite.Where[Q, int64](AssetFileColumns.CreatedBy),
		CreatedAt: sqlite.Where[Q, types.SQLiteDatetime](AssetFileColumns.CreatedAt),
		UpdatedAt: sqlite.Where[Q, types.SQLiteDatetime](AssetFileColumns.UpdatedAt),
	}
}

// FindAssetFile retrieves a single record by primary key
// If cols is empty Find will return all columns.
func FindAssetFile(ctx context.Context, exec bob.Executor, IDPK int64, cols ...string) (*AssetFile, error) {
	if len(cols) == 0 {
		return AssetFiles.Query(
			ctx, exec,
			SelectWhere.AssetFiles.ID.EQ(IDPK),
		).One()
	}

	return AssetFiles.Query(
		ctx, exec,
		SelectWhere.AssetFiles.ID.EQ(IDPK),
		sm.Columns(AssetFiles.Columns().Only(cols...)),
	).One()
}

// AssetFileExists checks the presence of a single record by primary key
func AssetFileExists(ctx context.Context, exec bob.Executor, IDPK int64) (bool, error) {
	return AssetFiles.Query(
		ctx, exec,
		SelectWhere.AssetFiles.ID.EQ(IDPK),
	).Exists()
}

// PrimaryKeyVals returns the primary key values of the AssetFile
func (o *AssetFile) PrimaryKeyVals() bob.Expression {
	return sqlite.Arg(o.ID)
}

// Update uses an executor to update the AssetFile
func (o *AssetFile) Update(ctx context.Context, exec bob.Executor, s *AssetFileSetter) error {
	return AssetFiles.Update(ctx, exec, s, o)
}

// Delete deletes a single AssetFile record with an executor
func (o *AssetFile) Delete(ctx context.Context, exec bob.Executor) error {
	return AssetFiles.Delete(ctx, exec, o)
}

// Reload refreshes the AssetFile using the executor
func (o *AssetFile) Reload(ctx context.Context, exec bob.Executor) error {
	o2, err := AssetFiles.Query(
		ctx, exec,
		SelectWhere.AssetFiles.ID.EQ(o.ID),
	).One()
	if err != nil {
		return err
	}
	o2.R = o.R
	*o = *o2

	return nil
}

func (o AssetFileSlice) UpdateAll(ctx context.Context, exec bob.Executor, vals AssetFileSetter) error {
	return AssetFiles.Update(ctx, exec, &vals, o...)
}

func (o AssetFileSlice) DeleteAll(ctx context.Context, exec bob.Executor) error {
	return AssetFiles.Delete(ctx, exec, o...)
}

func (o AssetFileSlice) ReloadAll(ctx context.Context, exec bob.Executor) error {
	var mods []bob.Mod[*dialect.SelectQuery]

	IDPK := make([]int64, len(o))

	for i, o := range o {
		IDPK[i] = o.ID
	}

	mods = append(mods,
		SelectWhere.AssetFiles.ID.In(IDPK...),
	)

	o2, err := AssetFiles.Query(ctx, exec, mods...).All()
	if err != nil {
		return err
	}

	for _, old := range o {
		for _, new := range o2 {
			if new.ID != old.ID {
				continue
			}
			new.R = old.R
			*old = *new
			break
		}
	}

	return nil
}

func assetFilesJoinCreatedByUser[Q dialect.Joinable](ctx context.Context, typ string) bob.Mod[Q] {
	return mods.QueryMods[Q]{
		dialect.Join[Q](typ, Users.Name(ctx)).On(
			UserColumns.ID.EQ(AssetFileColumns.CreatedBy),
		),
	}
}

// CreatedByUser starts a query for related objects on users
func (o *AssetFile) CreatedByUser(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) UsersQuery {
	return Users.Query(ctx, exec, append(mods,
		sm.Where(UserColumns.ID.EQ(sqlite.Arg(o.CreatedBy))),
	)...)
}

func (os AssetFileSlice) CreatedByUser(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) UsersQuery {
	PKArgs := make([]bob.Expression, len(os))
	for i, o := range os {
		PKArgs[i] = sqlite.ArgGroup(o.CreatedBy)
	}

	return Users.Query(ctx, exec, append(mods,
		sm.Where(sqlite.Group(UserColumns.ID).In(PKArgs...)),
	)...)
}

func (o *AssetFile) Preload(name string, retrieved any) error {
	if o == nil {
		return nil
	}

	switch name {
	case "CreatedByUser":
		rel, ok := retrieved.(*User)
		if !ok {
			return fmt.Errorf("assetFile cannot load %T as %q", retrieved, name)
		}

		o.R.CreatedByUser = rel

		if rel != nil {
			rel.R.CreatedByAssetFiles = AssetFileSlice{o}
		}
		return nil
	default:
		return fmt.Errorf("assetFile has no relationship %q", name)
	}
}

func PreloadAssetFileCreatedByUser(opts ...sqlite.PreloadOption) sqlite.Preloader {
	return sqlite.Preload[*User, UserSlice](orm.Relationship{
		Name: "CreatedByUser",
		Sides: []orm.RelSide{
			{
				From: "asset_files",
				To:   TableNames.Users,
				ToExpr: func(ctx context.Context) bob.Expression {
					return Users.Name(ctx)
				},
				FromColumns: []string{
					ColumnNames.AssetFiles.CreatedBy,
				},
				ToColumns: []string{
					ColumnNames.Users.ID,
				},
			},
		},
	}, Users.Columns().Names(), opts...)
}

func ThenLoadAssetFileCreatedByUser(queryMods ...bob.Mod[*dialect.SelectQuery]) sqlite.Loader {
	return sqlite.Loader(func(ctx context.Context, exec bob.Executor, retrieved any) error {
		loader, isLoader := retrieved.(interface {
			LoadAssetFileCreatedByUser(context.Context, bob.Executor, ...bob.Mod[*dialect.SelectQuery]) error
		})
		if !isLoader {
			return fmt.Errorf("object %T cannot load AssetFileCreatedByUser", retrieved)
		}

		err := loader.LoadAssetFileCreatedByUser(ctx, exec, queryMods...)

		// Don't cause an issue due to missing relationships
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		return err
	})
}

// LoadAssetFileCreatedByUser loads the assetFile's CreatedByUser into the .R struct
func (o *AssetFile) LoadAssetFileCreatedByUser(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if o == nil {
		return nil
	}

	// Reset the relationship
	o.R.CreatedByUser = nil

	related, err := o.CreatedByUser(ctx, exec, mods...).One()
	if err != nil {
		return err
	}

	related.R.CreatedByAssetFiles = AssetFileSlice{o}

	o.R.CreatedByUser = related
	return nil
}

// LoadAssetFileCreatedByUser loads the assetFile's CreatedByUser into the .R struct
func (os AssetFileSlice) LoadAssetFileCreatedByUser(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if len(os) == 0 {
		return nil
	}

	users, err := os.CreatedByUser(ctx, exec, mods...).All()
	if err != nil {
		return err
	}

	for _, o := range os {
		for _, rel := range users {
			if o.CreatedBy != rel.ID {
				continue
			}

			rel.R.CreatedByAssetFiles = append(rel.R.CreatedByAssetFiles, o)

			o.R.CreatedByUser = rel
			break
		}
	}

	return nil
}

func attachAssetFileCreatedByUser0(ctx context.Context, exec bob.Executor, assetFile0 *AssetFile, user1 *User) error {
	setter := &AssetFileSetter{
		CreatedBy: omit.From(user1.ID),
	}

	err := AssetFiles.Update(ctx, exec, setter, assetFile0)
	if err != nil {
		return fmt.Errorf("attachAssetFileCreatedByUser0: %w", err)
	}

	return nil
}

func (assetFile0 *AssetFile) InsertCreatedByUser(ctx context.Context, exec bob.Executor, related *UserSetter) error {
	user1, err := Users.Insert(ctx, exec, related)
	if err != nil {
		return fmt.Errorf("inserting related objects: %w", err)
	}

	err = attachAssetFileCreatedByUser0(ctx, exec, assetFile0, user1)
	if err != nil {
		return err
	}

	assetFile0.R.CreatedByUser = user1

	user1.R.CreatedByAssetFiles = append(user1.R.CreatedByAssetFiles, assetFile0)

	return nil
}

func (assetFile0 *AssetFile) AttachCreatedByUser(ctx context.Context, exec bob.Executor, user1 *User) error {
	var err error

	err = attachAssetFileCreatedByUser0(ctx, exec, assetFile0, user1)
	if err != nil {
		return err
	}

	assetFile0.R.CreatedByUser = user1

	user1.R.CreatedByAssetFiles = append(user1.R.CreatedByAssetFiles, assetFile0)

	return nil
}
