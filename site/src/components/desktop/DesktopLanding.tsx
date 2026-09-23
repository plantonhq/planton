import { Box, Typography } from '@mui/material';
import Link from 'next/link';
import type { FC } from 'react';
import {
  BodyText,
  Card,
  CenteredCards,
  CommandBlock,
  Divider,
  Doors,
  FeatureTitle,
  Grid,
  Metric,
  PageCard,
  PageHero,
  PlatformCounts,
  Section,
  SectionSubtitle,
  SectionTitle,
} from '@/components/marketing';
import { DESKTOP_LANDING } from '@/data/desktop';
import { DESKTOP_LANDING_PATH, desktopWaysLine } from '@/data/desktop-download';
import { DESKTOP_SCREENSHOTS } from '@/data/desktop-screenshots';
import { sitePage } from '@/data/site-pages';
import { Screenshot } from './Screenshot';

/**
 * Planton Desktop's landing, rendered from src/data/desktop.ts. The order is
 * the reader's: what this is and that it is free, the two ways they already
 * do this and the honest third, the two ways to ask, what actually runs on
 * the machine, each hub on the laptop (the measured record beside Service
 * Hub, with its provenance printed), the same Planton everywhere, why it is
 * free, three pages to read next, and the close. The template adds no
 * sentence of its own; a screenshot slot with no real capture renders
 * nothing rather than a placeholder.
 */
