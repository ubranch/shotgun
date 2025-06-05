/** @type {import('tailwindcss').Config} */
export default {
    content: ["./index.html", "./src/**/*.{vue,js,ts,jsx,tsx}"],
    darkMode: "class",
    theme: {
        extend: {
            fontFamily: {
                sans: ["Geist", "Inter", "sans-serif"],
                mono: ["GeistMono Nerd Font Mono", "monospace"],
            },
            colors: {
                // light mode colors
                "light-bg": "#f4f4f4",
                "light-fg": "#333333",
                "light-surface": "#ffffff",
                "light-border": "#e5e5e5",
                "light-accent": "#4f46e5",

                // dark mode colors
                "dark-bg": "#1a1a1a",
                "dark-fg": "#e5e5e5",
                "dark-surface": "#222222",
                "dark-border": "#353535",
                "dark-accent": "#6366f1",
            },
        },
    },
    plugins: [],
};
