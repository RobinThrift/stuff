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
    Alpine.directive(
        "autocomplete",
        (el, { expression }, { evaluateLater, effect, cleanup }) => {
            let input = el as HTMLInputElement
            let dl = document.createElement("datalist")
            dl.id = `${input.id}_autocomplete`
            input.setAttribute("list", dl.id)
            document.documentElement.appendChild(dl)

            let state = Alpine.reactive({
                items: [] as SuggestionItem[],
                autocompleter: undefined as unknown as AutoCompleter,
                debounce: 300,
            })

            let init = evaluateLater(expression)

            effect(() => {
                // biome-ignore lint/suspicious/noExplicitAny: The alpine types are bad
                init(({ source, itemsAt, valueAt, labelAt, debounce }: any) => {
                    state.autocompleter = new AutoCompleter({
                        source,
                        itemsAt,
                        valueAt,
                        labelAt: labelAt,
                    })
                    if (debounce) {
                        state.debounce = debounce
                    }
                })
            })

            effect(() => {
                dl.innerHTML = ""
                state.items.forEach((item) => {
                    let opt = document.createElement("option")
                    opt.value = item.value
                    opt.innerText = item.label
                    dl.appendChild(opt)
                })
            })

            let debounce: ReturnType<typeof setTimeout> | undefined
            let onchange = () => {
                if (!input.value.length) {
                    state.items = []
                    return
                }

                if (debounce) {
                    clearTimeout(debounce)
                }

                debounce = setTimeout(() => {
                    if (!input.value.length) {
                        state.items = []
                        return
                    }

                    state.autocompleter.fetch(input.value).then((items) => {
                        state.items = items
                    })
                }, state.debounce)
            }

            input.addEventListener("input", onchange)

            cleanup(() => {
                input.removeEventListener("input", onchange)
                dl.remove()
            })
        },
    )
}
