import { Box, Stack, Typography } from '@mui/material';
import Link from 'next/link';
import type { FC } from 'react';
import { ArrowRightIcon, BodyText, Card, CenteredCards, FeatureTitle, PrimaryButton, RecordWindow, Section, SecondaryButton } from '@/components/marketing';
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
      <Section>
        <Stack className="max-w-3xl mx-auto text-center items-center gap-4 pt-8">
          <Typography variant="h1" className="text-3xl md:text-5xl font-semibold text-white leading-[1.15] tracking-tight">
            {index.title}
          </Typography>
          <Typography className="text-base md:text-lg text-fg-secondary leading-relaxed">{STORY_SPINE}</Typography>
          <Typography className="text-sm text-fg-muted">Five pages, one proof each. Written for the person who signs and the reviewer who reads for the over-claim.</Typography>
          <Stack direction={{ xs: 'column', sm: 'row' }} className="gap-3 items-center justify-center mt-2">
            <Link href="/book-demo">
              <PrimaryButton className="text-sm px-8 py-3">
                Book a Demo
                <ArrowRightIcon />
              </PrimaryButton>
            </Link>
            <Link href="/pricing">
              <SecondaryButton className="text-sm px-8 py-3">Pricing</SecondaryButton>
            </Link>
          </Stack>
        </Stack>
      </Section>
      <Section>
        <Box className="max-w-3xl lg:max-w-4xl mx-auto mb-12">
          <Typography className="text-xs text-fg-muted mb-3 text-center">The first proof: what Planton stamps before a deploy.</Typography>
          <RecordWindow title={TRUST_PAGES[0].artifact.title} rows={TRUST_PAGES[0].artifact.rows} footer={TRUST_PAGES[0].artifact.footer} />
        </Box>
        <CenteredCards className="max-w-6xl mx-auto">
          {pages.map((page) => (
            <Card key={page.path}>
              <Box className="flex flex-col gap-3 h-full">
                <FeatureTitle>{page.title}</FeatureTitle>
                <BodyText className="flex-1">{page.description}</BodyText>
                <Link href={page.path} className="text-sm text-fg-secondary hover:text-white underline underline-offset-4">
                  {'What It Proves \u2192'}
                </Link>
              </Box>
            </Card>
          ))}
        </CenteredCards>
        <Stack direction={{ xs: 'column', sm: 'row' }} className="gap-3 items-center justify-center mt-12">
          <Link href="/book-demo">
            <PrimaryButton className="text-sm px-8 py-3">
              Book a Demo
              <ArrowRightIcon />
            </PrimaryButton>
          </Link>
          <Link href="/pricing">
            <SecondaryButton className="text-sm px-8 py-3">Pricing</SecondaryButton>
          </Link>
        </Stack>
      </Section>
    </main>
  );
};
