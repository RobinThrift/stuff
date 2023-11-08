package database

type ListTagsQuery struct {
	Search   string
	InUse    *bool
	Page     int
	PageSize int
	OrderBy  string
	OrderDir string
}

type ListAssetsQuery struct {
	SearchRaw    string
	SearchFields map[string]string

	IDs []int64

	Page     int
	PageSize int

	OrderBy  string
	OrderDir string

	AssetType string

	IncludePurchases bool
	IncludeParts     bool
	IncludeFiles     bool
	IncludeParent    bool
	IncludeChildren  bool
}

type GetAssetQuery struct {
	ID               int64
	Tag              string
	IncludePurchases bool
	IncludeParts     bool
	IncludeFiles     bool
	IncludeParent    bool
	IncludeChildren  bool
}

type ListCategoriesQuery struct {
	Search   string
	Page     int
	PageSize int
}

type ListLocationsQuery struct {
	Search   string
	Page     int
	PageSize int
}

type ListPositionCodesQuery struct {
	Search   string
	Page     int
	PageSize int
}

type ListCustomAttrsQuery struct {
	Search   string
	Page     int
	PageSize int
}

type ListFilesQuery struct {
	AssetID  int64
	Page     int
	PageSize int
	Hashes   [][]byte
}

type ListManufacturersQuery struct {
	Search   string
	Page     int
	PageSize int
}

type ListModelsQuery struct {
	Search   string
	Page     int
	PageSize int
}

type ListSuppliersQuery struct {
	Search   string
	Page     int
	PageSize int
}

type ListUsersQuery struct {
	Search   string
	Page     int
	PageSize int
	OrderBy  string
	OrderDir string
}
