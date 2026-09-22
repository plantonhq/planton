import { Box } from '@mui/material';
import type { FC } from 'react';
import { Doors, Grid, PageCard, PageHero, ProofList, Section, SectionSubtitle, SectionTitle } from '@/components/marketing';
import { DESKTOP_LANDING_PATH } from '@/data/desktop-download';
import { START_DOORS } from '@/data/doors';
import { chapter } from '@/data/story';
import { SITE_PAGES } from '@/data/site-pages';

/**
 * The door to the Distributions section: chapter 8's claim as the promise,
 * the three shapes as cards with the registry's own descriptions, and what
 * every shape has in common (the chapter's proof). For the person asking
 * "where does this run" before anything else.
 */
export const DistributionsIndex: FC = () => {
  const runs = chapter('runs-where-you-decide');
  const shapes = SITE_PAGES.filter((p) => p.group === 'distributions' && p.path !== '/distributions' && p.path !== `${DESKTOP_LANDING_PATH}/download`);
  const points = [
    { label: 'Keyless Connections', text: runs.proof[0] },
    { label: 'Open Source, and the Exit Path', text: runs.proof[1] },
    { label: 'One Model on Every Shape', text: runs.proof[3] },
  ];
  return (
    <main className="overflow-x-hidden">
      <PageHero eyebrow={{ label: 'Distributions' }} title={runs.title} lede={runs.claim} forWhom="For the person deciding where the platform itself will run, before deciding anything else.">
        <Doors {...START_DOORS} className="justify-center mt-2" />
      </PageHero>
      <Section>
        <Grid cols={3} className="max-w-6xl mx-auto">
          {shapes.map((page) => (
            <PageCard key={page.path} path={page.path} linkLabel="Read More" />
          ))}
        </Grid>
      </Section>
      <Section>
        <Box className="max-w-3xl mx-auto flex flex-col items-center gap-4">
          <SectionTitle>The Same in Every Shape</SectionTitle>
          <SectionSubtitle className="mt-0 text-center">What does not change when the shape does.</SectionSubtitle>
          <ProofList items={points} className="w-full text-left" />
          <Doors {...START_DOORS} className="mt-8" />
        </Box>
      </Section>
    </main>
  );
};
