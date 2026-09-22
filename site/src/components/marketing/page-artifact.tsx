/**
 * The proof beside a page's claims, rendered from its record: either the
 * shape of a record the product shows (an illustrated RecordWindow whose
 * footer says so) or the commands a person actually types (a CommandBlock
 * with the documentation it is quoted from as its caption). The page's
 * record decides which; this component never chooses words of its own.
 */
import { Box, Typography } from '@mui/material';
import Link from 'next/link';
import type { FC } from 'react';
import type { Artifact } from '@/data/page-shapes';
import { CommandBlock } from './command-block';
import { RecordWindow } from './record-window';

export const PageArtifact: FC<{ artifact: Artifact; className?: string }> = ({ artifact, className = '' }) => {
  if (artifact.kind === 'record') {
    return <RecordWindow title={artifact.title} rows={artifact.rows} footer={artifact.footer} className={className} />;
  }
  return (
    <Box className={`flex flex-col gap-3 ${className}`}>
      {artifact.file ? (
        <CommandBlock commands={artifact.file.body.split('\n')} label={`Copy ${artifact.file.name}`} title={artifact.file.name} />
      ) : null}
      <CommandBlock commands={artifact.commands} label={artifact.label} title={artifact.title} />
      <Typography className="text-sm text-fg-secondary">
        {artifact.caption}{' '}
        <Link href={artifact.source.href} className="underline underline-offset-4 hover:text-white">
          {artifact.source.label}
        </Link>
        .
      </Typography>
    </Box>
  );
};
