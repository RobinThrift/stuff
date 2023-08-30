// Code generated by BobGen sqlite v0.22.0. DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package factory

import (
	"context"

	"github.com/aarondl/opt/omit"
	"github.com/jaswdr/faker"
	models "github.com/kodeshack/stuff/storage/database/sqlite/models"
	"github.com/kodeshack/stuff/storage/database/sqlite/types"
	"github.com/stephenafamo/bob"
)

type SessionMod interface {
	Apply(*SessionTemplate)
}

type SessionModFunc func(*SessionTemplate)

func (f SessionModFunc) Apply(n *SessionTemplate) {
	f(n)
}

type SessionModSlice []SessionMod

func (mods SessionModSlice) Apply(n *SessionTemplate) {
	for _, f := range mods {
		f.Apply(n)
	}
}

// SessionTemplate is an object representing the database table.
// all columns are optional and should be set by mods
type SessionTemplate struct {
	ID        func() int64
	Token     func() string
	Data      func() []byte
	ExpiresAt func() types.SQLiteDatetime

	f *Factory
}

// Apply mods to the SessionTemplate
func (o *SessionTemplate) Apply(mods ...SessionMod) {
	for _, mod := range mods {
		mod.Apply(o)
	}
}

// toModel returns an *models.Session
// this does nothing with the relationship templates
func (o SessionTemplate) toModel() *models.Session {
	m := &models.Session{}

	if o.ID != nil {
		m.ID = o.ID()
	}
	if o.Token != nil {
		m.Token = o.Token()
	}
	if o.Data != nil {
		m.Data = o.Data()
	}
	if o.ExpiresAt != nil {
		m.ExpiresAt = o.ExpiresAt()
	}

	return m
}

// toModels returns an models.SessionSlice
// this does nothing with the relationship templates
func (o SessionTemplate) toModels(number int) models.SessionSlice {
	m := make(models.SessionSlice, number)

	for i := range m {
		m[i] = o.toModel()
	}

	return m
}

// setModelRels creates and sets the relationships on *models.Session
// according to the relationships in the template. Nothing is inserted into the db
func (t SessionTemplate) setModelRels(o *models.Session) {}

// BuildSetter returns an *models.SessionSetter
// this does nothing with the relationship templates
func (o SessionTemplate) BuildSetter() *models.SessionSetter {
	m := &models.SessionSetter{}

	if o.ID != nil {
		m.ID = omit.From(o.ID())
	}
	if o.Token != nil {
		m.Token = omit.From(o.Token())
	}
	if o.Data != nil {
		m.Data = omit.From(o.Data())
	}
	if o.ExpiresAt != nil {
		m.ExpiresAt = omit.From(o.ExpiresAt())
	}

	return m
}

// BuildManySetter returns an []*models.SessionSetter
// this does nothing with the relationship templates
func (o SessionTemplate) BuildManySetter(number int) []*models.SessionSetter {
	m := make([]*models.SessionSetter, number)

	for i := range m {
		m[i] = o.BuildSetter()
	}

	return m
}

// Build returns an *models.Session
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use SessionTemplate.Create
func (o SessionTemplate) Build() *models.Session {
	m := o.toModel()
	o.setModelRels(m)

	return m
}

// BuildMany returns an models.SessionSlice
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use SessionTemplate.CreateMany
func (o SessionTemplate) BuildMany(number int) models.SessionSlice {
	m := make(models.SessionSlice, number)

	for i := range m {
		m[i] = o.Build()
	}

	return m
}

func ensureCreatableSession(m *models.SessionSetter) {
	if m.Token.IsUnset() {
		m.Token = omit.From(random[string](nil))
	}
	if m.Data.IsUnset() {
		m.Data = omit.From(random[[]byte](nil))
	}
	if m.ExpiresAt.IsUnset() {
		m.ExpiresAt = omit.From(random[types.SQLiteDatetime](nil))
	}
}

// insertOptRels creates and inserts any optional the relationships on *models.Session
// according to the relationships in the template.
// any required relationship should have already exist on the model
func (o *SessionTemplate) insertOptRels(ctx context.Context, exec bob.Executor, m *models.Session) (context.Context, error) {
	var err error

	return ctx, err
}

// Create builds a session and inserts it into the database
// Relations objects are also inserted and placed in the .R field
func (o *SessionTemplate) Create(ctx context.Context, exec bob.Executor) (*models.Session, error) {
	_, m, err := o.create(ctx, exec)
	return m, err
}

// create builds a session and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// this returns a context that includes the newly inserted model
func (o *SessionTemplate) create(ctx context.Context, exec bob.Executor) (context.Context, *models.Session, error) {
	var err error
	opt := o.BuildSetter()
	ensureCreatableSession(opt)

	m, err := models.Sessions.Insert(ctx, exec, opt)
	if err != nil {
		return ctx, nil, err
	}
	ctx = sessionCtx.WithValue(ctx, m)

	ctx, err = o.insertOptRels(ctx, exec, m)
	return ctx, m, err
}

// CreateMany builds multiple sessions and inserts them into the database
// Relations objects are also inserted and placed in the .R field
func (o SessionTemplate) CreateMany(ctx context.Context, exec bob.Executor, number int) (models.SessionSlice, error) {
	_, m, err := o.createMany(ctx, exec, number)
	return m, err
}

