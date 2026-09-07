/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        canvas: "rgb(var(--color-canvas) / <alpha-value>)",
        cream: "rgb(var(--color-cream) / <alpha-value>)",
        ink: "rgb(var(--color-ink) / <alpha-value>)",
        soft: "rgb(var(--color-soft) / <alpha-value>)",
        coral: "rgb(var(--color-coral) / <alpha-value>)",
        "coral-dark": "rgb(var(--color-coral-dark) / <alpha-value>)",
        violet: "rgb(var(--color-violet) / <alpha-value>)",
        "violet-dark": "rgb(var(--color-violet-dark) / <alpha-value>)",
        mint: "rgb(var(--color-mint) / <alpha-value>)",
        "mint-dark": "rgb(var(--color-mint-dark) / <alpha-value>)",
        gold: "rgb(var(--color-gold) / <alpha-value>)",
        "gold-dark": "rgb(var(--color-gold-dark) / <alpha-value>)",
        sky: "rgb(var(--color-sky) / <alpha-value>)",
        "sky-dark": "rgb(var(--color-sky-dark) / <alpha-value>)",
      },
      fontFamily: {
        display: ['"Fredoka"', "Trebuchet MS", "ui-rounded", "system-ui", "sans-serif"],
        body: ['"Karla"', "Trebuchet MS", "system-ui", "sans-serif"],
      },
      animation: {
        float: "float 6s ease-in-out infinite",
        "float-slow": "float 10s ease-in-out infinite",
        wiggle: "wiggle 2.6s ease-in-out infinite",
        pop: "pop 0.45s cubic-bezier(0.16, 1, 0.3, 1) both",
        rise: "rise 0.6s cubic-bezier(0.16, 1, 0.3, 1) both",
        "pulse-ring": "pulse-ring 2.2s ease-out infinite",
        "spin-slow": "spin 14s linear infinite",
        "pulse-soft": "pulse-soft 2.4s ease-in-out infinite",
      },
      keyframes: {
        float: {
          "0%, 100%": { transform: "translateY(0) rotate(var(--tilt, 0deg))" },
          "50%": { transform: "translateY(-14px) rotate(var(--tilt, 0deg))" },
        },
        wiggle: {
          "0%, 100%": { transform: "rotate(-4deg)" },
          "50%": { transform: "rotate(4deg)" },
        },
        pop: {
          "0%": { opacity: "0", transform: "scale(0.9)" },
          "100%": { opacity: "1", transform: "scale(1)" },
        },
        rise: {
          "0%": { opacity: "0", transform: "translateY(16px)" },
          "100%": { opacity: "1", transform: "translateY(0)" },
        },
        "pulse-ring": {
          "0%": { boxShadow: "0 0 0 0 rgb(108 76 241 / 0.4)" },
          "70%": { boxShadow: "0 0 0 12px rgb(108 76 241 / 0)" },
          "100%": { boxShadow: "0 0 0 0 rgb(108 76 241 / 0)" },
        },
        "pulse-soft": {
          "0%, 100%": { opacity: "1" },
          "50%": { opacity: "0.55" },
        },
      },
    },
  },
  plugins: [],
};
