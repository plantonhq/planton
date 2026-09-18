'use client';

/**
 * One or more lines a person pastes, with a single copy affordance for the
 * lot, in a chrome bar above the text so the button never sits on a line.
 * What is copied is the plain text, never the markup. A single command is a
 * one-line list and a file is its lines under the file's name, so there is
 * exactly one way a marketing page shows something to paste.
 *
 * A command may carry its own line breaks (a shell continuation, a file's
 * lines); each physical line is a block. Within a line, long text wraps
 * between tokens first: each space-separated token is an inline block, so a
 * path or a flag value moves to the next line whole rather than splitting at
 * one of its own hyphens; only a token wider than the block itself may break
 * inside. Nothing is ever clipped.
 *
 * Two densities: the framed block (a chrome bar, then the lines) for a
 * command that is the section's subject, and `compact` (one row, the copy
 * glyph at its end, no chrome) for a one-liner that sits inside a list of
 * steps, where a frame would outweigh the line it holds.
 *
 * This is a client module because the copy button holds two seconds of
 * state; the tabs primitive that composes it is the other.
 */
import { Box, IconButton } from '@mui/material';
import { Check, Copy } from 'lucide-react';
import { type FC, useCallback, useState } from 'react';

export interface CommandBlockProps {
  /** One command per line; copied together, separated by newlines. */
  commands: readonly string[];
  /** Screen-reader name for the copy button, e.g. "Copy the install commands". */
  label: string;
  /** The chrome bar's text; defaults to the terminal's. */
  title?: string;
  /** One row with the copy glyph at its end and no chrome bar, for a one-liner inside a list of steps. */
  compact?: boolean;
  className?: string;
}

export const CommandBlock: FC<CommandBlockProps> = ({ commands, label, title = 'Terminal', compact = false, className = '' }) => {
  const [copied, setCopied] = useState(false);
  const text = commands.join('\n');

  const handleCopy = useCallback(() => {
    navigator.clipboard.writeText(text).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  }, [text]);

  if (compact) {
    return (
      <Box className={`inline-flex items-center gap-2 max-w-full rounded-lg bg-raised border border-edge pl-3 pr-1 py-1 ${className}`}>
        <code className="font-mono text-xs text-fg-body whitespace-pre-wrap [overflow-wrap:anywhere] text-left">{text}</code>
        <IconButton onClick={handleCopy} size="small" aria-label={copied ? 'Copied' : label} className="!text-fg-muted hover:!text-white">
          {copied ? <Check size={14} aria-hidden /> : <Copy size={14} aria-hidden />}
        </IconButton>
      </Box>
    );
  }

  return (
    <Box className={`rounded-xl bg-raised border border-edge overflow-hidden ${className}`}>
      <Box className="flex items-center justify-between h-9 pl-4 pr-1 bg-edge border-b border-edge-hover">
        <span className="text-xs text-fg-muted">{title}</span>
        <IconButton onClick={handleCopy} size="small" aria-label={copied ? 'Copied' : label} className="!text-fg-muted hover:!text-white">
          {copied ? <Check size={14} aria-hidden /> : <Copy size={14} aria-hidden />}
        </IconButton>
      </Box>
      <Box component="pre" className="p-4 font-mono text-xs leading-relaxed text-fg-body whitespace-pre-wrap m-0">
        {commands.flatMap((command) => command.split('\n')).map((line, i) => (
          <span key={i} className="block">
            {line.split(' ').map((token, j) => (
              <span key={`${i}-${j}`} className="inline-block whitespace-pre [overflow-wrap:anywhere]">
                {j > 0 ? ' ' : ''}
                {token}
              </span>
            ))}
          </span>
        ))}
      </Box>
    </Box>
  );
};
