/**
 * The first screen of a marketing page inside a section (Trust, Product,
 * Distributions): an eyebrow door back to the section's index, the page's
 * title, its lede (the chapter's claim, or the page's own opening sentence),
 * one quiet line saying who the page is for, and the doors. The hero adds no
 * sentence of its own; every string is passed in from the page's record.
 */
import { Stack, Typography } from '@mui/material';
import Link from 'next/link';
import type { FC, ReactNode } from 'react';
import { Section } from './section';

export interface PageHeroProps {
  /** The section this page belongs to: a door to its index, or, on the index itself, the section's name alone. */
  eyebrow?: { label: string; href?: string };
  title: string;
  /** Title Case; a short line under the title, e.g. a hub's one analogy. */
  kicker?: string;
  lede: string;
  /** Sentence case; who this page is written for and what it answers, in one line. */
  forWhom?: string;
  /** The doors, and anything else that sits under them. */
  children?: ReactNode;
}

export const PageHero: FC<PageHeroProps> = ({ eyebrow, title, kicker, lede, forWhom, children }) => (
  <Section>
    <Stack className="max-w-3xl mx-auto text-center items-center gap-4 pt-8">
      {eyebrow?.href ? (
        <Link href={eyebrow.href} className="text-xs font-medium tracking-wide text-fg-muted underline underline-offset-4 hover:text-white">
          {eyebrow.label}
        </Link>
      ) : eyebrow ? (
        <Typography className="text-xs font-medium tracking-wide text-fg-muted">{eyebrow.label}</Typography>
      ) : null}
      <Typography variant="h1" className="text-3xl md:text-5xl font-semibold text-white leading-[1.15] tracking-tight text-balance">
        {title}
      </Typography>
      {kicker ? <Typography className="text-sm md:text-base font-medium text-fg-secondary -mt-2">{kicker}</Typography> : null}
      <Typography className="text-base md:text-lg text-fg-secondary leading-relaxed">{lede}</Typography>
      {forWhom ? <Typography className="text-sm text-fg-secondary">{forWhom}</Typography> : null}
      {children}
    </Stack>
  </Section>
);
