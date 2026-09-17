import { Box, Typography } from '@mui/material';
import Link from 'next/link';
import type { FC } from 'react';
import { PERSONAS } from '@/data/personas';

/**
 * One line under the hero that routes each kind of reader to their page, so
 * the engineering leader, the consultancy, the founder, and the security
 * lead find their door on the first screen instead of at the ninth. Reads
 * from the persona records; the routes are today's pages until the persona
 * pages replace them.
 */
export const PERSONA_ROUTES: Record<string, string> = {
  'platform-engineer': '/solutions/by-role/platform-engineers',
  'engineering-leader': '/solutions/by-role/engineering-leader',
  'it-consultancy': '/solutions/by-size/growing-teams',
  'startup-founder': '/solutions/by-role/startup-founders',
  'security-and-governance-leader': '/solutions/by-size/enterprises',
};

export const PersonaRouter: FC = () => (
  <Box className="flex flex-wrap items-center justify-center gap-x-5 gap-y-1 max-w-3xl">
    <Typography className="text-xs text-fg-muted">Read it for your role:</Typography>
    {PERSONAS.map((persona) => (
      <Link key={persona.slug} href={PERSONA_ROUTES[persona.slug]} className="text-xs text-fg-secondary hover:text-white underline underline-offset-4">
        {persona.name}
      </Link>
    ))}
  </Box>
);