// createMany builds multiple sessions and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// this returns a context that includes the newly inserted models
func (o SessionTemplate) createMany(ctx context.Context, exec bob.Executor, number int) (context.Context, models.SessionSlice, error) {
	var err error
	m := make(models.SessionSlice, number)

	for i := range m {
		ctx, m[i], err = o.create(ctx, exec)
		if err != nil {
			return ctx, nil, err
		}
	}

	return ctx, m, nil
}

// Session has methods that act as mods for the SessionTemplate
var SessionMods sessionMods

type sessionMods struct{}

func (m sessionMods) RandomizeAllColumns(f *faker.Faker) SessionMod {
	return SessionModSlice{
		SessionMods.RandomID(f),
		SessionMods.RandomToken(f),
		SessionMods.RandomData(f),
		SessionMods.RandomExpiresAt(f),
	}
}

// Set the model columns to this value
func (m sessionMods) ID(val int64) SessionMod {
	return SessionModFunc(func(o *SessionTemplate) {
		o.ID = func() int64 { return val }
	})
}

// Set the Column from the function
func (m sessionMods) IDFunc(f func() int64) SessionMod {
	return SessionModFunc(func(o *SessionTemplate) {
		o.ID = f
	})
}

// Clear any values for the column
func (m sessionMods) UnsetID() SessionMod {
	return SessionModFunc(func(o *SessionTemplate) {
		o.ID = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m sessionMods) RandomID(f *faker.Faker) SessionMod {
	return SessionModFunc(func(o *SessionTemplate) {
		o.ID = func() int64 {
			return random[int64](f)
		}
	})
}

func (m sessionMods) ensureID(f *faker.Faker) SessionMod {
	return SessionModFunc(func(o *SessionTemplate) {
		if o.ID != nil {
			return
		}

		o.ID = func() int64 {
			return random[int64](f)
		}
	})
}

// Set the model columns to this value
func (m sessionMods) Token(val string) SessionMod {
	return SessionModFunc(func(o *SessionTemplate) {
		o.Token = func() string { return val }
	})
}

// Set the Column from the function
func (m sessionMods) TokenFunc(f func() string) SessionMod {
	return SessionModFunc(func(o *SessionTemplate) {
		o.Token = f
	})
}

// Clear any values for the column
func (m sessionMods) UnsetToken() SessionMod {
	return SessionModFunc(func(o *SessionTemplate) {
		o.Token = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m sessionMods) RandomToken(f *faker.Faker) SessionMod {
	return SessionModFunc(func(o *SessionTemplate) {
		o.Token = func() string {
			return random[string](f)
		}
	})
}

func (m sessionMods) ensureToken(f *faker.Faker) SessionMod {
	return SessionModFunc(func(o *SessionTemplate) {
		if o.Token != nil {
			return
		}

		o.Token = func() string {
			return random[string](f)
		}
	})
}

// Set the model columns to this value
func (m sessionMods) Data(val []byte) SessionMod {
	return SessionModFunc(func(o *SessionTemplate) {
		o.Data = func() []byte { return val }
	})
}

// Set the Column from the function
func (m sessionMods) DataFunc(f func() []byte) SessionMod {
	return SessionModFunc(func(o *SessionTemplate) {
		o.Data = f
	})
}

// Clear any values for the column
func (m sessionMods) UnsetData() SessionMod {
	return SessionModFunc(func(o *SessionTemplate) {
		o.Data = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m sessionMods) RandomData(f *faker.Faker) SessionMod {
	return SessionModFunc(func(o *SessionTemplate) {
		o.Data = func() []byte {
			return random[[]byte](f)
		}
	})
}

func (m sessionMods) ensureData(f *faker.Faker) SessionMod {
	return SessionModFunc(func(o *SessionTemplate) {
		if o.Data != nil {
			return
		}

		o.Data = func() []byte {
			return random[[]byte](f)
		}
	})
}

// Set the model columns to this value
func (m sessionMods) ExpiresAt(val types.SQLiteDatetime) SessionMod {
	return SessionModFunc(func(o *SessionTemplate) {
		o.ExpiresAt = func() types.SQLiteDatetime { return val }
	})
}

// Set the Column from the function
func (m sessionMods) ExpiresAtFunc(f func() types.SQLiteDatetime) SessionMod {
	return SessionModFunc(func(o *SessionTemplate) {
		o.ExpiresAt = f
	})
}

// Clear any values for the column
func (m sessionMods) UnsetExpiresAt() SessionMod {
	return SessionModFunc(func(o *SessionTemplate) {
		o.ExpiresAt = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m sessionMods) RandomExpiresAt(f *faker.Faker) SessionMod {
	return SessionModFunc(func(o *SessionTemplate) {
		o.ExpiresAt = func() types.SQLiteDatetime {
			return random[types.SQLiteDatetime](f)
		}
	})
}

func (m sessionMods) ensureExpiresAt(f *faker.Faker) SessionMod {
	return SessionModFunc(func(o *SessionTemplate) {
		if o.ExpiresAt != nil {
			return
		}

		o.ExpiresAt = func() types.SQLiteDatetime {
			return random[types.SQLiteDatetime](f)
		}
	})
}