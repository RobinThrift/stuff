import type _Alpine from "alpinejs"
import { get } from "./lib"

interface SuggestionItem {
    label: string
    value: string
}

class AutoCompleter {
    private _abort: AbortController | undefined
    private _source: string
    private _itemsAt?: string
    private _valueAt: string
    private _labelAt?: string

    constructor({
        source,
        itemsAt,
        valueAt,
        labelAt,
    }: {
        source: string
        itemsAt?: string
        valueAt: string
        labelAt?: string
    }) {
        this._source = source
        this._itemsAt = itemsAt
        this._valueAt = valueAt
        this._labelAt = labelAt
    }

    async fetch(query: string): Promise<SuggestionItem[]> {
        if (this._abort) {
            this._abort.abort()
        }

        this._abort = new AbortController()

        let url = new URL(location.origin + this._source)
        url.searchParams.set("query", query)

        let res = await fetch(url)
        let values = await res.json()

        if (this._itemsAt) {
            values = get(this._itemsAt, values) as Record<string, unknown>[]
        } else {
            values = values as Record<string, unknown>[]
        }

        let items: SuggestionItem[] = values.map(
            (v: Record<string, unknown>) => {
                let item = { value: v, label: v }

                if (this._valueAt) {
                    item.value = get(this._valueAt, v)
                    item.label = item.value
                }

                if (this._labelAt) {
                    item.label = get(this._labelAt, v)
                }

                return item
            },
        )

        this._abort = undefined

        return items
    }
}

export function plugin(Alpine: typeof _Alpine) {
    Alpine.data(
        "autocompleter",
        // biome-ignore lint/suspicious/noExplicitAny: The alpine types are bad
        ({ source, itemsAt, valueAt, labelAt, value }: any) => ({
            open: false,
            qs: null,
            items: [] as SuggestionItem[],
            value,

            autocompleter: new AutoCompleter({
                source,
                itemsAt,
                valueAt,
                labelAt: labelAt,
            }),

            async onChange(el: HTMLInputElement) {
                if (!el.value.length) {
                    this.items = []
                    this.open = false
                    return
                }

                this.items = await this.autocompleter.fetch(el.value)

                this.open = this.items.length > 0
                if (this.open) {
                    let pos = el.getBoundingClientRect()
                    let top = pos.top + window.scrollY
                    let left = pos.left + window.scrollX
                    this.$refs.suggestions.style.left = `${left}px`
                    this.$refs.suggestions.style.top = `${
                        top + el.offsetHeight
                    }px`
                    this.$refs.suggestions.style.minWidth = `${el.offsetWidth}px`
                }
            },

            onClickSuggestion(item: SuggestionItem) {
                this.value = ""
                this.value = item.value
                this.open = false
            },
        }),
    )
}
