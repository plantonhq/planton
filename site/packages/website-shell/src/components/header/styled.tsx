'use client';

import {
  type AccordionProps,
  Drawer,
  AccordionSummary as MuiAccordionSummary,
  Accordion as MuiAccordion,
  AccordionDetails as MuiAccordionDetails,
  type AccordionSummaryProps,
  SvgIcon,
  type SvgIconProps,
} from '@mui/material';
import { styled } from '@mui/material/styles';
import { scopedTokens as tokens } from '../../theme/tokens';

export const SvgIconSizeResponsive = styled(SvgIcon)<SvgIconProps>(({ theme }) => ({
  fontSize: 16,
  [theme.breakpoints.up('md')]: {
    fontSize: 24,
  },
}));

export const ShellDrawer = styled(Drawer)(({ theme }) => ({
  '& > .MuiPaper-root': {
    padding: theme.spacing(2.5, 3.5),
    width: '90vw',
    maxWidth: 350,
    backgroundColor: tokens.surface.panel,
  },
}));

export const ShellAccordion = styled((props: AccordionProps) => (
  <MuiAccordion disableGutters elevation={0} square {...props} />
))(({ theme }) => ({
  padding: 0,
  background: 'transparent',
  '&.Mui-expanded': {
    display: 'flex',
    flexDirection: 'column',
    gap: theme.spacing(3),
  },
  '&::before': {
    display: 'none',
  },
}));

export const ShellAccordionSummary = styled((props: AccordionSummaryProps) => (
  <MuiAccordionSummary {...props} />
))(({ theme }) => ({
  padding: 0,
  minHeight: 'unset',
  '& .MuiAccordionSummary-content': {
    margin: 0,
    display: 'flex',
    flexDirection: 'row',
    justifyContent: 'space-between',
    gap: theme.spacing(1),
  },
  '& .MuiAccordionSummary-content .shell-accordion-title': {
    fontSize: theme.spacing(2),
    fontWeight: 600,
    color: tokens.text.secondary,
  },
  '&.Mui-expanded .shell-accordion-title': {
    color: tokens.text.primary,
  },
}));

export const ShellAccordionDetails = styled(MuiAccordionDetails)({
  padding: 0,
  borderBottom: 'none',
});
