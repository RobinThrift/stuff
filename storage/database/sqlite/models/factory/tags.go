// Code generated by BobGen sqlite v0.22.0. DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package factory

import (
	"context"

	"github.com/aarondl/opt/null"
	"github.com/aarondl/opt/omit"
	"github.com/jaswdr/faker"
	models "github.com/kodeshack/stuff/storage/database/sqlite/models"
	"github.com/kodeshack/stuff/storage/database/sqlite/types"
	"github.com/stephenafamo/bob"
)

type TagMod interface {
	Apply(*TagTemplate)
}

type TagModFunc func(*TagTemplate)

func (f TagModFunc) Apply(n *TagTemplate) {
	f(n)
}

type TagModSlice []TagMod

func (mods TagModSlice) Apply(n *TagTemplate) {
	for _, f := range mods {
		f.Apply(n)
	}
}

// TagTemplate is an object representing the database table.
// all columns are optional and should be set by mods
type TagTemplate struct {
	ID        func() int64
	Tag       func() string
	CreatedAt func() types.SQLiteDatetime
	UpdatedAt func() types.SQLiteDatetime

	r tagR
	f *Factory
}

type tagR struct {
	Assets []*tagRAssetsR
}

type tagRAssetsR struct {
	number int
	o      *AssetTemplate
}

// Apply mods to the TagTemplate
func (o *TagTemplate) Apply(mods ...TagMod) {
	for _, mod := range mods {
		mod.Apply(o)
	}
}

// toModel returns an *models.Tag
// this does nothing with the relationship templates
func (o TagTemplate) toModel() *models.Tag {
	m := &models.Tag{}

	if o.ID != nil {
		m.ID = o.ID()
	}
	if o.Tag != nil {
		m.Tag = o.Tag()
	}
	if o.CreatedAt != nil {
		m.CreatedAt = o.CreatedAt()
	}
	if o.UpdatedAt != nil {
		m.UpdatedAt = o.UpdatedAt()
	}

	return m
}

// toModels returns an models.TagSlice
// this does nothing with the relationship templates
func (o TagTemplate) toModels(number int) models.TagSlice {
	m := make(models.TagSlice, number)

	for i := range m {
		m[i] = o.toModel()
	}

	return m
}

// setModelRels creates and sets the relationships on *models.Tag
// according to the relationships in the template. Nothing is inserted into the db
func (t TagTemplate) setModelRels(o *models.Tag) {
	if t.r.Assets != nil {
		rel := models.AssetSlice{}
		for _, r := range t.r.Assets {
			related := r.o.toModels(r.number)
			for _, rel := range related {
				rel.TagID = null.From(o.ID)
				rel.R.Tag = o
			}
			rel = append(rel, related...)
		}
		o.R.Assets = rel
	}

}

// BuildSetter returns an *models.TagSetter
// this does nothing with the relationship templates
func (o TagTemplate) BuildSetter() *models.TagSetter {
	m := &models.TagSetter{}

	if o.ID != nil {
		m.ID = omit.From(o.ID())
	}
	if o.Tag != nil {
		m.Tag = omit.From(o.Tag())
	}
	if o.CreatedAt != nil {
		m.CreatedAt = omit.From(o.CreatedAt())
	}
	if o.UpdatedAt != nil {
		m.UpdatedAt = omit.From(o.UpdatedAt())
	}

	return m
}

// BuildManySetter returns an []*models.TagSetter
// this does nothing with the relationship templates
func (o TagTemplate) BuildManySetter(number int) []*models.TagSetter {
	m := make([]*models.TagSetter, number)

	for i := range m {
		m[i] = o.BuildSetter()
	}

	return m
}

// Build returns an *models.Tag
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use TagTemplate.Create
func (o TagTemplate) Build() *models.Tag {
	m := o.toModel()
	o.setModelRels(m)

	return m
}

// BuildMany returns an models.TagSlice
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use TagTemplate.CreateMany
func (o TagTemplate) BuildMany(number int) models.TagSlice {
	m := make(models.TagSlice, number)

	for i := range m {
		m[i] = o.Build()
	}

	return m
}

