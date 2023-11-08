import { components, operations } from "./apiv1"

export type Asset = components["schemas"]["Asset"]
export type AssetListPage = components["schemas"]["AssetListPage"]
export type AssetListQueryParams =
    operations["ListAssets"]["parameters"]["query"]

export class API {
    public baseURL: string
    private _abort: AbortController | undefined

    constructor({ baseURL }: { baseURL: string }) {
        this.baseURL = baseURL
    }

    async listAssets(query?: AssetListQueryParams): Promise<AssetListPage> {
        return this.fetch<AssetListPage>("/assets", query)
    }

    async fetch<R = Record<string, unknown>>(
        path: string,
        query?: AssetListQueryParams,
    ): Promise<R> {
        if (this._abort) {
            this._abort.abort()
        }

        this._abort = new AbortController()

        let url = new URL(`${this.baseURL}${path}`)
        if (query) {
            for (let k in query) {
                // biome-ignore lint/suspicious/noExplicitAny: TypeScript doesn't seem to like this
                let v = query[k as keyof AssetListQueryParams] as any
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
