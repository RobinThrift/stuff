import type _Alpine from "alpinejs"

type ThemeNames = "default" | "retro"
type ThemeModes = "system" | "light" | "dark"

class Theme {
    name: ThemeNames = "default"
    mode: ThemeModes = "system"

    get dark(): boolean {
        return (
            this.mode === "dark" ||
            (this.mode === "system" &&
                window.matchMedia("(prefers-color-scheme: dark)").matches)
        )
    }

    constructor() {
        let storedTheme = localStorage.getItem("theme")
        if (storedTheme) {
            let loaded = JSON.parse(storedTheme) as {
                name: ThemeNames
                mode: ThemeModes
            }
            this.name = loaded.name
            this.mode = loaded.mode
        }

        this._onChange()
    }

    setTheme(name: Theme["name"]) {
        this.name = name
        this._onChange()
    }

    setMode(mode: Theme["mode"]) {
        this.mode = mode
        this._onChange()
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

        this._save()
    }

    _save() {
        localStorage.setItem(
            "theme",
            JSON.stringify({
                name: this.name,
                mode: this.mode,
            }),
        )
    }
}

export function plugin(Alpine: typeof _Alpine) {
    Alpine.store("theme", new Theme())
}
