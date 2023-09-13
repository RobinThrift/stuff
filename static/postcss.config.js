export default {
	plugins: {
		"postcss-import": {},
		"tailwindcss/nesting": {},
		tailwindcss: { config: "./static/tailwind.config.js" },
		autoprefixer: {},
	},
};
