import type { FC } from 'react';
import { CenteredCards, ChapterSection, PersonaCard } from '@/components/marketing';
import { chapter } from '@/data/story';
import { PERSONAS } from '@/data/personas';

/**
 * Chapter 9. Two people are named first (the user and the one who signs),
 * then the five personas as doors to their pages. Persona copy is the
 * persona record's own framing; this section adds none.
 */

const ch = chapter('who-it-is-for');

export const WhoItIsFor: FC = () => (
  <ChapterSection chapter={ch}>
    <CenteredCards className="max-w-6xl mx-auto">
      {PERSONAS.map((persona) => (
        <PersonaCard key={persona.slug} slug={persona.slug} />
      ))}
    </CenteredCards>
  </ChapterSection>
);
