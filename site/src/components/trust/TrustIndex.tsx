import { Box, Typography } from '@mui/material';
import type { FC } from 'react';
import { CenteredCards, Doors, PageCard, PageHero, RecordWindow, Section } from '@/components/marketing';
import { BUYER_DOORS } from '@/data/doors';
import { TRUST_PAGES } from '@/data/trust';
import { STORY_SPINE } from '@/data/story';
import { SITE_PAGES, sitePage } from '@/data/site-pages';

/**
 * The door to the Trust section: the story's spine as the promise, the five
 * pages as cards with the registry's own descriptions, and the two doors the
 * section offers the reader who will never open a console. For the person
 * who was forwarded a link and has five minutes.
 */
export const TrustIndex: FC = () => {
  const index = sitePage('/trust');
  const pages = SITE_PAGES.filter((p) => p.group === 'trust' && p.path !== '/trust');
  return (
    <main className="overflow-x-hidden">
      <PageHero
        title={index.title}
        lede={STORY_SPINE}
        forWhom="Five pages, one proof each. Written for the person who signs and the reviewer who reads for the over-claim."
      >
        <Doors {...BUYER_DOORS} className="justify-center mt-2" />
      </PageHero>
      <Section>
        <Box className="max-w-3xl lg:max-w-4xl mx-auto mb-12">
          <Typography className="text-xs text-fg-muted mb-3 text-center">The first proof: what Planton stamps before a deploy.</Typography>
          <RecordWindow title={TRUST_PAGES[0].artifact.title} rows={TRUST_PAGES[0].artifact.rows} footer={TRUST_PAGES[0].artifact.footer} />
        </Box>
        <CenteredCards className="max-w-6xl mx-auto">
          {pages.map((page) => (
            <PageCard key={page.path} path={page.path} linkLabel="What It Proves" />
          ))}
        </CenteredCards>
        <Doors {...BUYER_DOORS} className="justify-center mt-12" />
      </Section>
    </main>
  );
};
