import { Box, Stack } from '@mui/material';
import Link from 'next/link';
import type { FC } from 'react';
import {
  BodyText,
  Card,
  Doors,
  Grid,
  PageCard,
  PageHero,
  ProofList,
  RecordWindow,
  Section,
  SectionSubtitle,
  SectionTitle,
} from '@/components/marketing';
import { BUYER_DOORS } from '@/data/doors';
import { chapter } from '@/data/story';
import { sitePage } from '@/data/site-pages';
import { trustPage } from '@/data/trust';

/**
 * One template renders every Trust page from its record in src/data/trust.ts.
 * The order is the reviewer's order: who this page is for, the promise, the
 * proof (an illustrated record of what the product shows, beside the claims
 * it supports), the honesty statements (the platform's own grammar, said out
 * loud), then the doors: the two Trust pages to read next and, for the
 * reader who will never open a console, a demo. The template adds no
 * sentence of its own.
 */
export const TrustPage: FC<{ path: string }> = ({ path }) => {
  const page = sitePage(path);
  const record = trustPage(path);
  const ch = chapter(record.chapter);

  return (
    <main className="overflow-x-hidden">
      <PageHero eyebrow={{ label: 'Trust', href: '/trust' }} title={page.title} lede={record.lede ?? ch.claim} forWhom={record.forWhom}>
        <Doors {...BUYER_DOORS} className="justify-center mt-2" />
      </PageHero>

      <Section>
        <Box className="grid grid-cols-1 lg:grid-cols-12 gap-10 items-start max-w-6xl mx-auto">
          <Box className="lg:col-span-5 flex flex-col gap-4">
            <SectionTitle>What You Get</SectionTitle>
            <SectionSubtitle className="mt-0">Shipped behavior, stated as it works today. The record beside it is the shape of what the product stamps; its figures are examples.</SectionSubtitle>
            <ProofList items={record.points} />
            <Link href={record.seeIt.href} className="text-sm text-fg-secondary hover:text-white underline underline-offset-4 mt-2">
              {`${record.seeIt.label} \u2192`}
            </Link>
          </Box>
          <Box className="lg:col-span-7">
            <RecordWindow title={record.artifact.title} rows={record.artifact.rows} footer={record.artifact.footer} />
          </Box>
        </Box>
      </Section>

      <Section>
        <Box className="max-w-3xl mx-auto">
          <Box className="text-center mb-8 flex flex-col items-center">
            <SectionTitle>What We Will Not Say</SectionTitle>
            <SectionSubtitle className="mx-auto">The words this page holds itself to, so a reviewer does not have to find the over-claim.</SectionSubtitle>
          </Box>
          <Card hover={false}>
            <Stack component="ol" className="gap-4 list-decimal pl-5 m-0">
              {record.honesty.map((line) => (
                <BodyText key={line} component="li" className="text-fg">
                  {line}
                </BodyText>
              ))}
            </Stack>
          </Card>
        </Box>
      </Section>

      <Section>
        <Box className="max-w-3xl mx-auto text-center flex flex-col items-center gap-6">
          <SectionTitle>Read Next</SectionTitle>
          <Grid cols={2} className="w-full">
            {record.next.map((path) => (
              <PageCard key={path} path={path} linkLabel="What It Proves" />
            ))}
          </Grid>
          <Doors {...BUYER_DOORS} className="mt-4" />
        </Box>
      </Section>
    </main>
  );
};
