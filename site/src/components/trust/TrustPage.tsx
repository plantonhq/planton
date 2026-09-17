import { Box, Stack, Typography } from '@mui/material';
import Link from 'next/link';
import type { FC } from 'react';
import {
  ArrowRightIcon,
  BodyText,
  Card,
  FeatureTitle,
  Grid,
  PrimaryButton,
  RecordWindow,
  Section,
  SectionSubtitle,
  SectionTitle,
  SecondaryButton,
} from '@/components/marketing';
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
      <Section>
        <Stack className="max-w-3xl mx-auto text-center items-center gap-4 pt-8">
          <Link href="/trust" className="text-xs font-medium tracking-wide text-fg-muted underline underline-offset-4 hover:text-white">
            Trust
          </Link>
          <Typography variant="h1" className="text-3xl md:text-5xl font-semibold text-white leading-[1.15] tracking-tight text-balance">
            {page.title}
          </Typography>
          <Typography className="text-base md:text-lg text-fg-secondary leading-relaxed">{record.lede ?? ch.claim}</Typography>
          <Typography className="text-sm text-fg-muted">{record.forWhom}</Typography>
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
        <Box className="grid grid-cols-1 lg:grid-cols-12 gap-10 items-start max-w-6xl mx-auto">
          <Box className="lg:col-span-5 flex flex-col gap-4">
            <SectionTitle>What You Get</SectionTitle>
            <SectionSubtitle className="mt-0">Shipped behavior, stated as it works today. The record beside it is the shape of what the product stamps; its figures are examples.</SectionSubtitle>
            <Box className="flex flex-col gap-4 mt-2">
              {record.points.map((point) => (
                <Box key={point.label} className="border-l-2 border-edge-hover pl-4">
                  <FeatureTitle className="text-sm md:text-base mb-1">{point.label}</FeatureTitle>
                  <BodyText>{point.text}</BodyText>
                </Box>
              ))}
            </Box>
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
            {record.next.map((door) => {
              const target = sitePage(door.href);
              return (
                <Card key={door.href}>
                  <Box className="flex flex-col gap-3 h-full text-left">
                    <FeatureTitle className="text-balance">{target.title}</FeatureTitle>
                    <BodyText className="flex-1">{target.description}</BodyText>
                    <Link href={door.href} className="text-sm text-fg-secondary hover:text-white underline underline-offset-4">
                      {'What It Proves \u2192'}
                    </Link>
                  </Box>
                </Card>
              );
            })}
          </Grid>
          <Stack direction={{ xs: 'column', sm: 'row' }} className="gap-3 items-center mt-4">
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
        </Box>
      </Section>
    </main>
  );
};
