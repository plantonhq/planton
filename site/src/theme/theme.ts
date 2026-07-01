"use client";

import { createTheme } from "@mui/material/styles";
import { mono, semantic, radius } from "@/theme/tokens";

/**
 * MUI theme for the docs chrome (Drawer, IconButton, etc.). Monochrome, dark,
 * built from the same tokens the rest of the site uses. MUI receives concrete
 * hex values (it derives contrast/alpha internally and cannot read CSS vars).
 */
export const theme = createTheme({
  palette: {
    mode: "dark",
    primary: {
      main: semantic.primary,
      contrastText: semantic.primaryForeground,
    },
    secondary: {
      main: mono[400],
      contrastText: semantic.background,
    },
    background: {
      default: semantic.background,
      paper: semantic.card,
    },
    text: {
      primary: mono[100],
      secondary: mono[400],
      disabled: mono[500],
    },
    divider: semantic.border,
  },
  shape: {
    borderRadius: 12,
  },
  typography: {
    fontFamily:
      "var(--font-inter), ui-sans-serif, system-ui, -apple-system, sans-serif",
    fontWeightLight: 300,
    fontWeightRegular: 400,
    fontWeightMedium: 500,
    fontWeightBold: 700,
  },
  components: {
    MuiButton: {
      styleOverrides: {
        root: { textTransform: "none", borderRadius: radius.md },
      },
    },
    // Kill MUI's dark-mode elevation overlay so surfaces stay flat monochrome.
    MuiPaper: {
      styleOverrides: { root: { backgroundImage: "none" } },
    },
    MuiAppBar: {
      styleOverrides: { root: { backgroundImage: "none" } },
    },
  },
});

export default theme;
