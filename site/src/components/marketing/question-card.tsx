/**
 * A question a visitor asks and its answer, as one card, with a door to the
 * page that answers it in full when the record names one. The words come
 * from the page's record (a persona's objections, the Compare page's
 * questions); this card adds none. The door's label is the target page's own
 * title from the route registry, never retyped.
 */
import { Box } from '@mui/material';
import Link from 'next/link';
import type { FC } from 'react';
import { Card } from './card';
import { BodyText, FeatureTitle } from './typography';
import { sitePage } from '@/data/site-pages';

export interface QuestionCardProps {
  question: string;
  answer: string;
  /** A registered route; the door reads its title. */
  readMore?: string;
}

export const QuestionCard: FC<QuestionCardProps> = ({ question, answer, readMore }) => {
  const proven = readMore ? sitePage(readMore) : undefined;
  return (
    <Card>
      <Box className="flex flex-col gap-3 h-full text-left">
        <FeatureTitle className="text-balance">{question}</FeatureTitle>
        <BodyText className="flex-1">{answer}</BodyText>
        {readMore && proven ? (
          <Link href={readMore} className="text-sm text-fg-secondary hover:text-white underline underline-offset-4">
            {`${proven.title} \u2192`}
          </Link>
        ) : null}
      </Box>
    </Card>
  );
};
