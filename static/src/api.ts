export interface QueryParams {
    page_size?: number
    page?: number
    order_by?: string
    order_dir?: "ASC" | "DESC"
    query?: string
}

export class API {
    public baseURL: string
    private _abort: AbortController | undefined

    constructor({ baseURL }: { baseURL: string }) {
        this.baseURL = baseURL
    }

    async listAssets(query?: QueryParams): Promise<AssetListPage> {
        return this.fetch<AssetListPage>("/assets", query)
    }

    async fetch<R = Record<string, unknown>>(
        path: string,
        query?: QueryParams,
    ): Promise<R> {
        if (this._abort) {
            this._abort.abort()
        }

        this._abort = new AbortController()

        let url = new URL(`${this.baseURL}${path}`)
        if (query) {
            for (let k in query) {
                let v = query[k as keyof QueryParams]
                if (v) {
                    url.searchParams.set(k, v.toString())
                }
            }
        }

        let res = await fetch(url)
        let data: R = await res.json()

        this._abort = undefined

        return data
    }
}

export interface AssetListPage {
    assets: Asset[]
    total: number
    numPages: number
    page: number
    pageSize: number
}

export interface Asset {
    id: number
    tag: string
    status: string
    name: string
    category: string
    model: string
    modelNo: string
    manufacturer: string
    imageURL: string
    thumbnailURL: string
    warrantyUntil: string
    location: string
    positionCode: string
    purchaseDate: string
    purchaseCurrency: string
    partsTotalCounter?: number
}

export interface AssetPart {
    id: number
    assetID: number
    tag: string
    name: string
    location: string
    positionCode: string
    notes?: string
}
