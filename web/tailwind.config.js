/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        bg: "var(--color-bg)",
        surface: "var(--color-surface)",
        "surface-2": "var(--color-surface-2)",
        stroke: "var(--color-stroke)",
        paper: "var(--color-paper)",
        "text-2": "var(--color-text-2)",
        "text-3": "var(--color-text-3)",
        gold: "var(--color-gold)",
        "gold-deep": "var(--color-gold-deep)",
        "gold-ink": "var(--color-gold-ink)",
        naira: "var(--color-naira)",
        "naira-deep": "var(--color-naira-deep)",
        red: "var(--color-red)",
        violet: "var(--color-violet)",
      },
      fontFamily: {
        display: ["var(--font-display)", "Bricolage Grotesque", "sans-serif"],
        body: ["var(--font-body)", "Work Sans", "sans-serif"],
      },
      borderRadius: {
        DEFAULT: "var(--radius)",
        lg: "var(--radius-lg)",
      },
    },
  },
  plugins: [],
};
