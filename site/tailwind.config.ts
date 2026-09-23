import type { Config } from "tailwindcss";
import plugin from "tailwindcss/plugin";
import { tailwindColors } from "./packages/website-shell/src/theme/tokens";

export default {
  content: [
    "./src/pages/**/*.{js,ts,jsx,tsx}",
    "./src/components/**/*.{js,ts,jsx,tsx}",
    "./src/app/**/*.{js,ts,jsx,tsx}",
    "./src/lib/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    container: {
      center: true,
    },
    extend: {
      fontFamily: {
        'inter': ['var(--font-inter)', 'Inter', 'sans-serif'],
      },
      colors: {
        // The palette, by role: bg-canvas, bg-card, text-fg-secondary,
        // border-edge, text-ok. Defined once in the website-shell package
        // and shared with the MUI theme; components never type a hex.
        ...Object.fromEntries(Object.entries(tailwindColors).map(([key, value]) => {
          const color = (name: string, hex: string) => `rgb(var(--marketing-${name}, ${[1,3,5].map(i => parseInt(hex.slice(i,i+2),16)).join(' ')}) / <alpha-value>)`;
          return [key, typeof value === 'string' ? color(key, value) : Object.fromEntries(Object.entries(value).map(([role, hex]) => [role, color(role === 'DEFAULT' ? key : `${key}-${role}`, hex)]))];
        })),
        white: 'rgb(var(--marketing-fg, 237 237 237) / <alpha-value>)',
        // One leftover the pricing FAQ still reads; it goes when pricing is rebuilt.
        text: { secondary: '#999999' },
      },
      backgroundImage: {
        "gradient-radial": "radial-gradient(var(--tw-gradient-stops))",
        "bg-gradient": "",
      },
      animation: {
        blink: 'blink 1s linear infinite',
        wiggle: 'wiggle 1s ease-in-out infinite',
      },
      keyframes: {
        blink: {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '0' },
        },
        wiggle: {
          "0%, 100%": { transform: "rotate(-3deg)" },
          "50%": { transform: "rotate(3deg)" },
        },
      },
      gridTemplateColumns: {
        "13": "repeat(13, minmax(0, 1fr))",
        "14": "repeat(14, minmax(0, 1fr))",
        "15": "repeat(15, minmax(0, 1fr))",
        "16": "repeat(16, minmax(0, 1fr))",
        "17": "repeat(17, minmax(0, 1fr))",
        "18": "repeat(18, minmax(0, 1fr))",
        "19": "repeat(19, minmax(0, 1fr))",
        "20": "repeat(20, minmax(0, 1fr))",
      },
    },
  },
  plugins: [
    plugin(function ({ matchUtilities, theme }) {
      matchUtilities(
        {
          "bg-gradient": (angle) => ({
            "background-image": `linear-gradient(${angle}deg, var(--tw-gradient-stops))`,
          }),
        },
        {
          // values from config and defaults you wish to use most
          values: theme("bgGradientDeg", {}), // name of config key. Must be unique,
        }
      );
    }),
  ],
} satisfies Config;
