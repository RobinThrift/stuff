// const defaultTheme = require("tailwindcss/defaultTheme")
/** @type {import('tailwindcss').Config} */
module.exports = {
    content: ["./views/templates/**/*.html.tmpl"],
    darkMode: "class",
    theme: {
        extend: {
            colors: {
                blue: {
                    50: "rgb(var(--colour-blue-50) / <alpha-value>)",
                    100: "rgb(var(--colour-blue-100) / <alpha-value>)",
                    200: "rgb(var(--colour-blue-200) / <alpha-value>)",
                    300: "rgb(var(--colour-blue-300) / <alpha-value>)",
                    400: "rgb(var(--colour-blue-400) / <alpha-value>)",
                    500: "rgb(var(--colour-blue-500) / <alpha-value>)",
                    600: "rgb(var(--colour-blue-600) / <alpha-value>)",
                    700: "rgb(var(--colour-blue-700) / <alpha-value>)",
                    800: "rgb(var(--colour-blue-800) / <alpha-value>)",
                    900: "rgb(var(--colour-blue-900) / <alpha-value>)",
                    950: "rgb(var(--colour-blue-950) / <alpha-value>)",
                },

                red: {
                    50: "rgb(var(--colour-red-50) / <alpha-value>)",
                    100: "rgb(var(--colour-red-100) / <alpha-value>)",
                    200: "rgb(var(--colour-red-200) / <alpha-value>)",
                    300: "rgb(var(--colour-red-300) / <alpha-value>)",
                    400: "rgb(var(--colour-red-400) / <alpha-value>)",
                    500: "rgb(var(--colour-red-500) / <alpha-value>)",
                    600: "rgb(var(--colour-red-600) / <alpha-value>)",
                    700: "rgb(var(--colour-red-700) / <alpha-value>)",
                    800: "rgb(var(--colour-red-800) / <alpha-value>)",
                    900: "rgb(var(--colour-red-900) / <alpha-value>)",
                    950: "rgb(var(--colour-red-950) / <alpha-value>)",
                },

                yellow: {
                    50: "rgb(var(--colour-yellow-50) / <alpha-value>)",
                    100: "rgb(var(--colour-yellow-100) / <alpha-value>)",
                    200: "rgb(var(--colour-yellow-200) / <alpha-value>)",
                    300: "rgb(var(--colour-yellow-300) / <alpha-value>)",
                    400: "rgb(var(--colour-yellow-400) / <alpha-value>)",
                    500: "rgb(var(--colour-yellow-500) / <alpha-value>)",
                    600: "rgb(var(--colour-yellow-600) / <alpha-value>)",
                    700: "rgb(var(--colour-yellow-700) / <alpha-value>)",
                    800: "rgb(var(--colour-yellow-800) / <alpha-value>)",
                    900: "rgb(var(--colour-yellow-900) / <alpha-value>)",
                    950: "rgb(var(--colour-yellow-950) / <alpha-value>)",
                },

                green: {
                    50: "rgb(var(--colour-green-50) / <alpha-value>)",
                    100: "rgb(var(--colour-green-100) / <alpha-value>)",
                    200: "rgb(var(--colour-green-200) / <alpha-value>)",
                    300: "rgb(var(--colour-green-300) / <alpha-value>)",
                    400: "rgb(var(--colour-green-400) / <alpha-value>)",
                    500: "rgb(var(--colour-green-500) / <alpha-value>)",
                    600: "rgb(var(--colour-green-600) / <alpha-value>)",
                    700: "rgb(var(--colour-green-700) / <alpha-value>)",
                    800: "rgb(var(--colour-green-800) / <alpha-value>)",
                    900: "rgb(var(--colour-green-900) / <alpha-value>)",
                    950: "rgb(var(--colour-green-950) / <alpha-value>)",
                },

                background: {
                    100: "rgb(var(--colour-background-100) / <alpha-value>)",
                    200: "rgb(var(--colour-background-200) / <alpha-value>)",
                    300: "rgb(var(--colour-background-300) / <alpha-value>)",
                },

                content: "rgb(var(--colour-content) / <alpha-value>)",
                "content-inverse":
                    "rgb(var(--colour-content-inverse) / <alpha-value>)",

                primary: {
                    50: "rgb(var(--colour-primary-50) / <alpha-value>)",
                    100: "rgb(var(--colour-primary-100) / <alpha-value>)",
                    200: "rgb(var(--colour-primary-200) / <alpha-value>)",
                    300: "rgb(var(--colour-primary-300) / <alpha-value>)",
                    400: "rgb(var(--colour-primary-400) / <alpha-value>)",
                    500: "rgb(var(--colour-primary-500) / <alpha-value>)",
                    600: "rgb(var(--colour-primary-600) / <alpha-value>)",
                    700: "rgb(var(--colour-primary-700) / <alpha-value>)",
                    800: "rgb(var(--colour-primary-800) / <alpha-value>)",
                    900: "rgb(var(--colour-primary-900) / <alpha-value>)",
                    950: "rgb(var(--colour-primary-950) / <alpha-value>)",
                },

                secondary: {
                    50: "rgb(var(--colour-secondary-50) / <alpha-value>)",
                    100: "rgb(var(--colour-secondary-100) / <alpha-value>)",
                    200: "rgb(var(--colour-secondary-200) / <alpha-value>)",
                    300: "rgb(var(--colour-secondary-300) / <alpha-value>)",
                    400: "rgb(var(--colour-secondary-400) / <alpha-value>)",
                    500: "rgb(var(--colour-secondary-500) / <alpha-value>)",
                    600: "rgb(var(--colour-secondary-600) / <alpha-value>)",
                    700: "rgb(var(--colour-secondary-700) / <alpha-value>)",
                    800: "rgb(var(--colour-secondary-800) / <alpha-value>)",
                    900: "rgb(var(--colour-secondary-900) / <alpha-value>)",
                    950: "rgb(var(--colour-secondary-950) / <alpha-value>)",
                },

                accent: {
                    50: "rgb(var(--colour-accent-50) / <alpha-value>)",
                    100: "rgb(var(--colour-accent-100) / <alpha-value>)",
                    200: "rgb(var(--colour-accent-200) / <alpha-value>)",
                    300: "rgb(var(--colour-accent-300) / <alpha-value>)",
                    400: "rgb(var(--colour-accent-400) / <alpha-value>)",
                    500: "rgb(var(--colour-accent-500) / <alpha-value>)",
                    600: "rgb(var(--colour-accent-600) / <alpha-value>)",
                    700: "rgb(var(--colour-accent-700) / <alpha-value>)",
                    800: "rgb(var(--colour-accent-800) / <alpha-value>)",
                    900: "rgb(var(--colour-accent-900) / <alpha-value>)",
                    950: "rgb(var(--colour-accent-950) / <alpha-value>)",
                },

                neutral: {
                    50: "rgb(var(--colour-neutral-50) / <alpha-value>)",
                    100: "rgb(var(--colour-neutral-100) / <alpha-value>)",
                    200: "rgb(var(--colour-neutral-200) / <alpha-value>)",
                    300: "rgb(var(--colour-neutral-300) / <alpha-value>)",
                    400: "rgb(var(--colour-neutral-400) / <alpha-value>)",
                    500: "rgb(var(--colour-neutral-500) / <alpha-value>)",
                    600: "rgb(var(--colour-neutral-600) / <alpha-value>)",
                    700: "rgb(var(--colour-neutral-700) / <alpha-value>)",
                    800: "rgb(var(--colour-neutral-800) / <alpha-value>)",
                    900: "rgb(var(--colour-neutral-900) / <alpha-value>)",
                    950: "rgb(var(--colour-neutral-950) / <alpha-value>)",
                },
            },
        },
    },
    plugins: [],
}