func ensureCreatableTag(m *models.TagSetter) {
	if m.Tag.IsUnset() {
		m.Tag = omit.From(random[string](nil))
	}
}

// insertOptRels creates and inserts any optional the relationships on *models.Tag
// according to the relationships in the template.
// any required relationship should have already exist on the model
func (o *TagTemplate) insertOptRels(ctx context.Context, exec bob.Executor, m *models.Tag) (context.Context, error) {
	var err error

	if o.r.Assets != nil {
		for _, r := range o.r.Assets {
			var rel0 models.AssetSlice
			ctx, rel0, err = r.o.createMany(ctx, exec, r.number)
			if err != nil {
				return ctx, err
			}

			err = m.AttachAssets(ctx, exec, rel0...)
			if err != nil {
				return ctx, err
			}
		}
	}

	return ctx, err
}

// Create builds a tag and inserts it into the database
// Relations objects are also inserted and placed in the .R field
func (o *TagTemplate) Create(ctx context.Context, exec bob.Executor) (*models.Tag, error) {
	_, m, err := o.create(ctx, exec)
	return m, err
}

// create builds a tag and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// this returns a context that includes the newly inserted model
func (o *TagTemplate) create(ctx context.Context, exec bob.Executor) (context.Context, *models.Tag, error) {
	var err error
	opt := o.BuildSetter()
	ensureCreatableTag(opt)

	m, err := models.Tags.Insert(ctx, exec, opt)
	if err != nil {
		return ctx, nil, err
	}
	ctx = tagCtx.WithValue(ctx, m)

	ctx, err = o.insertOptRels(ctx, exec, m)
	return ctx, m, err
}

// CreateMany builds multiple tags and inserts them into the database
// Relations objects are also inserted and placed in the .R field
func (o TagTemplate) CreateMany(ctx context.Context, exec bob.Executor, number int) (models.TagSlice, error) {
	_, m, err := o.createMany(ctx, exec, number)
	return m, err
}

// createMany builds multiple tags and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// this returns a context that includes the newly inserted models
func (o TagTemplate) createMany(ctx context.Context, exec bob.Executor, number int) (context.Context, models.TagSlice, error) {
	var err error
	m := make(models.TagSlice, number)

	for i := range m {
		ctx, m[i], err = o.create(ctx, exec)
		if err != nil {
			return ctx, nil, err
		}
	}

	return ctx, m, nil
}

// Tag has methods that act as mods for the TagTemplate
var TagMods tagMods

type tagMods struct{}

func (m tagMods) RandomizeAllColumns(f *faker.Faker) TagMod {
	return TagModSlice{
		TagMods.RandomID(f),
		TagMods.RandomTag(f),
		TagMods.RandomCreatedAt(f),
		TagMods.RandomUpdatedAt(f),
	}
}

// Set the model columns to this value
func (m tagMods) ID(val int64) TagMod {
	return TagModFunc(func(o *TagTemplate) {
		o.ID = func() int64 { return val }
	})
}

// Set the Column from the function
func (m tagMods) IDFunc(f func() int64) TagMod {
	return TagModFunc(func(o *TagTemplate) {
		o.ID = f
	})
}

