'use client';
import { createWebsiteTheme } from '@planton/website-shell/theme';
import { homepageLightTokens as p, lightVariables as css } from './homepage';
export const lightWebsiteTheme = createWebsiteTheme({
  palette: { mode:'light', background:{default:p.surface.canvas,paper:p.surface.panel}, text:{primary:p.text.primary,secondary:p.text.secondary,disabled:p.text.muted}, divider:p.edge.default, primary:{main:p.cta.background,contrastText:p.cta.text}, grey:{100:p.text.primary,200:p.surface.raised,300:p.text.secondary,500:p.text.muted,600:p.text.faint,700:p.edge.hover,800:p.edge.default,900:p.surface.raised} },
  components: { MuiModal:{styleOverrides:{root:{...css}}}, MuiPopover:{styleOverrides:{root:{...css}}}, MuiDrawer:{styleOverrides:{root:{...css}}} },
});
