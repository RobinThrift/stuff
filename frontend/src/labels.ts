import type _Alpine from "alpinejs"
import type { AlpineComponent } from "alpinejs"
import { API, Asset } from "./api"

interface SheetTemplate {
    page: {
        size: string
        // Cols per page.
        cols: number
        // Rows per page.
        rows: number
        // Left margin of the page in mm.
        marginLeft: number
        // Top margin of the page in mm.
        marginTop: number
        // Right margin of the page in mm.
        marginRight: number
        // Bottom margin of the page in mm.
        marginBottom: number
    }

    label: {
        fontSize: number
        // Height of a single label in mm including padding.
        height: number
        // Width of a single label in mm including padding.
        width: number
        // VerticalPadding applied to the inside of the label in mm.
        verticalPadding: number
        // HorizontalPadding applied to the inside of the label in mm.
        horizontalPadding: number
        // VerticalSpacing between each label in mm.
        verticalSpacing: number
        // HorizontalSpacing between each label in mm.
        horizontalSpacing: number
    }
}

const templates: Record<string, SheetTemplate> = {
    "Avery L78710-20": {
        page: {
            size: "A4",
            marginTop: 13.3,
            marginBottom: 13.0,
            marginLeft: 8.5,
            marginRight: 8.5,
            cols: 7,
            rows: 27,
        },
        label: {
            fontSize: 4,
            height: 10,
            width: 25.4,
            horizontalPadding: 1,
            verticalPadding: 1,
            horizontalSpacing: 2.5,
            verticalSpacing: 0,
        },
    },
}

interface Data {
    assetSearchQuery: string
    fields: Element[]
    api: API
    assets: Asset[]
    selected: Asset[]
    selectedIDs: string[]
}

export function plugin(Alpine: typeof _Alpine) {
    Alpine.data("labelSheetCreator", (_init) => {
        let init = _init as { selected: Asset[] }
        let selectedIDs = init.selected.map((s) => s.id.toString())

        let data: AlpineComponent<Data> = {
            assetSearchQuery: "",
            fields: [],
            api: new API({ baseURL: `${location.origin}/api/v1` }),
            assets: [],
            selected: init.selected,
            selectedIDs,

            setTemplate(e: Event) {
                let el = e.target as HTMLSelectElement
                let template = templates[el.value]
                if (!template) {
                    return
                }

                this.fields.forEach((f) => {
                    let field = f as HTMLInputElement | HTMLSelectElement
                    let value = getFieldValue(field.name, template)
                    if (value && value !== "undefined") {
                        field.value = value
                    }
                })
            },

            setAssetSearchQuery(newValue: string) {
                this.assetSearchQuery = newValue
                if (newValue === "") {
                    this.assets = []
                    return
                }
                this.fetchAssets().catch(console.error)
            },

            async fetchAssets() {
                if (!this.assetSearchQuery) {
                    this.assets = []
                    return
                }
                let assetsPage = await this.api.listAssets({
                    query: this.assetSearchQuery,
                })
                this.assets = assetsPage?.assets ?? []
            },

            addAssetToSelection(asset: Asset) {
                if (this.isSelected(asset)) {
                    return
                }

                this.selected.push({ ...asset })
                this.selectedIDs.push(asset.id.toString())
            },

            addAllAssetsToSelection() {
                let selected = [...this.selected]
                this.assets.forEach((a) => {
                    if (!selected.find((s) => s.id === a.id)) {
                        selected.push(a)
                    }
                })

                let selectedIDs = new Set([
                    ...this.selectedIDs,
                    ...this.assets.map((s) => s.id.toString()),
                ])

                this.selected = selected
                this.selectedIDs = [...selectedIDs]
            },

            isSelected(asset: Asset) {
                return this.selected.find((a) => a.id === asset.id)
            },

            init() {
                this.fields = [...this.$el.querySelectorAll("input,select")]
            },
        }

        return data
    })
}

function getFieldValue(
    field: string,
    // biome-ignore lint/suspicious/noExplicitAny: Fix later
    tmpl: Record<string, any>,
): string | undefined {
    let [element, ...rest] = field.split("_")

    if (rest.length === 0) {
        return
    }

    let prop = rest[0]
    if (rest.length === 2) {
        prop = `${rest[0]}${rest[1]
            .substring(0, 1)
            .toUpperCase()}${rest[1].substring(1)}`
    }

    if (!tmpl[element]) {
        return ""
    }

    return `${tmpl[element][prop]}`
}
