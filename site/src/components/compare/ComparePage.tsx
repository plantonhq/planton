import { Box } from '@mui/material';
import Link from 'next/link';
import type { FC } from 'react';
import {
  BodyText,
  Doors,
  Grid,
  PageArtifact,
  PageCard,
  PageHero,
  ProofList,
  QuestionCard,
  Section,
  SectionSubtitle,
  SectionTitle,
} from '@/components/marketing';
import { artifactOf } from '@/data/artifacts';
import { COMPARE, type ComparisonCategory } from '@/data/compare';
import { POSITIONING } from '@/data/positioning';
import { sitePage } from '@/data/site-pages';

/**
 * The Compare page, rendered from src/data/compare.ts. The order is the
 * comparer's order: the promise and who it is for, with a route to each kind
 * of tool on the first screen; the one difference the page turns on, with the
 * record every deployment leaves beside it; each kind of tool in its own
 * section (what it does, what the reader keeps, what Planton does at the
 * same moment, and the proving page's record beside all of it); the
 * questions a person comparing asks; then two pages to read next and the
 * doors. Categories are described and never named, and there is no table
 * with a competitor column: chapter 11's own rule. The template adds no
 * sentence of its own.
 */
export const ComparePage: FC = () => {
  const difference = artifactOf(COMPARE.difference.provenAt);
  const differenceProven = sitePage(COMPARE.difference.provenAt);
  return (
    <main className="overflow-x-hidden">
      <PageHero eyebrow={{ label: 'Compare' }} title={COMPARE.headline} kicker={`${POSITIONING.umbrella.tagline}, beside the Tools You Already Run`} lede={COMPARE.lede} forWhom={COMPARE.forWhom}>
        <Doors {...COMPARE.doors} className="justify-center mt-2" />
        <Box className="flex flex-wrap justify-center gap-x-5 gap-y-1 mt-1">
          {COMPARE.categories.map((c) => (
            <Link key={c.id} href={`#${c.id}`} className="text-sm text-fg-secondary hover:text-white underline underline-offset-4">
              {`Beside ${c.shortTitle}`}
            </Link>
          ))}
        </Box>
      </PageHero>

      <Section>
        <Box className="grid grid-cols-1 lg:grid-cols-12 gap-x-10 gap-y-6 items-start max-w-6xl mx-auto">
          <Box className="lg:col-start-1 lg:col-span-5 flex flex-col gap-4">
            <SectionTitle>Where the Difference Is</SectionTitle>
            <SectionSubtitle className="mt-0">{COMPARE.difference.claim}</SectionSubtitle>
            <Link href={COMPARE.difference.provenAt} className="text-sm text-fg-secondary hover:text-white underline underline-offset-4">
              {`${differenceProven.title} \u2192`}
            </Link>
          </Box>
          <Box className="lg:col-start-6 lg:col-span-7">{difference ? <PageArtifact artifact={difference} /> : null}</Box>
        </Box>
      </Section>

      {COMPARE.categories.map((category, i) => (
        <CategorySection key={category.id} category={category} layout={i % 2 === 0 ? 'split-reverse' : 'split'} />
      ))}

      <Section>
        <Box className="max-w-6xl mx-auto flex flex-col items-center gap-8">
          <SectionTitle>You Will Ask</SectionTitle>
          <Grid cols={3} className="w-full">
            {COMPARE.questions.map((q) => (
              <QuestionCard key={q.question} question={q.question} answer={q.answer} readMore={q.readMore} />
            ))}
          </Grid>
        </Box>
      </Section>

      <Section>
        <Box className="max-w-3xl mx-auto text-center flex flex-col items-center gap-6">
          <SectionTitle>Read Next</SectionTitle>
          <Grid cols={2} className="w-full">
            {COMPARE.next.map((path) => (
              <PageCard key={path} path={path} linkLabel="Read More" />
            ))}
          </Grid>
          <Doors {...COMPARE.doors} className="mt-4" />
        </Box>
      </Section>
    </main>
  );
};

/**
 * One kind of tool. The text column: the category, what it does (body),
 * what the reader keeps when they run both (the section's largest sentence
 * after its title, because it is the one the reader came for), then what
 * Planton does at the same moment as labeled proof points, and the door to
 * the proving page. Beside it, that page's own record, so the proof is in
 * view and is the same illustration the proving page shows. The DOM order is
 * the phone order (the claim, the record, then the proof points and the
 * door, so the record is read before the door that leaves the section); grid
 * placement makes the two-column shape from lg, alternating sides so three
 * sections do not read as one template repeated (the chapter frame's own
 * move).
 */
const CategorySection: FC<{ category: ComparisonCategory; layout: 'split' | 'split-reverse' }> = ({ category, layout }) => {
  const proven = sitePage(category.provenAt);
  const artifact = artifactOf(category.provenAt);
  const textCol = layout === 'split-reverse' ? 'lg:col-start-8 lg:col-span-5' : 'lg:col-start-1 lg:col-span-5';
  const artCol = layout === 'split-reverse' ? 'lg:col-start-1 lg:col-span-7' : 'lg:col-start-6 lg:col-span-7';
  return (
    <Section id={category.id}>
      <Box className="grid grid-cols-1 lg:grid-cols-12 gap-x-10 gap-y-6 items-start max-w-6xl mx-auto">
        <Box className={`${textCol} lg:row-start-1 flex flex-col gap-4`}>
          <SectionTitle>{`Beside ${category.title}`}</SectionTitle>
          <BodyText>{category.theyDo}</BodyText>
          <BodyText className="text-base md:text-lg text-fg">{category.both}</BodyText>
        </Box>
        <Box className={`${artCol} lg:row-start-1 lg:row-span-2`}>{artifact ? <PageArtifact artifact={artifact} /> : null}</Box>
        <Box className={`${textCol} lg:row-start-2 flex flex-col gap-4`}>
          <ProofList items={category.planton} className="mt-0" />
          <Link href={category.provenAt} className="text-sm text-fg-secondary hover:text-white underline underline-offset-4">
            {`${proven.title} \u2192`}
          </Link>
        </Box>
      </Box>
    </Section>
  );
};
