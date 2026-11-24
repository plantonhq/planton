'use client';

import { Drawer } from '@mui/material';
import { styled } from '@mui/material/styles';
import { scopedTokens as tokens } from '../../theme/tokens';

export const ShellDrawer = styled(Drawer)(({ theme }) => ({
  '& > .MuiPaper-root': {
    // The header's own gutter, so the logo does not jump when the drawer opens.
    padding: theme.spacing(2.5, 2.5),
    width: '90vw',
    maxWidth: 350,
    backgroundColor: tokens.surface.panel,
  },
}));
