/**
 * A terminal chrome for command samples: traffic lights, a title, and a
 * monospace body. The one place a marketing page shows a command.
 */
import { Box, Typography } from '@mui/material';
import type { FC, ReactNode } from 'react';

interface TerminalWindowProps {
  children: ReactNode;
  className?: string;
  title?: string;
}

export const TerminalWindow: FC<TerminalWindowProps> = ({
  children,
  className = '',
  title = 'Terminal',
}) => (
  <Box
    className={`
      rounded-xl bg-raised border border-edge
      overflow-hidden
      ${className}
    `}
  >
    <Box className="flex items-center gap-2 px-4 py-3 bg-edge border-b border-edge-hover">
      <Box className="w-3 h-3 rounded-full bg-danger" />
      <Box className="w-3 h-3 rounded-full bg-warn" />
      <Box className="w-3 h-3 rounded-full bg-ok" />
      <Typography className="ml-3 text-xs text-fg-muted">{title}</Typography>
    </Box>
    
    <Box className="p-4 font-mono text-sm">
      {children}
    </Box>
  </Box>
);
