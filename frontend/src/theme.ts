import type _Alpine from "alpinejs"
import { API } from "./api"

type ThemeNames = "default" | "retro"
type ThemeModes = "system" | "light" | "dark"

class Theme {
    name: ThemeNames = "default"
    mode: ThemeModes = "system"

    private api: API = new API({ baseURL: `${location.origin}` })

    get dark(): boolean {
        return (
            this.mode === "dark" ||
            (this.mode === "system" &&
                window.matchMedia("(prefers-color-scheme: dark)").matches)
        )
    }

    constructor() {
        if (document.documentElement.classList.contains("dark")) {
            this.mode = "dark"
        } else if (document.documentElement.classList.contains("light")) {
            this.mode = "light"
        }

        if (document.documentElement.classList.contains("default")) {
            this.name = "default"
        } else if (document.documentElement.classList.contains("retro")) {
            this.name = "retro"
        }

        this._onChange()
    }

    setTheme(name: Theme["name"]) {
        this.name = name
        this._onChange()
        this._save()
    }

    setMode(mode: Theme["mode"]) {
        this.mode = mode
        this._onChange()
        this._save()
    }

    _onChange() {
        if (this.dark) {
            document.documentElement.classList.add("dark")
        } else {
            document.documentElement.classList.remove("dark")
        }

        if (this.name === "retro") {
            document.documentElement.classList.add("retro")
        } else {
            document.documentElement.classList.remove("retro")
        }
    }

    _save() {
        this.api
            .setSettings({
                theme_name: this.name,
                theme_mode: this.mode,
            })
            .catch(console.error)
    }
}

export function plugin(Alpine: typeof _Alpine) {
    Alpine.store("theme", new Theme())
}
