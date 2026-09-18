/**
 * The frame a page gives one chapter of the story: the chapter's title as
 * the heading, its claim as the subtitle, the section's own composition
 * below, and an optional door to the page that tells the chapter in full.
 * The landing renders every chapter through it; a persona page renders the
 * chapters that person needs, in their order. Two layouts keep a page from
 * reading as one template repeated: `center` stacks heading and content;
 * `split` sets the heading and claim in a left column and the section's
 * artifact on the right, so a claim and its proof share one screen. The
 * words come from the chapter; this component adds none. Chapter numbers
 * are the story's internal order and never appear on the page.
 */
import { Box } from '@mui/material';
import Link from 'next/link';
import type { FC, ReactNode } from 'react';
import type { Chapter } from '@/data/story';
import { Section } from './section';
import { SectionSubtitle, SectionTitle } from './typography';

export interface ChapterSectionProps {
  chapter: Chapter;
  children?: ReactNode;
  /** A page that tells this chapter in full, rendered as the section's one link. */
  readMore?: { href: string; label: string };
  layout?: 'center' | 'split' | 'split-reverse';
  /** A different data sentence under the title when the chapter's claim already appears above on the page. */
  subtitle?: string;
  /** Extra content under the heading in the split layout (the proof sentences, typically). */
  aside?: ReactNode;
  className?: string;
}

const ReadMore: FC<{ readMore: ChapterSectionProps['readMore']; className?: string }> = ({ readMore, className }) =>
  readMore ? (
    <Box className={className}>
      <Link href={readMore.href} className="text-sm text-fg-secondary hover:text-white underline underline-offset-4">
        {`${readMore.label} \u2192`}
      </Link>
    </Box>
  ) : null;

export const ChapterSection: FC<ChapterSectionProps> = ({ chapter, children, readMore, layout = 'center', className, subtitle, aside }) => {
  if (layout === 'center') {
    return (
      <Section id={chapter.id} className={className}>
        <Box className="text-center mb-10 flex flex-col items-center">
          <SectionTitle>{chapter.title}</SectionTitle>
          <SectionSubtitle className="mx-auto">{subtitle ?? chapter.claim}</SectionSubtitle>
        </Box>
        {children}
        <ReadMore readMore={readMore} className="mt-8 text-center" />
      </Section>
    );
  }
  // Stacked (phones): heading, list, record, door, in reading order. Two
  // columns (lg): heading and list on one side, the record on the other, the
  // door under the list. The DOM order is the phone order; grid placement
  // makes the desktop shape.
  const textCol = layout === 'split-reverse' ? 'lg:col-start-8 lg:col-span-5' : 'lg:col-start-1 lg:col-span-5';
  const artCol = layout === 'split-reverse' ? 'lg:col-start-1 lg:col-span-7' : 'lg:col-start-6 lg:col-span-7';
  return (
    <Section id={chapter.id} className={className}>
      <Box className="grid grid-cols-1 lg:grid-cols-12 gap-x-10 gap-y-6 items-start max-w-6xl mx-auto">
        <Box className={`${textCol} lg:row-start-1 flex flex-col gap-4`}>
          <SectionTitle>{chapter.title}</SectionTitle>
          <SectionSubtitle className="mt-0">{subtitle ?? chapter.claim}</SectionSubtitle>
          {aside}
        </Box>
        <Box className={`${artCol} lg:row-start-1 lg:row-span-2`}>{children}</Box>
        <ReadMore readMore={readMore} className={`${textCol} lg:row-start-2`} />
      </Box>
    </Section>
  );
};
