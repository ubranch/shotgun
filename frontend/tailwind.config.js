/** @type {import('tailwindcss').Config} */
export default {
    content: ["./index.html", "./src/**/*.{vue,js,ts,jsx,tsx}"],
    darkMode: "class",
    theme: {
        extend: {
            fontFamily: {
                sans: ["var(--font-sans)"],
                mono: ["var(--font-mono)"],
                serif: ["var(--font-serif)"],
            },
            colors: {
                // Use CSS variables for colors
                background: "var(--background)",
                foreground: "var(--foreground)",
                card: "var(--card)",
                "card-foreground": "var(--card-foreground)",
                popover: "var(--popover)",
                "popover-foreground": "var(--popover-foreground)",
                primary: "var(--primary)",
                "primary-foreground": "var(--primary-foreground)",
                secondary: "var(--secondary)",
                "secondary-foreground": "var(--secondary-foreground)",
                muted: "var(--muted)",
                "muted-foreground": "var(--muted-foreground)",
                accent: "var(--accent)",
                "accent-foreground": "var(--accent-foreground)",
                destructive: "var(--destructive)",
                "destructive-foreground": "var(--destructive-foreground)",
                border: "var(--border)",
                input: "var(--input)",
                ring: "var(--ring)",

                // Sidebar specific colors
                sidebar: "var(--sidebar)",
                "sidebar-foreground": "var(--sidebar-foreground)",
                "sidebar-primary": "var(--sidebar-primary)",
                "sidebar-primary-foreground": "var(--sidebar-primary-foreground)",
                "sidebar-accent": "var(--sidebar-accent)",
                "sidebar-accent-foreground": "var(--sidebar-accent-foreground)",
                "sidebar-border": "var(--sidebar-border)",

                // Chart colors
                "chart-1": "var(--chart-1)",
                "chart-2": "var(--chart-2)",
                "chart-3": "var(--chart-3)",
                "chart-4": "var(--chart-4)",
                "chart-5": "var(--chart-5)",

                // Legacy colors (kept for backward compatibility)
                "light-bg": "var(--background)",
                "light-fg": "var(--foreground)",
                "light-surface": "var(--card)",
                "light-border": "var(--border)",
                "light-accent": "var(--accent)",
                "light-accent-hover": "var(--accent-foreground)",
                "dark-bg": "var(--background)",
                "dark-fg": "var(--foreground)",
                "dark-surface": "var(--card)",
                "dark-border": "var(--border)",
                "dark-accent": "var(--accent)",
                "dark-accent-hover": "var(--accent-foreground)",
            },
            borderRadius: {
                sm: "var(--radius-sm)",
                md: "var(--radius-md)",
                lg: "var(--radius-lg)",
                xl: "var(--radius-xl)",
            },
            boxShadow: {
                "2xs": "var(--shadow-2xs)",
                xs: "var(--shadow-xs)",
                sm: "var(--shadow-sm)",
                DEFAULT: "var(--shadow)",
                md: "var(--shadow-md)",
                lg: "var(--shadow-lg)",
                xl: "var(--shadow-xl)",
                "2xl": "var(--shadow-2xl)",
            },
        },
    },
    plugins: [],
};
