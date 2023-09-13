import _Alpine from "alpinejs";
import { plugin as autocomplete } from "./autocompleter";
import { plugin as commandpalette } from "./command_palette";
import { plugin as sidebar } from "./sidebar";

declare global {
	namespace globalThis {
		// biome-ignore lint/style/noVar: TypeScript needs it to be var
		var Alpine: typeof _Alpine;
	}
}

_Alpine.plugin(autocomplete);
_Alpine.plugin(commandpalette);
_Alpine.plugin(sidebar);

globalThis.Alpine = _Alpine;
globalThis.Alpine.start();
