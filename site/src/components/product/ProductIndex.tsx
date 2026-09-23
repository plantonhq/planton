import type { FC } from 'react';
import { Box, Typography } from '@mui/material';
import { CenteredCards, Doors, Metric, PageCard, PageHero, Section } from '@/components/marketing';
import { PLATFORM_COUNTS, PLATFORM_STATS } from '@/data/platform-stats';
import { START_DOORS } from '@/data/doors';
import { chapter } from '@/data/story';
import { SITE_PAGES } from '@/data/site-pages';

/**
 * The door to the Product section: the umbrella as the headline, its sentence
 * as the promise, and the product pages as cards with the registry's own
 * descriptions, in the order a platform engineer meets them. For the developer who clicked "Product" to find out what this
 * actually is.
 */
export const ProductIndex: FC = () => {
  const what = chapter('what-planton-is');
  const pages = SITE_PAGES.filter((p) => p.group === 'product' && p.path !== '/product');
  return (
    <main className="overflow-x-hidden">
      <PageHero eyebrow={{ label: 'Product' }} title={what.title} lede={what.claim} forWhom="For the platform engineer deciding what to hand their team, and the developer who wants to see what their agent gets.">
        <Doors {...START_DOORS} className="justify-center mt-2" />
        <Box className="grid grid-cols-2 sm:grid-cols-4 gap-6 mt-10 w-full">
          <Metric value={PLATFORM_STATS.DEPLOYMENT_MODULE_COUNT} label="Component Kinds" />
          <Metric value={PLATFORM_STATS.CLOUD_PROVIDER_COUNT} label="Providers" />
          <Metric value={PLATFORM_STATS.INFRA_CHART_COUNT} label="Infra Charts" />
          <Metric value={PLATFORM_STATS.IN_PRODUCTION_SINCE} label="In Production Since" />
        </Box>
        <Typography className="text-xs text-fg-muted">Kinds, providers, and charts counted from the open-source repository on {PLATFORM_COUNTS.countedOn}.</Typography>
      </PageHero>
      <Section>
        <CenteredCards className="max-w-6xl mx-auto">
          {pages.map((page) => (
            <PageCard key={page.path} path={page.path} linkLabel="Read More" />
          ))}
        </CenteredCards>
        <Doors {...START_DOORS} className="justify-center mt-12" />
      </Section>
    </main>
  );
};
