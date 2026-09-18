'use client';

/**
 * Several things a person might paste, of which they pick one: a row of tabs
 * over a `CommandBlock`, with one teaching line under the block saying what
 * the chosen commands do. Use this when a page offers alternatives (the
 * follow-ons after an install: the agent, the CLI, the build cluster) and
 * `CommandBlock` alone when there is exactly one thing to paste. The words
 * come from the page's record; this component adds none.
 *
 * A client module because the selected tab is state; the copy affordance is
 * `CommandBlock`'s own.
 */
import { Box } from '@mui/material';
import { type FC, useState } from 'react';
import { CommandBlock } from './command-block';
import { BodyText } from './typography';

export interface CommandTabsProps {
  tabs: readonly { label: string; commands: readonly string[]; description: string }[];
  /** Screen-reader name for the copy button, e.g. "Copy the selected commands". */
  copyLabel: string;
  className?: string;
}

export const CommandTabs: FC<CommandTabsProps> = ({ tabs, copyLabel, className = '' }) => {
  const [selected, setSelected] = useState(0);
  const tab = tabs[selected] ?? tabs[0];
  return (
    <Box className={className}>
      <Box role="tablist" className="flex flex-wrap gap-2 mb-3">
        {tabs.map((t, i) => {
          const active = i === selected;
          return (
            <button
              key={t.label}
              role="tab"
              type="button"
              aria-selected={active}
              onClick={() => setSelected(i)}
              className={`px-3 py-1.5 rounded-lg border text-xs font-medium transition-colors focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-white ${
                active ? 'border-white bg-raised text-white' : 'border-edge bg-transparent text-fg-secondary hover:border-edge-hover hover:text-white'
              }`}
            >
              {t.label}
            </button>
          );
        })}
      </Box>
      <Box role="tabpanel">
        <CommandBlock commands={tab.commands} label={copyLabel} />
        <BodyText className="mt-3">{tab.description}</BodyText>
      </Box>
    </Box>
  );
};
