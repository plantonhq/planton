import { Box, Stack } from '@mui/material';
import type { FC } from 'react';
import { BodyText, CenteredCards, Doors, PageArtifact, PageCard, PageHero, ProofList, Section, SectionSubtitle, SectionTitle } from '@/components/marketing';
import { DESKTOP_LANDING_PATH } from '@/data/desktop-download';
import { distributionPage } from '@/data/distributions';
import { sitePage } from '@/data/site-pages';

/**
 * One template renders the hosted and self-hosted pages from their records in
 * src/data/distributions.ts (the desktop distribution has its own landing).
 * The order is the decider's order: the shape in two sentences and who
 * chooses it, what you get (the shape's own facts beside the record it
 * produces or the commands that install it), what stays yours in this shape
 * said plainly, then the other shapes and the doors. The template adds no
 * sentence of its own.
 */
export const DistributionPage: FC<{ path: string }> = ({ path }) => {
  const page = sitePage(path);
  const record = distributionPage(path);

  return (
    <main className="overflow-x-hidden">
      <PageHero eyebrow={{ label: 'Distributions', href: '/distributions' }} title={page.title} lede={record.lede} forWhom={record.forWhom}>
        <Doors {...record.doors} className="justify-center mt-2" />
      </PageHero>

      <Section>
        <Box className="grid grid-cols-1 lg:grid-cols-12 gap-10 items-start max-w-6xl mx-auto">
          <Box className="lg:col-span-5 flex flex-col gap-4">
            <SectionTitle>What You Get</SectionTitle>
            <SectionSubtitle className="mt-0">Shipped behavior, stated as it works today.</SectionSubtitle>
            <ProofList items={record.points} />
          </Box>
          <Box className="lg:col-span-7 lg:sticky lg:top-24">
            <PageArtifact artifact={record.artifact} />
          </Box>
        </Box>
      </Section>

      <Section>
        <Box className="max-w-3xl mx-auto">
          <Box className="text-center mb-8 flex flex-col items-center">
            <SectionTitle>What Stays Yours</SectionTitle>
            <SectionSubtitle className="mx-auto">In this shape, and in every other one.</SectionSubtitle>
          </Box>
          <Stack component="ul" className="gap-4 list-none p-0 m-0">
            {record.yours.map((line) => (
              <BodyText key={line} component="li" className="text-fg border-l-2 border-edge-hover pl-4">
                {line}
              </BodyText>
            ))}
          </Stack>
        </Box>
      </Section>

      <Section>
        <Box className="max-w-5xl mx-auto text-center flex flex-col items-center gap-6">
          <SectionTitle>The Other Shapes</SectionTitle>
          <CenteredCards className="w-full">
            {[...record.next, DESKTOP_LANDING_PATH].map((sibling) => (
              <PageCard key={sibling} path={sibling} linkLabel="Read More" />
            ))}
          </CenteredCards>
          <Doors {...record.doors} className="mt-4" />
        </Box>
      </Section>
    </main>
  );
};