export const DesktopLanding: FC = () => {
  const page = sitePage(DESKTOP_LANDING_PATH);
  const d = DESKTOP_LANDING;
  return (
    <main className="overflow-x-hidden">
      <PageHero eyebrow={{ label: 'Distributions', href: '/distributions' }} title={d.headline} kicker={page.title} lede={d.lede} forWhom={desktopWaysLine()}>
        <Doors {...d.doors} className="justify-center mt-2" />
        <Screenshot shot={DESKTOP_SCREENSHOTS.home} priority className="max-w-4xl mx-auto mt-8" />
      </PageHero>
      <Box className="max-w-5xl mx-auto -mt-10 mb-4 px-4">
        <PlatformCounts />
      </Box>

      <Section id="two-ways">
        <Box className="max-w-5xl mx-auto">
          <Box className="text-center mb-10">
            <SectionTitle>{d.twoWays.title}</SectionTitle>
          </Box>
          <Grid cols={2} className="mb-8">
            {d.twoWays.ways.map((way) => (
              <Card key={way.title} hover={false}>
                <FeatureTitle className="mb-1">{way.title}</FeatureTitle>
                <Typography className="text-sm text-fg-secondary mb-3">{way.trade}</Typography>
                <BodyText>{way.body}</BodyText>
              </Card>
            ))}
          </Grid>
          <Box className="text-center max-w-3xl mx-auto">
            <Typography className="text-base md:text-lg text-white font-medium">{d.twoWays.turn}</Typography>
            <Typography className="text-sm text-fg-muted mt-3">{d.twoWays.concession}</Typography>
          </Box>
        </Box>
      </Section>

      <Section id="two-ways-to-ask">
        <Box className="max-w-5xl mx-auto">
          <Box className="text-center mb-10">
            <SectionTitle>{d.ask.title}</SectionTitle>
            <SectionSubtitle className="mx-auto">{d.ask.lede}</SectionSubtitle>
          </Box>
          <Grid cols={2} gap="lg">
            <Card hover={false} className="!p-6 md:!p-8">
              <FeatureTitle className="mb-3">{d.ask.agent.title}</FeatureTitle>
              <BodyText className="mb-4">{d.ask.agent.body}</BodyText>
              {/* A phone cannot run the command; the sentence above still says what it does. */}
              <CommandBlock commands={[d.ask.agent.command]} label="Copy the skills install command" className="hidden sm:block" />
            </Card>
            <Card hover={false} className="!p-6 md:!p-8">
              <FeatureTitle className="mb-3">{d.ask.assistant.title}</FeatureTitle>
              {d.ask.assistant.body.map((sentence) => (
                <BodyText key={sentence} className="mb-4">
                  {sentence}
                </BodyText>
              ))}
              <Screenshot shot={DESKTOP_SCREENSHOTS.chooser} className="mt-6" />
            </Card>
          </Grid>
          <Box className="mt-12">
            <Typography component="h3" className="text-center text-base md:text-lg font-semibold text-white mb-6">
              {d.ask.addsTitle}
            </Typography>
            <Box component="ul" className="grid grid-cols-1 md:grid-cols-2 gap-x-10 gap-y-3 list-none p-0 m-0 max-w-4xl mx-auto">
              {d.ask.adds.map((point) => (
                <Box component="li" key={point} className="flex items-start gap-3">
                  <Box className="mt-[7px] w-1.5 h-1.5 rounded-full bg-edge-hover shrink-0" aria-hidden />
                  <BodyText>{point}</BodyText>
                </Box>
              ))}
            </Box>
          </Box>
          <Typography className="text-center text-sm text-fg-secondary mt-8 max-w-3xl mx-auto">{d.ask.close}</Typography>
        </Box>
      </Section>

      <Section id="what-runs">
        <Box className="max-w-5xl mx-auto">
          <Box className="text-center mb-10">
            <SectionTitle>{d.runs.title}</SectionTitle>
            <SectionSubtitle className="mx-auto">{d.runs.lede}</SectionSubtitle>
          </Box>
          <Grid cols={4} gap="sm">
            {d.runs.tiles.map((tile) => (
              <Card key={tile.title} hover={false}>
                <FeatureTitle className="mb-2 !text-base">{tile.title}</FeatureTitle>
                <BodyText className="!text-[13px]">{tile.body}</BodyText>
              </Card>
            ))}
          </Grid>
        </Box>
      </Section>

      <Section id="infra-hub">
        <Box className="max-w-5xl mx-auto">
          <Box className="text-center mb-10">
            <Typography className="text-xs font-medium tracking-wide text-fg-muted mb-3">{d.infraHub.kicker}</Typography>
            <SectionTitle>{d.infraHub.title}</SectionTitle>
            <SectionSubtitle className="mx-auto !max-w-3xl">{d.infraHub.lede}</SectionSubtitle>
          </Box>
          <Screenshot shot={DESKTOP_SCREENSHOTS.studio} className="max-w-4xl mx-auto mb-10" />
          <Grid cols={3}>
            {d.infraHub.points.map((point) => (
              <Card key={point.title} hover={false}>
                <FeatureTitle className="mb-2 !text-base">{point.title}</FeatureTitle>
                <BodyText className="!text-[13px]">{point.body}</BodyText>
                {point.door ? (
                  <Link href={point.door.href} className="inline-block mt-3 text-sm text-fg-secondary hover:text-white underline underline-offset-4">
                    {`${point.door.label} \u2192`}
                  </Link>
                ) : null}
              </Card>
            ))}
          </Grid>
        </Box>
      </Section>

      <Section id="service-hub">
        <Box className="max-w-5xl mx-auto">
          <Box className="text-center mb-10">
            <Typography className="text-xs font-medium tracking-wide text-fg-muted mb-3">{d.serviceHub.kicker}</Typography>
            <SectionTitle>{d.serviceHub.title}</SectionTitle>
            <SectionSubtitle className="mx-auto !max-w-3xl">{d.serviceHub.lede}</SectionSubtitle>
          </Box>
          <Box className="grid grid-cols-2 md:grid-cols-4 gap-6 max-w-3xl mx-auto mb-3">
            {d.serviceHub.measured.figures.map((figure) => (
              <Metric key={figure.label} value={figure.value} label={figure.label} />
            ))}
          </Box>
          <BodyText className="text-center text-xs text-fg-muted mb-10 max-w-2xl mx-auto">{d.serviceHub.measured.provenance}</BodyText>
          <Grid cols={3}>
            {d.serviceHub.points.map((point) => (
              <Card key={point.title} hover={false}>
                <FeatureTitle className="mb-2 !text-base">{point.title}</FeatureTitle>
                <BodyText className="!text-[13px]">{point.body}</BodyText>
              </Card>
            ))}
          </Grid>
          <BodyText className="text-center mt-8 !text-[13px] text-fg-secondary max-w-2xl mx-auto">{d.serviceHub.dockerNote}</BodyText>
        </Box>
      </Section>

      <Section className="!py-0">
        <Divider />
      </Section>
      <Section id="everywhere">
        <Box className="max-w-3xl mx-auto text-center">
          <SectionTitle>{d.everywhere.title}</SectionTitle>
          <BodyText className="mt-5 !text-base text-white">{d.everywhere.body}</BodyText>
          <BodyText className="mt-4 !text-base text-white">{d.everywhere.exit}</BodyText>
          <Typography className="text-sm text-fg-muted mt-6">{d.everywhere.umbrella}</Typography>
        </Box>
      </Section>

      <Section id="why-free">
        <Box className="max-w-2xl mx-auto">
          <SectionTitle className="text-center">{d.whyFree.title}</SectionTitle>
          {d.whyFree.body.map((paragraph) => (
            <BodyText key={paragraph} className="mt-5 !text-base">
              {paragraph}
            </BodyText>
          ))}
          <Link href={d.whyFree.door.href} className="inline-block mt-5 text-sm text-fg-secondary hover:text-white underline underline-offset-4">
            {`${d.whyFree.door.label} \u2192`}
          </Link>
        </Box>
      </Section>

      <Section>
        <Box className="max-w-5xl mx-auto text-center flex flex-col items-center gap-6">
          <SectionTitle>Read Next</SectionTitle>
          <CenteredCards className="w-full">
            {d.next.map((path) => (
              <PageCard key={path} path={path} linkLabel="Read More" />
            ))}
          </CenteredCards>
        </Box>
      </Section>

      <Section className="!py-0">
        <Divider />
      </Section>
      <Section>
        <Card hover={false} className="!p-8 md:!p-12 text-center max-w-3xl mx-auto">
          <SectionTitle>{d.close.title}</SectionTitle>
          <BodyText className="!text-base mx-auto max-w-xl mt-4 mb-8">{d.close.body}</BodyText>
          <Doors {...d.doors} className="justify-center" />
          <Link href={d.close.guide.href} className="inline-block mt-5 text-sm text-fg-secondary hover:text-white underline underline-offset-4">
            {`${d.close.guide.label} \u2192`}
          </Link>
        </Card>
      </Section>
    </main>
  );
};
