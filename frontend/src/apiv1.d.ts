/**
 * This file was auto-generated by openapi-typescript.
 * Do not make direct changes to the file.
 */

export interface paths {
    "/v1/assets": {
        get: operations["ListAssets"]
        post: operations["CreateAsset"]
    }
    "/v1/assets/{tagOrID}": {
        get: operations["GetAsset"]
        put: operations["UpdateAsset"]
        delete: operations["DeleteAsset"]
        parameters: {
            path: {
                tagOrID: string
            }
        }
    }
    "/v1/tags": {
        get: operations["ListTags"]
    }
    "/v1/categories": {
        get: operations["ListCategories"]
    }
    "/v1/locations": {
        get: operations["ListLocations"]
    }
    "/v1/locations/position_codes": {
        get: operations["ListPositionCodes"]
    }
    "/v1/models": {
        get: operations["ListModels"]
    }
    "/v1/manufacturers": {
        get: operations["ListManufacturers"]
    }
    "/v1/custom_attrs": {
        get: operations["ListCustomAttrs"]
    }
    "/v1/suppliers": {
        get: operations["ListSuppliers"]
    }
    "/v1/users": {
        get: operations["ListUsers"]
    }
}

export type webhooks = Record<string, never>

export interface components {
    schemas: {
        Asset: {
            id: number
            type: string
            parentAssetID?: number
            checkedOutTo?: number
            /** @enum {string} */
            status: "IN_STORAGE" | "IN_USE" | "ARCHIVED"
            tag: string
            name: string
            category?: string
            model?: string
            modelNo?: string
            serialNo?: string
            manufacturer?: string
            notes?: string
            imageURL?: string
            thumbnailURL?: string
            /** Format: date */
            warrantyUntil?: string
            quantity?: number
            quantityUnit?: string
            customAttrs: components["schemas"]["CustomAttr"][]
            location?: string
            positionCode?: string
            purchases: components["schemas"]["Purchase"][]
            partsTotalCount?: number
            parts: components["schemas"]["AssetPart"][]
            files: components["schemas"]["AssetFile"][]
            children?: components["schemas"]["Asset"][]
            createdBy: number
            /** Format: date */
            createdAt: string
            /** Format: date */
            updatedAt: string
        }
        CustomAttr: {
            name: string
            value: unknown
        }
        Purchase: {
            supplier?: string
            orderNo?: string
            /** Format: date */
            date?: string
            amount?: number
            currency?: string
        }
        AssetPart: {
            id: number
            assetID: number
            name: string
            tag: string
            location?: string
            positionCode?: string
            notes?: string
            createdBy: number
            /** Format: date */
            createdAt: string
            /** Format: date */
            updatedAt: string
        }
        AssetFile: {
            id: number
            assetID: number
            name: string
            filetype: string
            sizeBytes: number
            publicPath: string
            sha256: string
            createdBy: number
            /** Format: date */
            createdAt: string
            /** Format: date */
            updatedAt: string
        }
        AssetListPage: {
            total: number
            numPages: number
            page: number
            pageSize: number
            assets: components["schemas"]["Asset"][]
        }
        Tag: {
            id: number
            tag: string
            inUse: boolean
            /** Format: date */
            createdAt: string
            /** Format: date */
            updatedAt: string
        }
        TagListPage: {
            total: number
            numPages: number
            page: number
            pageSize: number
            tags: components["schemas"]["Tag"][]
        }
        Category: {
            name: string
        }
        CategoryListPage: {
            total: number
            numPages: number
            page: number
            pageSize: number
            categories: components["schemas"]["Category"][]
        }
        Location: {
            name: string
        }
        LocationListPage: {
            total: number
            numPages: number
            page: number
            pageSize: number
            locations: components["schemas"]["Location"][]
        }
        PositionCode: {
            code: string
        }
        PositionCodeListPage: {
            total: number
            numPages: number
            page: number
            pageSize: number
            positionCodes: components["schemas"]["PositionCode"][]
        }
        Model: {
            name: string
            modelNo?: string
        }
        ModelListPage: {
            total: number
            numPages: number
            page: number
            pageSize: number
            models: components["schemas"]["Model"][]
        }
        Manufacturer: {
            name: string
            modelNo?: string
        }
        ManufacturerListPage: {
            total: number
            numPages: number
            page: number
            pageSize: number
            manufacturers: components["schemas"]["Manufacturer"][]
        }
        CustomAttrListPage: {
            total: number
            numPages: number
            page: number
            pageSize: number
            customAttrs: components["schemas"]["CustomAttr"][]
        }
        Supplier: {
            name: string
            modelNo?: string
        }
        SupplierListPage: {
            total: number
            numPages: number
            page: number
            pageSize: number
            suppliers: components["schemas"]["Supplier"][]
        }
        User: {
            id: number
            username: string
            displayName: string
            isAdmin: boolean
            /** Format: date */
            createdAt: string
            /** Format: date */
            updatedAt: string
        }
        UserListPage: {
            total: number
            numPages: number
            page: number
            pageSize: number
            users: components["schemas"]["User"][]
        }
        /** @description API error object that follow RFC7807 (https://datatracker.ietf.org/doc/html/rfc7807). */
        Error: {
            code: number
            type: string
            title: string
            detail: string
        }
    }
    responses: never
    parameters: never
    requestBodies: {
        /** @description The asset to be created. If not tag is set a free tag will be used or a new tag created if no free tag is available. */
        CreateAssetRequest: {
            content: {
                "application/json": {
                    type: string
                    parentAssetID?: number
                    checkedOutTo?: number
                    /** @enum {string} */
                    status: "IN_STORAGE" | "IN_USE" | "ARCHIVED"
                    tag?: string
                    name: string
                    category?: string
                    model?: string
                    modelNo?: string
                    serialNo?: string
                    manufacturer?: string
                    notes?: string
                    /** Format: date */
                    warrantyUntil?: string
                    quantity?: number
                    quantityUnit?: string
                    customAttrs: components["schemas"]["CustomAttr"][]
                    location?: string
                    positionCode?: string
                    purchases: components["schemas"]["Purchase"][]
                    partsTotalCount?: number
                    parts: components["schemas"]["AssetPart"][]
                }
            }
        }
        /** @description Updated asset data. */
        UpdateAssetRequest: {
            content: {
                "application/json": {
                    type: string
                    parentAssetID?: number
                    checkedOutTo?: number
                    /** @enum {string} */
                    status: "IN_STORAGE" | "IN_USE" | "ARCHIVED"
                    tag?: string
                    name: string
                    category?: string
                    model?: string
                    modelNo?: string
                    serialNo?: string
                    manufacturer?: string
                    notes?: string
                    /** Format: date */
                    warrantyUntil?: string
                    quantity?: number
                    quantityUnit?: string
                    customAttrs: components["schemas"]["CustomAttr"][]
                    location?: string
                    positionCode?: string
                    purchases: components["schemas"]["Purchase"][]
                    partsTotalCount?: number
                    parts: components["schemas"]["AssetPart"][]
                }
            }
        }
    }
    headers: never
    pathItems: never
}

