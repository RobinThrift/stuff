import type _Alpine from "alpinejs";

export function plugin(Alpine: typeof _Alpine) {
	Alpine.magic("sidebar", () => (value: boolean, csrf: string) => {
		fetch("/users/session/sidebar/open", {
			method: "post",
			headers: {
				"X-CSRF-Token": csrf,
			},
			body: JSON.stringify({
				closed: value,
			}),
		}).catch(console.error);
	});
}
