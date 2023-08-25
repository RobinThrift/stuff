// Code generated by BobGen sqlite v0.22.0. DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"github.com/aarondl/opt/null"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/sqlite"
)

// StorageLocation is an object representing the database table.
type StorageLocation struct {
	Name null.Val[string] `db:"name" `
}

// StorageLocationSlice is an alias for a slice of pointers to StorageLocation.
// This should almost always be used instead of []*StorageLocation.
type StorageLocationSlice []*StorageLocation

// StorageLocations contains methods to work with the storage_locations view
var StorageLocations = sqlite.NewViewx[*StorageLocation, StorageLocationSlice]("", "storage_locations")

// StorageLocationsQuery is a query on the storage_locations view
type StorageLocationsQuery = *sqlite.ViewQuery[*StorageLocation, StorageLocationSlice]

// StorageLocationsStmt is a prepared statment on storage_locations
type StorageLocationsStmt = bob.QueryStmt[*StorageLocation, StorageLocationSlice]

type storageLocationColumnNames struct {
	Name string
}

var StorageLocationColumns = struct {
	Name sqlite.Expression
}{
	Name: sqlite.Quote("storage_locations", "name"),
}

type storageLocationWhere[Q sqlite.Filterable] struct {
	Name sqlite.WhereNullMod[Q, string]
}

func StorageLocationWhere[Q sqlite.Filterable]() storageLocationWhere[Q] {
	return storageLocationWhere[Q]{
		Name: sqlite.WhereNull[Q, string](StorageLocationColumns.Name),
	}
}