export type $defs = Record<string, never>

export type external = Record<string, never>

export interface operations {
    ListAssets: {
        parameters: {
            query?: {
                page_size?: number
                page?: number
                type?: string
                order_by?: string
                order_dir?: string
                query?: string
            }
        }
        responses: {
            /** @description A paginated list of assets. */
            200: {
                content: {
                    "application/json": components["schemas"]["AssetListPage"]
                }
            }
        }
    }
    CreateAsset: {
        requestBody: components["requestBodies"]["CreateAssetRequest"]
        responses: {
            /** @description The newly created asset. */
            201: {
                content: {
                    "application/json": components["schemas"]["Asset"]
                }
            }
            /** @description Bad request. */
            400: {
                content: {
                    "application/json": components["schemas"]["Error"]
                }
            }
            /** @description Unauthorized. */
            401: {
                content: {
                    "application/json": components["schemas"]["Error"]
                }
            }
        }
    }
    GetAsset: {
        parameters: {
            query?: {
                include_children?: boolean
            }
            path: {
                tagOrID: string
            }
        }
        responses: {
            /** @description The requested asset. */
            200: {
                content: {
                    "application/json": components["schemas"]["Asset"]
                }
            }
            /** @description Unauthorized. */
            401: {
                content: {
                    "application/json": components["schemas"]["Error"]
                }
            }
            /** @description Asset not found. */
            404: {
                content: {
                    "application/json": components["schemas"]["Error"]
                }
            }
        }
    }
    UpdateAsset: {
        parameters: {
            path: {
                tagOrID: string
            }
        }
        requestBody: components["requestBodies"]["UpdateAssetRequest"]
        responses: {
            /** @description The updated asset. */
            200: {
                content: {
                    "application/json": components["schemas"]["Asset"]
                }
            }
            /** @description Bad Request. */
            400: {
                content: {
                    "application/json": components["schemas"]["Error"]
                }
            }
            /** @description Unauthorized. */
            401: {
                content: {
                    "application/json": components["schemas"]["Error"]
                }
            }
            /** @description Not found */
            404: {
                content: {
                    "application/json": components["schemas"]["Error"]
                }
            }
        }
    }
    DeleteAsset: {
        parameters: {
            path: {
                tagOrID: string
            }
        }
        responses: {
            /** @description The asset was deleted successfully. */
            204: {
                content: never
            }
            /** @description Unauthorized. */
            401: {
                content: {
                    "application/json": components["schemas"]["Error"]
                }
            }
            /** @description Not found. */
            404: {
                content: {
                    "application/json": components["schemas"]["Error"]
                }
            }
        }
    }
    ListTags: {
        parameters: {
            query?: {
                page_size?: number
                page?: number
                query?: string
            }
        }
        responses: {
            /** @description A paginated list of tags. */
            200: {
                content: {
                    "application/json": components["schemas"]["TagListPage"]
                }
            }
        }
    }
    ListCategories: {
        parameters: {
            query?: {
                page_size?: number
                page?: number
                query?: string
            }
        }
        responses: {
            /** @description A paginated list of categories. */
            200: {
                content: {
                    "application/json": components["schemas"]["CategoryListPage"]
                }
            }
        }
    }
    ListLocations: {
        parameters: {
            query?: {
                page_size?: number
                page?: number
                query?: string
            }
        }
        responses: {
            /** @description A paginated list of locations. */
            200: {
                content: {
                    "application/json": components["schemas"]["LocationListPage"]
                }
            }
        }
    }
    ListPositionCodes: {
        parameters: {
            query?: {
                page_size?: number
                page?: number
                query?: string
            }
        }
        responses: {
            /** @description A paginated list of position codes. */
            200: {
                content: {
                    "application/json": components["schemas"]["PositionCodeListPage"]
                }
            }
        }
    }
    ListModels: {
        parameters: {
            query?: {
                page_size?: number
                page?: number
                query?: string
            }
        }
        responses: {
            /** @description A paginated list of models. */
            200: {
                content: {
                    "application/json": components["schemas"]["ModelListPage"]
                }
            }
        }
    }
    ListManufacturers: {
        parameters: {
            query?: {
                page_size?: number
                page?: number
                query?: string
            }
        }
        responses: {
            /** @description A paginated list of manufacturers. */
            200: {
                content: {
                    "application/json": components["schemas"]["ManufacturerListPage"]
                }
            }
        }
    }
    ListCustomAttrs: {
        parameters: {
            query?: {
                page_size?: number
                page?: number
                query?: string
            }
        }
        responses: {
            /** @description A paginated list of custom attributes. */
            200: {
                content: {
                    "application/json": components["schemas"]["CustomAttrListPage"]
                }
            }
        }
    }
    ListSuppliers: {
        parameters: {
            query?: {
                page_size?: number
                page?: number
                query?: string
            }
        }
        responses: {
            /** @description A paginated list of suppliers. */
            200: {
                content: {
                    "application/json": components["schemas"]["SupplierListPage"]
                }
            }
        }
    }
    ListUsers: {
        parameters: {
            query?: {
                page_size?: number
                page?: number
                type?: string
                order_by?: string
                order_dir?: string
                query?: string
            }
        }
        responses: {
            /** @description A paginated list of users. */
            200: {
                content: {
                    "application/json": components["schemas"]["UserListPage"]
                }
            }
        }
    }
}
