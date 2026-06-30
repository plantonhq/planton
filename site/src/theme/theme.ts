'use client';

import { createTheme } from '@mui/material/styles';

export const theme = createTheme({
  palette: {
    mode: 'dark',
    primary: {
      main: '#a855f7', // Purple-500 to match Planton branding
      light: '#c084fc', // Purple-400
      dark: '#9333ea', // Purple-600
    },
    secondary: {
      main: '#64748b', // Slate-500
      light: '#94a3b8', // Slate-400
      dark: '#475569', // Slate-600
    },
    background: {
      default: '#0f172a', // Slate-950
      paper: '#1e293b', // Slate-900
    },
    text: {
      primary: '#f8fafc', // Slate-50
      secondary: '#cbd5e1', // Slate-300
    },
  },
  typography: {
    fontFamily: 'var(--font-geist-sans), -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
    fontWeightLight: 300,
    fontWeightRegular: 400,
    fontWeightMedium: 500,
    fontWeightBold: 700,
  },
  components: {
    MuiButton: {
      styleOverrides: {
        root: {
          textTransform: 'none',
          borderRadius: '0.375rem',
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          backgroundImage: 'none',
        },
      },
    },
  },
});

