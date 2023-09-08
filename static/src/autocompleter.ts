import type _Alpine from "alpinejs"
import { get } from "./lib"

class AutoCompleter {
    private _abort: AbortController | undefined
    private _source: string
    private _valueAt: string

    constructor({ source, valueAt }: { source: string; valueAt: string }) {
        this._source = source
        this._valueAt = valueAt
    }

    async fetch(query: string) {
        if (this._abort) {
            this._abort.abort()
        }

        this._abort = new AbortController()

        let url = new URL(location.origin + this._source)
        url.searchParams.set("query", query)

        let res = await fetch(url)
        let values = await res.json()

        if (this._valueAt) {
            values = get(this._valueAt, values)
        }

        this._abort = undefined
        return values
    }
}

export function plugin(Alpine: typeof _Alpine) {
    Alpine.data("autocompleter", ({ source, valueAt, value }: any) => ({
        open: false,
        qs: null,
        items: [],
        value,

        autocompleter: new AutoCompleter({ source, valueAt }),

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
                this.$refs.suggestions.style.left = left + "px"
                this.$refs.suggestions.style.top = top + el.offsetHeight + "px"
                this.$refs.suggestions.style.minWidth = el.offsetWidth + "px"
            }
        },

        onClickSuggestion(newValue: string) {
            this.value = ""
            this.value = newValue
            this.open = false
        },
    }))
}
