import { Box, Typography } from '@mui/material';
import type { FC } from 'react';
import { TerminalWindow } from './terminal-window';

/**
 * A record of labeled facts inside a window: the shape of what the product
 * shows after a deploy, a pause, an import, a connection. Marketing pages
 * use it as the proof beside a claim. The fields are the product's real
 * fields; the values are illustrations and the footer says so wherever a
 * value is a number. Each value is one fact or a few, separated by a middle
 * dot the caller writes, never by whitespace. On phones the label sits above
 * its value so nothing is clipped.
 */
export interface RecordRow {
  label: string;
  value: string;
}

export interface RecordWindowProps {
  title: string;
  rows: readonly RecordRow[];
  footer?: string;
  className?: string;
}

export const RecordWindow: FC<RecordWindowProps> = ({ title, rows, footer, className }) => (
  <TerminalWindow title={title} className={className}>
    <Box className="grid grid-cols-1 sm:grid-cols-[12rem_1fr] gap-x-4 gap-y-2">
      {rows.map((row) => (
        <Box key={row.label} className="contents">
          <Typography className="text-xs text-fg-muted font-mono sm:pt-0.5">{row.label}</Typography>
          <Typography className="text-xs text-fg font-mono break-words">{row.value}</Typography>
        </Box>
      ))}
    </Box>
    {footer && <Typography className="text-xs text-fg-muted mt-4">{footer}</Typography>}
  </TerminalWindow>
);
