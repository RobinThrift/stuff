module.exports = {
    plugins: {
        "postcss-import": {},
        tailwindcss: { config: "./static/tailwind.config.js" },
        "tailwindcss/nesting": {},
        autoprefixer: {},
    },
}
