// Code generated by BobGen sqlite v0.22.0. DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/RobinThrift/stuff/storage/database/sqlite/types"
	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/clause"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/dialect"
	"github.com/stephenafamo/bob/dialect/sqlite/im"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
	"github.com/stephenafamo/bob/dialect/sqlite/um"
	"github.com/stephenafamo/bob/mods"
)

// Tag is an object representing the database table.
type Tag struct {
	ID        int64                `db:"id,pk" `
	Tag       string               `db:"tag" `
	InUse     bool                 `db:"in_use" `
	CreatedAt types.SQLiteDatetime `db:"created_at" `
	UpdatedAt types.SQLiteDatetime `db:"updated_at" `

	R tagR `db:"-" `
}

// TagSlice is an alias for a slice of pointers to Tag.
// This should almost always be used instead of []*Tag.
type TagSlice []*Tag

// Tags contains methods to work with the tags table
var Tags = sqlite.NewTablex[*Tag, TagSlice, *TagSetter]("", "tags")

// TagsQuery is a query on the tags table
type TagsQuery = *sqlite.ViewQuery[*Tag, TagSlice]

// TagsStmt is a prepared statment on tags
type TagsStmt = bob.QueryStmt[*Tag, TagSlice]

// tagR is where relationships are stored.
type tagR struct {
	Assets AssetSlice // fk_assets_2
}

// TagSetter is used for insert/upsert/update operations
// All values are optional, and do not have to be set
// Generated columns are not included
type TagSetter struct {
	ID        omit.Val[int64]                `db:"id,pk"`
	Tag       omit.Val[string]               `db:"tag"`
	InUse     omit.Val[bool]                 `db:"in_use"`
	CreatedAt omit.Val[types.SQLiteDatetime] `db:"created_at"`
	UpdatedAt omit.Val[types.SQLiteDatetime] `db:"updated_at"`
}

func (s TagSetter) SetColumns() []string {
	vals := make([]string, 0, 5)
	if !s.ID.IsUnset() {
		vals = append(vals, "id")
	}

	if !s.Tag.IsUnset() {
		vals = append(vals, "tag")
	}

	if !s.InUse.IsUnset() {
		vals = append(vals, "in_use")
	}

	if !s.CreatedAt.IsUnset() {
		vals = append(vals, "created_at")
	}

	if !s.UpdatedAt.IsUnset() {
		vals = append(vals, "updated_at")
	}

	return vals
}

func (s TagSetter) Overwrite(t *Tag) {
	if !s.ID.IsUnset() {
		t.ID, _ = s.ID.Get()
	}
	if !s.Tag.IsUnset() {
		t.Tag, _ = s.Tag.Get()
	}
	if !s.InUse.IsUnset() {
		t.InUse, _ = s.InUse.Get()
	}
	if !s.CreatedAt.IsUnset() {
		t.CreatedAt, _ = s.CreatedAt.Get()
	}
	if !s.UpdatedAt.IsUnset() {
		t.UpdatedAt, _ = s.UpdatedAt.Get()
	}
}

func (s TagSetter) Apply(q *dialect.UpdateQuery) {
	if !s.ID.IsUnset() {
		um.Set("id").ToArg(s.ID).Apply(q)
	}
	if !s.Tag.IsUnset() {
		um.Set("tag").ToArg(s.Tag).Apply(q)
	}
	if !s.InUse.IsUnset() {
		um.Set("in_use").ToArg(s.InUse).Apply(q)
	}
	if !s.CreatedAt.IsUnset() {
		um.Set("created_at").ToArg(s.CreatedAt).Apply(q)
	}
	if !s.UpdatedAt.IsUnset() {
		um.Set("updated_at").ToArg(s.UpdatedAt).Apply(q)
	}
}

func (s TagSetter) Insert() bob.Mod[*dialect.InsertQuery] {
	vals := make([]bob.Expression, 0, 5)
	if !s.ID.IsUnset() {
		vals = append(vals, sqlite.Arg(s.ID))
	}

	if !s.Tag.IsUnset() {
		vals = append(vals, sqlite.Arg(s.Tag))
	}

	if !s.InUse.IsUnset() {
		vals = append(vals, sqlite.Arg(s.InUse))
	}

	if !s.CreatedAt.IsUnset() {
		vals = append(vals, sqlite.Arg(s.CreatedAt))
	}

	if !s.UpdatedAt.IsUnset() {
		vals = append(vals, sqlite.Arg(s.UpdatedAt))
	}

	return im.Values(vals...)
}

type tagColumnNames struct {
	ID        string
	Tag       string
	InUse     string
	CreatedAt string
	UpdatedAt string
}

type tagRelationshipJoins[Q dialect.Joinable] struct {
	Assets bob.Mod[Q]
}

