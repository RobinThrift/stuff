import type _Alpine from "alpinejs"

export function plugin(Alpine: typeof _Alpine) {
    let theme = {
        name: "default",
        dark: window.matchMedia("(prefers-color-scheme: dark)").matches,

        set(name: string) {
            document.documentElement.classList.remove(`theme-${this.name}`)

            this.name = name
            document.documentElement.classList.add(`theme-${this.name}`)
            localStorage.set(
                "theme",
                JSON.stringify({ name: this.name, dark: this.dark }),
            )
        },

        toggleDark() {
            this.dark = !this.dark

            if (this.dark) {
                document.documentElement.classList.add("dark")
            } else {
                document.documentElement.classList.remove("dark")
            }

            localStorage.set(
                "theme",
                JSON.stringify({ name: this.name, dark: this.dark }),
            )
        },
    }

    let storedJSON = localStorage.getItem("theme")

    if (storedJSON) {
        let stored = JSON.parse(storedJSON)
        theme.dark = stored.dark
        theme.name = stored.name
    }

    if (theme.dark) {
        document.documentElement.classList.add("dark")
    }

    Alpine.store("theme", theme)
}
