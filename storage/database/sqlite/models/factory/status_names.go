// Code generated by BobGen sqlite v0.22.0. DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package factory

import (
	"github.com/aarondl/opt/null"
	"github.com/jaswdr/faker"
	models "github.com/kodeshack/stuff/storage/database/sqlite/models"
)

type StatusNameMod interface {
	Apply(*StatusNameTemplate)
}

type StatusNameModFunc func(*StatusNameTemplate)

func (f StatusNameModFunc) Apply(n *StatusNameTemplate) {
	f(n)
}

type StatusNameModSlice []StatusNameMod

func (mods StatusNameModSlice) Apply(n *StatusNameTemplate) {
	for _, f := range mods {
		f.Apply(n)
	}
}

// StatusNameTemplate is an object representing the database table.
// all columns are optional and should be set by mods
type StatusNameTemplate struct {
	Name func() null.Val[string]

	f *Factory
}

// Apply mods to the StatusNameTemplate
func (o *StatusNameTemplate) Apply(mods ...StatusNameMod) {
	for _, mod := range mods {
		mod.Apply(o)
	}
}

// toModel returns an *models.StatusName
// this does nothing with the relationship templates
func (o StatusNameTemplate) toModel() *models.StatusName {
	m := &models.StatusName{}

	if o.Name != nil {
		m.Name = o.Name()
	}

	return m
}

// toModels returns an models.StatusNameSlice
// this does nothing with the relationship templates
func (o StatusNameTemplate) toModels(number int) models.StatusNameSlice {
	m := make(models.StatusNameSlice, number)

	for i := range m {
		m[i] = o.toModel()
	}

	return m
}

// setModelRels creates and sets the relationships on *models.StatusName
// according to the relationships in the template. Nothing is inserted into the db
func (t StatusNameTemplate) setModelRels(o *models.StatusName) {}

// Build returns an *models.StatusName
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use StatusNameTemplate.Create
func (o StatusNameTemplate) Build() *models.StatusName {
	m := o.toModel()
	o.setModelRels(m)

	return m
}

// BuildMany returns an models.StatusNameSlice
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use StatusNameTemplate.CreateMany
func (o StatusNameTemplate) BuildMany(number int) models.StatusNameSlice {
	m := make(models.StatusNameSlice, number)

	for i := range m {
		m[i] = o.Build()
	}

	return m
}

// StatusName has methods that act as mods for the StatusNameTemplate
var StatusNameMods statusNameMods

type statusNameMods struct{}

func (m statusNameMods) RandomizeAllColumns(f *faker.Faker) StatusNameMod {
	return StatusNameModSlice{
		StatusNameMods.RandomName(f),
	}
}

// Set the model columns to this value
func (m statusNameMods) Name(val null.Val[string]) StatusNameMod {
	return StatusNameModFunc(func(o *StatusNameTemplate) {
		o.Name = func() null.Val[string] { return val }
	})
}

// Set the Column from the function
func (m statusNameMods) NameFunc(f func() null.Val[string]) StatusNameMod {
	return StatusNameModFunc(func(o *StatusNameTemplate) {
		o.Name = f
	})
}

// Clear any values for the column
func (m statusNameMods) UnsetName() StatusNameMod {
	return StatusNameModFunc(func(o *StatusNameTemplate) {
		o.Name = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m statusNameMods) RandomName(f *faker.Faker) StatusNameMod {
	return StatusNameModFunc(func(o *StatusNameTemplate) {
		o.Name = func() null.Val[string] {
			return randomNull[string](f)
		}
	})
}

func (m statusNameMods) ensureName(f *faker.Faker) StatusNameMod {
	return StatusNameModFunc(func(o *StatusNameTemplate) {
		if o.Name != nil {
			return
		}

		o.Name = func() null.Val[string] {
			return randomNull[string](f)
		}
	})
}
