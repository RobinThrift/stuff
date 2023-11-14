const defaultTheme = require("tailwindcss/defaultTheme")
/** @type {import('tailwindcss').Config} */
module.exports = {
    content: ["./views/templates/**/*.html.tmpl"],
    darkMode: "class",

    plugins: [
        require("@tailwindcss/forms")({
            strategy: "base", // only generate global styles
        }),
    ],

    theme: {
        extend: {
            ...defaultTheme.colors,
            colors: {
                background: {
                    default:
                        "rgb(var(--colour-background-default) / <alpha-value>)",
                    accent: "rgb(var(--colour-background-accent) / <alpha-value>)",
                    hover: "rgb(var(--colour-background-hover) / <alpha-value>)",
                    "accent-lighter":
                        "rgb(var(--colour-background-accent-lighter) / <alpha-value>)",
                },

                border: {
                    default:
                        "rgb(var(--colour-border-default) / <alpha-value>)",
                    light: "rgb(var(--colour-border-light) / <alpha-value>)",
                    lighter:
                        "rgb(var(--colour-border-lighter) / <alpha-value>)",
                },

                content: {
                    default:
                        "rgb(var(--colour-content-default) / <alpha-value>)",
                    light: "rgb(var(--colour-content-light) / <alpha-value>)",
                    lighter:
                        "rgb(var(--colour-content-lighter) / <alpha-value>)",
                },

                primary: {
                    default:
                        "rgb(var(--colour-primary-default) / <alpha-value>)",
                    darker: "rgb(var(--colour-primary-darker) / <alpha-value>)",
                    lighter:
                        "rgb(var(--colour-primary-lighter) / <alpha-value>)",
                    hover: "rgb(var(--colour-primary-hover) / <alpha-value>)",
                    active: "rgb(var(--colour-primary-active) / <alpha-value>)",
                },

                danger: {
                    default:
                        "rgb(var(--colour-danger-default) / <alpha-value>)",
                    hover: "rgb(var(--colour-danger-hover) / <alpha-value>)",
                    active: "rgb(var(--colour-danger-active) / <alpha-value>)",
                },

                success: {
                    default:
                        "rgb(var(--colour-success-default) / <alpha-value>)",
                    lighter:
                        "rgb(var(--colour-success-lighter) / <alpha-value>)",
                    border: "rgb(var(--colour-success-border) / <alpha-value>)",
                    hover: "rgb(var(--colour-success-hover) / <alpha-value>)",
                    active: "rgb(var(--colour-success-active) / <alpha-value>)",
                },
            },
        },
    },
}
