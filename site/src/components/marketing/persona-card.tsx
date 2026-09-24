/**
 * A door to a persona's page: the person's name, the wall they hit in their
 * own words, and one link line. The landing's chapter 9 and the Solutions
 * index lay these out; the words are the persona record's, and the link is
 * derived from the slug so it cannot point anywhere but that person's page.
 */
import { Box } from '@mui/material';
import Link from 'next/link';
import type { FC } from 'react';
import { persona, personaPagePath, type PersonaSlug } from '@/data/personas';
import { Card } from './card';
import { BodyText, FeatureTitle } from './typography';

export const PersonaCard: FC<{ slug: PersonaSlug }> = ({ slug }) => {
  const p = persona(slug);
  return (
    <Card>
      <Box className="flex flex-col gap-3 h-full">
        <FeatureTitle>{p.name}</FeatureTitle>
        <BodyText className="flex-1">{p.wall}</BodyText>
        <Link href={personaPagePath(slug)} className="text-sm text-fg-secondary hover:text-white underline underline-offset-4">
          {'How Planton Fits Your Work \u2192'}
        </Link>
      </Box>
    </Card>
  );
};
