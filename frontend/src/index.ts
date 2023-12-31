import _Alpine from "alpinejs"
import { plugin as autocomplete } from "./autocompleter"
import { plugin as barcodeScanner } from "./barcode_scanner"
import { plugin as commandpalette } from "./command_palette"
import { plugin as labels } from "./labels"
import { plugin as settings } from "./settings"
import { plugin as theme } from "./theme"
import { plugin as uploader } from "./uploader"

declare global {
    namespace globalThis {
        // biome-ignore lint/style/noVar: TypeScript needs it to be var
        var Alpine: typeof _Alpine
    }
}

_Alpine.plugin(autocomplete)
_Alpine.plugin(barcodeScanner)
_Alpine.plugin(commandpalette)
_Alpine.plugin(settings)
_Alpine.plugin(uploader)
_Alpine.plugin(labels)
_Alpine.plugin(theme)

globalThis.Alpine = _Alpine
globalThis.Alpine.start()