func buildtagRelationshipJoins[Q dialect.Joinable](ctx context.Context, typ string) tagRelationshipJoins[Q] {
	return tagRelationshipJoins[Q]{
		Assets: tagsJoinAssets[Q](ctx, typ),
	}
}

func tagsJoin[Q dialect.Joinable](ctx context.Context) joinSet[tagRelationshipJoins[Q]] {
	return joinSet[tagRelationshipJoins[Q]]{
		InnerJoin: buildtagRelationshipJoins[Q](ctx, clause.InnerJoin),
		LeftJoin:  buildtagRelationshipJoins[Q](ctx, clause.LeftJoin),
		RightJoin: buildtagRelationshipJoins[Q](ctx, clause.RightJoin),
	}
}

var TagColumns = struct {
	ID        sqlite.Expression
	Tag       sqlite.Expression
	InUse     sqlite.Expression
	CreatedAt sqlite.Expression
	UpdatedAt sqlite.Expression
}{
	ID:        sqlite.Quote("tags", "id"),
	Tag:       sqlite.Quote("tags", "tag"),
	InUse:     sqlite.Quote("tags", "in_use"),
	CreatedAt: sqlite.Quote("tags", "created_at"),
	UpdatedAt: sqlite.Quote("tags", "updated_at"),
}

type tagWhere[Q sqlite.Filterable] struct {
	ID        sqlite.WhereMod[Q, int64]
	Tag       sqlite.WhereMod[Q, string]
	InUse     sqlite.WhereMod[Q, bool]
	CreatedAt sqlite.WhereMod[Q, types.SQLiteDatetime]
	UpdatedAt sqlite.WhereMod[Q, types.SQLiteDatetime]
}

func TagWhere[Q sqlite.Filterable]() tagWhere[Q] {
	return tagWhere[Q]{
		ID:        sqlite.Where[Q, int64](TagColumns.ID),
		Tag:       sqlite.Where[Q, string](TagColumns.Tag),
		InUse:     sqlite.Where[Q, bool](TagColumns.InUse),
		CreatedAt: sqlite.Where[Q, types.SQLiteDatetime](TagColumns.CreatedAt),
		UpdatedAt: sqlite.Where[Q, types.SQLiteDatetime](TagColumns.UpdatedAt),
	}
}

// FindTag retrieves a single record by primary key
// If cols is empty Find will return all columns.
func FindTag(ctx context.Context, exec bob.Executor, IDPK int64, cols ...string) (*Tag, error) {
	if len(cols) == 0 {
		return Tags.Query(
			ctx, exec,
			SelectWhere.Tags.ID.EQ(IDPK),
		).One()
	}

	return Tags.Query(
		ctx, exec,
		SelectWhere.Tags.ID.EQ(IDPK),
		sm.Columns(Tags.Columns().Only(cols...)),
	).One()
}

// TagExists checks the presence of a single record by primary key
func TagExists(ctx context.Context, exec bob.Executor, IDPK int64) (bool, error) {
	return Tags.Query(
		ctx, exec,
		SelectWhere.Tags.ID.EQ(IDPK),
	).Exists()
}

// PrimaryKeyVals returns the primary key values of the Tag
func (o *Tag) PrimaryKeyVals() bob.Expression {
	return sqlite.Arg(o.ID)
}

// Update uses an executor to update the Tag
func (o *Tag) Update(ctx context.Context, exec bob.Executor, s *TagSetter) error {
	return Tags.Update(ctx, exec, s, o)
}

// Delete deletes a single Tag record with an executor
func (o *Tag) Delete(ctx context.Context, exec bob.Executor) error {
	return Tags.Delete(ctx, exec, o)
}

// Reload refreshes the Tag using the executor
func (o *Tag) Reload(ctx context.Context, exec bob.Executor) error {
	o2, err := Tags.Query(
		ctx, exec,
		SelectWhere.Tags.ID.EQ(o.ID),
	).One()
	if err != nil {
		return err
	}
	o2.R = o.R
	*o = *o2

	return nil
}

func (o TagSlice) UpdateAll(ctx context.Context, exec bob.Executor, vals TagSetter) error {
	return Tags.Update(ctx, exec, &vals, o...)
}

func (o TagSlice) DeleteAll(ctx context.Context, exec bob.Executor) error {
	return Tags.Delete(ctx, exec, o...)
}

