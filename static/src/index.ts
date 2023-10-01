import _Alpine from "alpinejs"
import { plugin as autocomplete } from "./autocompleter"
import { plugin as commandpalette } from "./command_palette"
import { plugin as settings } from "./settings"
import { plugin as theme } from "./theme"

declare global {
    namespace globalThis {
        // biome-ignore lint/style/noVar: TypeScript needs it to be var
        var Alpine: typeof _Alpine
    }
}

_Alpine.plugin(autocomplete)
_Alpine.plugin(commandpalette)
_Alpine.plugin(settings)
_Alpine.plugin(theme)

globalThis.Alpine = _Alpine
globalThis.Alpine.start()
