import { Box } from '@mui/material';
import Link from 'next/link';
import type { FC } from 'react';
import { BodyText, Card, CenteredCards, ChapterSection, FeatureTitle } from '@/components/marketing';
import { chapter } from '@/data/story';
import { PERSONAS } from '@/data/personas';
import { PERSONA_ROUTES } from './PersonaRouter';
/**
 * Chapter 9. Two people are named first (the user and the one who signs),
 * then the five personas as doors. Persona copy is the persona record's own
 * framing; the pages behind the doors are today's solutions pages until the
 * persona pages replace them, so the doors point where the story says they
 * will and the link gate keeps them honest.
 */

const ch = chapter('who-it-is-for');

export const WhoItIsFor: FC = () => (
  <ChapterSection chapter={ch}>
    <CenteredCards className="max-w-6xl mx-auto">
      {PERSONAS.map((persona) => (
        <Card key={persona.slug}>
          <Box className="flex flex-col gap-3 h-full">
            <FeatureTitle>{persona.name}</FeatureTitle>
            <BodyText className="flex-1">{persona.wall}</BodyText>
            <Link href={PERSONA_ROUTES[persona.slug]} className="text-sm text-fg-secondary hover:text-white underline underline-offset-4">
              {'How Planton Fits Your Work \u2192'}
            </Link>
          </Box>
        </Card>
      ))}
    </CenteredCards>
  </ChapterSection>
);
