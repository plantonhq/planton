import { Box, Typography } from '@mui/material';
import type { FC } from 'react';
import { BodyText, Card, CommandTabs, FeatureTitle, Grid, Section, SectionSubtitle, SectionTitle } from '@/components/marketing';
import { DESKTOP_DOWNLOAD } from '@/data/desktop';
import type { DesktopPlatform } from '@/data/desktop-download';

/**
 * What happens after the download finishes, the part download pages skip:
 * the three beats a person will see on first launch, then the three
 * follow-ons as tabs whose commands are only what they type, with the
 * teaching under the block. Every word is the record's; the CLI note is the
 * selected platform's.
 */
export const AfterInstall: FC<{ platform: DesktopPlatform }> = ({ platform }) => {
  const words = DESKTOP_DOWNLOAD.afterInstall;
  return (
    <Section id="after-you-install">
      <Box className="max-w-5xl mx-auto">
        <Box className="text-center mb-10">
          <SectionTitle>{words.title}</SectionTitle>
          <SectionSubtitle className="mx-auto">{words.lede}</SectionSubtitle>
        </Box>

        <Grid cols={3} className="mb-12">
          {words.beats.map((beat, i) => (
            <Card key={beat.title} hover={false}>
              <Typography className="text-xs font-mono text-fg-muted mb-3">{String(i + 1).padStart(2, '0')}</Typography>
              <FeatureTitle className="mb-2">{beat.title}</FeatureTitle>
              <BodyText>{beat.body}</BodyText>
            </Card>
          ))}
        </Grid>

        {/* The follow-ons are commands a phone cannot run; the three first-launch beats above are the phone's whole story. */}
        <Box className="hidden sm:block">
          <Box className="text-center mb-6">
            <Typography component="h3" className="text-base md:text-lg font-semibold text-white">
              {words.thenTitle}
            </Typography>
            <Typography className="text-sm text-fg-secondary mt-2 max-w-2xl mx-auto">{words.thenLede}</Typography>
          </Box>
          <CommandTabs tabs={DESKTOP_DOWNLOAD.followOns(platform)} copyLabel="Copy the selected commands" className="max-w-3xl mx-auto" />
        </Box>
      </Box>
    </Section>
  );
};
