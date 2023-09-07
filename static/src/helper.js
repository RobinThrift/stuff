function get(path, obj) {
    let res = obj
    let parts = path.split(".")
    for (let i = 0; i <= parts.length; i++) {
        let p = parts[i]
        if (Array.isArray(res)) {
            res = res.map((r) => get(parts.slice(i).join("."), r))
            continue
        }

        if (res[p]) {
            res = res[p]
        }
    }

    return res
}

class AutoCompleter {
    /** @type AbortController | undefined */
    _abort
    /** @type string */
    _source
    /** @type string */
    _valueAt

    constructor({ source, valueAt }) {
        this._source = source
        this._valueAt = valueAt
    }

    async fetch(query) {
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

globalThis.AutoCompleter = AutoCompleter
