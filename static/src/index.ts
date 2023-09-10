import _Alpine from "alpinejs"
import persist from "@alpinejs/persist"
import { plugin as autocomplete } from "./autocompleter"
import { plugin as commandpalette } from "./command_palette"

declare global {
    namespace globalThis {
        // biome-ignore lint/style/noVar: TypeScript needs it to be var
        var Alpine: typeof _Alpine
    }
}

_Alpine.plugin(persist)
_Alpine.plugin(autocomplete)
_Alpine.plugin(commandpalette)

globalThis.Alpine = _Alpine
globalThis.Alpine.start()