func (o TagSlice) ReloadAll(ctx context.Context, exec bob.Executor) error {
	var mods []bob.Mod[*dialect.SelectQuery]

	IDPK := make([]int64, len(o))

	for i, o := range o {
		IDPK[i] = o.ID
	}

	mods = append(mods,
		SelectWhere.Tags.ID.In(IDPK...),
	)

	o2, err := Tags.Query(ctx, exec, mods...).All()
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

func tagsJoinAssets[Q dialect.Joinable](ctx context.Context, typ string) bob.Mod[Q] {
	return mods.QueryMods[Q]{
		dialect.Join[Q](typ, Assets.Name(ctx)).On(
			AssetColumns.Tag.EQ(TagColumns.Tag),
		),
	}
}

// Assets starts a query for related objects on assets
func (o *Tag) Assets(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) AssetsQuery {
	return Assets.Query(ctx, exec, append(mods,
		sm.Where(AssetColumns.Tag.EQ(sqlite.Arg(o.Tag))),
	)...)
}

func (os TagSlice) Assets(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) AssetsQuery {
	PKArgs := make([]bob.Expression, len(os))
	for i, o := range os {
		PKArgs[i] = sqlite.ArgGroup(o.Tag)
	}

	return Assets.Query(ctx, exec, append(mods,
		sm.Where(sqlite.Group(AssetColumns.Tag).In(PKArgs...)),
	)...)
}

func (o *Tag) Preload(name string, retrieved any) error {
	if o == nil {
		return nil
	}

	switch name {
	case "Assets":
		rels, ok := retrieved.(AssetSlice)
		if !ok {
			return fmt.Errorf("tag cannot load %T as %q", retrieved, name)
		}

		o.R.Assets = rels

		return nil
	default:
		return fmt.Errorf("tag has no relationship %q", name)
	}
}

func ThenLoadTagAssets(queryMods ...bob.Mod[*dialect.SelectQuery]) sqlite.Loader {
	return sqlite.Loader(func(ctx context.Context, exec bob.Executor, retrieved any) error {
		loader, isLoader := retrieved.(interface {
			LoadTagAssets(context.Context, bob.Executor, ...bob.Mod[*dialect.SelectQuery]) error
		})
		if !isLoader {
			return fmt.Errorf("object %T cannot load TagAssets", retrieved)
		}

		err := loader.LoadTagAssets(ctx, exec, queryMods...)

		// Don't cause an issue due to missing relationships
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		return err
	})
}

// LoadTagAssets loads the tag's Assets into the .R struct
func (o *Tag) LoadTagAssets(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if o == nil {
		return nil
	}

	// Reset the relationship
	o.R.Assets = nil

	related, err := o.Assets(ctx, exec, mods...).All()
	if err != nil {
		return err
	}

	o.R.Assets = related
	return nil
}

// LoadTagAssets loads the tag's Assets into the .R struct
func (os TagSlice) LoadTagAssets(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if len(os) == 0 {
		return nil
	}

	assets, err := os.Assets(ctx, exec, mods...).All()
	if err != nil {
		return err
	}

	for _, o := range os {
		o.R.Assets = nil
	}

	for _, o := range os {
		for _, rel := range assets {
			if o.Tag != rel.Tag.GetOrZero() {
				continue
			}

			o.R.Assets = append(o.R.Assets, rel)
		}
	}

	return nil
}

func insertTagAssets0(ctx context.Context, exec bob.Executor, assets1 []*AssetSetter, tag0 *Tag) (AssetSlice, error) {
	for _, asset1 := range assets1 {
		asset1.Tag = omitnull.From(tag0.Tag)
	}

	ret, err := Assets.InsertMany(ctx, exec, assets1...)
	if err != nil {
		return ret, fmt.Errorf("insertTagAssets0: %w", err)
	}

	return ret, nil
}

func attachTagAssets0(ctx context.Context, exec bob.Executor, assets1 AssetSlice, tag0 *Tag) error {
	setter := &AssetSetter{
		Tag: omitnull.From(tag0.Tag),
	}

	err := Assets.Update(ctx, exec, setter, assets1...)
	if err != nil {
		return fmt.Errorf("attachTagAssets0: %w", err)
	}

	return nil
}

func (tag0 *Tag) InsertAssets(ctx context.Context, exec bob.Executor, related ...*AssetSetter) error {
	if len(related) == 0 {
		return nil
	}

	asset1, err := insertTagAssets0(ctx, exec, related, tag0)
	if err != nil {
		return err
	}

	tag0.R.Assets = append(tag0.R.Assets, asset1...)

	return nil
}

func (tag0 *Tag) AttachAssets(ctx context.Context, exec bob.Executor, related ...*Asset) error {
	if len(related) == 0 {
		return nil
	}

	var err error
	asset1 := AssetSlice(related)

	err = attachTagAssets0(ctx, exec, asset1, tag0)
	if err != nil {
		return err
	}

	tag0.R.Assets = append(tag0.R.Assets, asset1...)

	return nil
}
