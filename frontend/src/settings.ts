import type _Alpine from "alpinejs"
import { API } from "./api"

export function plugin(Alpine: typeof _Alpine) {
    let api: API = new API({ baseURL: `${location.origin}` })
    Alpine.magic("setting", () => (value: Record<string, unknown>) => {
        api.setSettings(value).catch(console.error)
    })
}
