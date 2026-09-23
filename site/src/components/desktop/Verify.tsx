import { Box, Typography } from '@mui/material';
import Link from 'next/link';
import type { FC } from 'react';
import { BodyText, Card, CommandBlock, Divider, FeatureTitle, Section } from '@/components/marketing';
import { DESKTOP_DOWNLOAD } from '@/data/desktop';
import { DESKTOP_LANDING_PATH, desktopChecksumsUrl, type DesktopPlatform } from '@/data/desktop-download';

/**
 * How to verify the downloaded file on the selected platform (a card with
 * the platform's own commands), how updates arrive (three sentences, so a
 * paragraph and not a card stretched beside a code block), then the doors
 * for a person who wants something else: the landing's argument on its own
 * line, and the CLI alone or the self-hosted edition on the next. The verify
 * commands are the platform's from the download data; the words are the
 * record's.
 */
export const Verify: FC<{ platform: DesktopPlatform; version: string | null }> = ({ platform, version }) => {
  const verify = DESKTOP_DOWNLOAD.verify(platform);
  const updates = DESKTOP_DOWNLOAD.updates(platform);
  const also = DESKTOP_DOWNLOAD.alsoConsider;
  const checksumsUrl = version ? desktopChecksumsUrl(version) : null;
  const linkClass = 'underline underline-offset-4 hover:text-white';
  return (
    <>
      <Section className="!py-0">
        <Divider />
      </Section>
      <Section>
        <Box className="max-w-3xl mx-auto flex flex-col gap-8">
          <Card hover={false}>
            <FeatureTitle className="mb-2">{verify.title}</FeatureTitle>
            <BodyText className="mb-4">
              {verify.lead}{' '}
              {checksumsUrl ? (
                <a href={checksumsUrl} className={linkClass}>
                  {verify.checksumsLabel}
                </a>
              ) : (
                verify.checksumsLabel
              )}
              .{verify.macNote ? ` ${verify.macNote}` : null}
            </BodyText>
            <CommandBlock commands={platform.verifyCommands} label={`Copy the ${platform.name} verification commands`} />
          </Card>

          <Box className="text-center">
            <FeatureTitle className="mb-2">{updates.title}</FeatureTitle>
            <BodyText>
              {updates.body}
              {updates.homebrew ? (
                <>
                  {' '}
                  On Homebrew, <code className="font-mono text-white">{updates.homebrew}</code> does the same.
                </>
              ) : null}{' '}
              {updates.tail}
            </BodyText>
          </Box>
        </Box>

        <Box className="text-center mt-12 flex flex-col items-center gap-2">
          <Link href={DESKTOP_LANDING_PATH} className={`text-sm text-fg-secondary ${linkClass}`}>
            {`${also.landingLabel} \u2192`}
          </Link>
          <Typography className="text-sm text-fg-muted max-w-2xl">
            {also.question}{' '}
            <Link href={also.cliHref} className={linkClass}>
              {also.cliLabel}
            </Link>
            , or{' '}
            <Link href={also.selfHostedHref} className={linkClass}>
              {also.selfHostedLabel}
            </Link>
            .
          </Typography>
        </Box>
      </Section>
    </>
  );
};
