import { Box, Typography } from '@mui/material';
import Link from 'next/link';
import type { FC } from 'react';
import { PERSONAS, personaPagePath } from '@/data/personas';

/**
 * One line under the hero that routes each kind of reader to their page, so
 * the engineering leader, the consultancy, the founder, and the security
 * lead find their door on the first screen instead of at the ninth. Reads
 * from the persona records; each link is derived from the persona's slug.
 */
export const PersonaRouter: FC = () => (
  <Box className="flex flex-wrap items-center justify-center gap-x-5 gap-y-1 max-w-3xl">
    <Typography className="text-xs text-fg-muted">Read it for your role:</Typography>
    {PERSONAS.map((persona) => (
      <Link key={persona.slug} href={personaPagePath(persona.slug)} className="text-xs text-fg-secondary hover:text-white underline underline-offset-4">
        {persona.name}
      </Link>
    ))}
  </Box>
);
