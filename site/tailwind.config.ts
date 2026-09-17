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
        ...tailwindColors,
        white: '#ededed',
        // The numeric ramps below predate the role names above. No marketing
        // component uses them; they stay only until their last consumer is
        // rebuilt, and nothing new may reach for them.
        primary: {
          0: "#000000",
          10: "#1a1a1a",
          20: "#2a2a2a",
          30: "#3a3a3a",
          40: "#555555",
          50: "#ededed",
          60: "#d4d4d4",
          70: "#a0a0a0",
          80: "#888888",
          90: "#666666",
          95: "#f5f5f5",
          100: "#ededed",
        },
        secondary: {
          0: "#000000",
          10: "#1a1a1a",
          20: "#2a2a2a",
          30: "#3a3a3a",
          40: "#555555",
          50: "#666666",
          60: "#888888",
          70: "#a0a0a0",
          80: "#b0b0b0",
          90: "#d4d4d4",
          95: "#f5f5f5",
          100: "#ededed",
        },
        text: { secondary: '#999999' },
        
        tour: {
          // shadcn/ui style primary for tour components
          primary: {
            DEFAULT: "#171717",
            foreground: "#ffffff",
          },
          // shadcn/ui style secondary for tour components
          secondary: {
            DEFAULT: "#f5f5f5",
            foreground: "#171717",
          },
          // Base colors for tour components
          border: "#e5e7eb",
          input: "#e5e7eb", 
          ring: "#171717",
          background: "#ffffff",
          foreground: "hsl(var(--foreground))",
          
          // Tour component color variants
          destructive: {
            DEFAULT: "#ef4444",
            foreground: "#fefefe",
          },
          muted: {
            DEFAULT: "#f5f5f5",
            foreground: "#737373",
          },
          accent: {
            DEFAULT: "#f5f5f5",
            foreground: "#171717",
          },
          popover: {
            DEFAULT: "#ffffff",
            foreground: "#171717",
          },
          card: {
            DEFAULT: "#ffffff",
            foreground: "#171717",
          },
        },
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
