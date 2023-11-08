import type _Alpine from "alpinejs"

export function plugin(Alpine: typeof _Alpine) {
    Alpine.magic("setting", () => (value: Record<string, unknown>) => {
        let csrf =
            document
                .querySelector(`meta[name="csrf-token"]`)
                ?.getAttribute("content") ?? ""

        fetch("/users/settings", {
            method: "post",
            credentials: "same-origin",
            headers: {
                "X-CSRF-Token": csrf,
                "Content-Type": "application/json; charset=utf-8",
            },
            body: JSON.stringify(value),
        }).catch(console.error)
    })
}