// Clear any values for the column
func (m tagMods) UnsetID() TagMod {
	return TagModFunc(func(o *TagTemplate) {
		o.ID = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m tagMods) RandomID(f *faker.Faker) TagMod {
	return TagModFunc(func(o *TagTemplate) {
		o.ID = func() int64 {
			return random[int64](f)
		}
	})
}

func (m tagMods) ensureID(f *faker.Faker) TagMod {
	return TagModFunc(func(o *TagTemplate) {
		if o.ID != nil {
			return
		}

		o.ID = func() int64 {
			return random[int64](f)
		}
	})
}

// Set the model columns to this value
func (m tagMods) Tag(val string) TagMod {
	return TagModFunc(func(o *TagTemplate) {
		o.Tag = func() string { return val }
	})
}

// Set the Column from the function
func (m tagMods) TagFunc(f func() string) TagMod {
	return TagModFunc(func(o *TagTemplate) {
		o.Tag = f
	})
}

// Clear any values for the column
func (m tagMods) UnsetTag() TagMod {
	return TagModFunc(func(o *TagTemplate) {
		o.Tag = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m tagMods) RandomTag(f *faker.Faker) TagMod {
	return TagModFunc(func(o *TagTemplate) {
		o.Tag = func() string {
			return random[string](f)
		}
	})
}

func (m tagMods) ensureTag(f *faker.Faker) TagMod {
	return TagModFunc(func(o *TagTemplate) {
		if o.Tag != nil {
			return
		}

		o.Tag = func() string {
			return random[string](f)
		}
	})
}

// Set the model columns to this value
func (m tagMods) CreatedAt(val types.SQLiteDatetime) TagMod {
	return TagModFunc(func(o *TagTemplate) {
		o.CreatedAt = func() types.SQLiteDatetime { return val }
	})
}

// Set the Column from the function
func (m tagMods) CreatedAtFunc(f func() types.SQLiteDatetime) TagMod {
	return TagModFunc(func(o *TagTemplate) {
		o.CreatedAt = f
	})
}

// Clear any values for the column
func (m tagMods) UnsetCreatedAt() TagMod {
	return TagModFunc(func(o *TagTemplate) {
		o.CreatedAt = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m tagMods) RandomCreatedAt(f *faker.Faker) TagMod {
	return TagModFunc(func(o *TagTemplate) {
		o.CreatedAt = func() types.SQLiteDatetime {
			return random[types.SQLiteDatetime](f)
		}
	})
}

func (m tagMods) ensureCreatedAt(f *faker.Faker) TagMod {
	return TagModFunc(func(o *TagTemplate) {
		if o.CreatedAt != nil {
			return
		}

		o.CreatedAt = func() types.SQLiteDatetime {
			return random[types.SQLiteDatetime](f)
		}
	})
}

// Set the model columns to this value
func (m tagMods) UpdatedAt(val types.SQLiteDatetime) TagMod {
	return TagModFunc(func(o *TagTemplate) {
		o.UpdatedAt = func() types.SQLiteDatetime { return val }
	})
}

// Set the Column from the function
func (m tagMods) UpdatedAtFunc(f func() types.SQLiteDatetime) TagMod {
	return TagModFunc(func(o *TagTemplate) {
		o.UpdatedAt = f
	})
}

// Clear any values for the column
func (m tagMods) UnsetUpdatedAt() TagMod {
	return TagModFunc(func(o *TagTemplate) {
		o.UpdatedAt = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m tagMods) RandomUpdatedAt(f *faker.Faker) TagMod {
	return TagModFunc(func(o *TagTemplate) {
		o.UpdatedAt = func() types.SQLiteDatetime {
			return random[types.SQLiteDatetime](f)
		}
	})
}

func (m tagMods) ensureUpdatedAt(f *faker.Faker) TagMod {
	return TagModFunc(func(o *TagTemplate) {
		if o.UpdatedAt != nil {
			return
		}

		o.UpdatedAt = func() types.SQLiteDatetime {
			return random[types.SQLiteDatetime](f)
		}
	})
}

func (m tagMods) WithAssets(number int, related *AssetTemplate) TagMod {
	return TagModFunc(func(o *TagTemplate) {
		o.r.Assets = []*tagRAssetsR{{
			number: number,
			o:      related,
		}}
	})
}

func (m tagMods) WithNewAssets(number int, mods ...AssetMod) TagMod {
	return TagModFunc(func(o *TagTemplate) {

		related := o.f.NewAsset(mods...)
		m.WithAssets(number, related).Apply(o)
	})
}

func (m tagMods) AddAssets(number int, related *AssetTemplate) TagMod {
	return TagModFunc(func(o *TagTemplate) {
		o.r.Assets = append(o.r.Assets, &tagRAssetsR{
			number: number,
			o:      related,
		})
	})
}

func (m tagMods) AddNewAssets(number int, mods ...AssetMod) TagMod {
	return TagModFunc(func(o *TagTemplate) {

		related := o.f.NewAsset(mods...)
		m.AddAssets(number, related).Apply(o)
	})
}

func (m tagMods) WithoutAssets() TagMod {
	return TagModFunc(func(o *TagTemplate) {
		o.r.Assets = nil
	})
}
