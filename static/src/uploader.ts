import type _Alpine from "alpinejs"

export function plugin(Alpine: typeof _Alpine) {
    Alpine.magic("upload", () => async (assetID: number, dt: DataTransfer) => {
        let csrf =
            document
                .querySelector(`meta[name="csrf-token"]`)
                ?.getAttribute("content") ?? ""

        let formdata = new FormData()
        ;[...dt.files].forEach((file) => {
            formdata.append(file.name, file)
        })

        let res = await fetch(`/assets/${assetID}/files`, {
            method: "POST",
            credentials: "same-origin",
            headers: { "X-CSRF-Token": csrf },
            body: formdata,
        })

        if (!res.ok) {
            // @TODO: replace with proper user facing error reporting
            console.error(
                `error uploading files: ${res.status} ${res.statusText}`,
                await res.text(),
            )
            return
        }

        // @TODO: find a nicer way to show new files :]
        location.reload()
    })
}
