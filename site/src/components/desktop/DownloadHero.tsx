'use client';

import { Box, Typography } from '@mui/material';
import { sendGAEvent } from '@next/third-parties/google';
import Link from 'next/link';
import type { FC } from 'react';
import { BodyText, Card, CommandBlock, PageHero, PrimaryButton, SecondaryButton } from '@/components/marketing';
import { DESKTOP_DOWNLOAD } from '@/data/desktop';
import { DESKTOP_DOWNLOAD_PATH, DESKTOP_PLATFORMS, desktopChecksumsUrl, type DesktopArtifact, type DesktopPlatform, type DesktopPlatformId } from '@/data/desktop-download';
import { DESKTOP_RELEASE, formatDownloadSize } from '@/data/desktop-release';
import { sitePage } from '@/data/site-pages';

interface DownloadHeroProps {
  platform: DesktopPlatform;
  onSelect: (id: DesktopPlatformId) => void;
  version: string | null;
}

const recordDownload = (platform: DesktopPlatformId, artifact: DesktopArtifact, version: string | null) => {
  sendGAEvent('event', 'desktop_download', { platform, artifact: artifact.key, version: version ?? 'unknown' });
};

const artifactSize = (artifact: DesktopArtifact): string | null => {
  const size = DESKTOP_RELEASE.artifacts[artifact.key];
  return size ? formatDownloadSize(size.bytes) : null;
};

/** The platform switch: every platform listed, the detected or chosen one active. */
const PlatformTabs: FC<{ selected: DesktopPlatformId; onSelect: (id: DesktopPlatformId) => void }> = ({ selected, onSelect }) => (
  <Box role="tablist" aria-label="Platform" className="flex justify-center gap-2 mb-6">
    {DESKTOP_PLATFORMS.map((platform) => {
      const active = platform.id === selected;
      return (
        <button
          key={platform.id}
          role="tab"
          type="button"
          aria-selected={active}
          aria-controls={`platform-${platform.id}`}
          onClick={() => onSelect(platform.id)}
          className={`px-4 py-2 rounded-lg border text-sm font-medium transition-colors duration-200 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-white ${
            active ? 'border-white bg-raised text-white' : 'border-edge bg-transparent text-fg-secondary hover:border-edge-hover hover:text-white'
          }`}
        >
          {platform.name}
        </button>
      );
    })}
  </Box>
);

/**
 * The door for a platform whose installer is off right now: why, and the two
 * real ways forward with a link each, never a dead end.
 */
const UnavailableCard: FC<{ platform: DesktopPlatform; onSelect: (id: DesktopPlatformId) => void }> = ({ platform, onSelect }) => {
  const words = DESKTOP_DOWNLOAD.unavailable(platform);
  const alternatives = DESKTOP_PLATFORMS.filter((p) => p.available);
  return (
    <Card hover={false} className="max-w-2xl mx-auto !p-8 md:!p-10">
      <Box id={`platform-${platform.id}`} role="tabpanel" className="flex flex-col items-center text-center">
        <Typography component="h2" className="text-lg font-semibold text-white mb-2">
          {words.title}
        </Typography>
        <BodyText className="!text-base max-w-lg">{words.body}</BodyText>
        <Box className="flex flex-col sm:flex-row flex-wrap gap-3 justify-center mt-6">
          <Link href={words.cliDoor.href}>
            <PrimaryButton className="!whitespace-nowrap">{words.cliDoor.label}</PrimaryButton>
          </Link>
          {alternatives.map((alt) => (
            <SecondaryButton key={alt.id} className="!whitespace-nowrap" onClick={() => onSelect(alt.id)}>
              {`Switch to ${alt.name}`}
            </SecondaryButton>
          ))}
        </Box>
        <Typography className="text-xs text-fg-muted mt-5 max-w-md">{words.note}</Typography>
      </Box>
    </Card>
  );
};

/** The platform's installers, facts, install steps, and its one-command install where that is live. */
const PlatformCard: FC<{ platform: DesktopPlatform; version: string | null }> = ({ platform, version }) => {
  const [primary, ...secondary] = [...platform.artifacts.filter((a) => a.primary), ...platform.artifacts.filter((a) => !a.primary)];
  const primarySize = artifactSize(primary);
  const command = platform.installCommand;

  return (
    <Card hover={false} className="max-w-2xl mx-auto !p-8 md:!p-10">
      <Box id={`platform-${platform.id}`} role="tabpanel" className="flex flex-col items-center text-center">
        <PrimaryButton href={primary.href} onClick={() => recordDownload(platform.id, primary, version)} className="!px-8 !py-3 !text-base">
          {primary.label}
          {primarySize && <span className="ml-2 text-fg-muted font-normal">{primarySize}</span>}
        </PrimaryButton>

        {secondary.length > 0 && (
          <Box className="flex flex-wrap justify-center gap-3 mt-3">
            {secondary.map((artifact) => {
              const size = artifactSize(artifact);
              return (
                <SecondaryButton key={artifact.key} href={artifact.href} onClick={() => recordDownload(platform.id, artifact, version)}>
                  {artifact.label}
                  {size && <span className="ml-2 text-fg-muted font-normal">{size}</span>}
                </SecondaryButton>
              );
            })}
          </Box>
        )}

        <Typography component="ul" className="flex flex-wrap justify-center gap-x-4 gap-y-1 mt-5 text-xs text-fg-secondary list-none p-0">
          {platform.facts.map((fact) => (
            <li key={fact}>{fact}</li>
          ))}
        </Typography>

        <Box className="mt-6 flex flex-col items-center gap-3 max-w-lg w-full">
          {platform.installSteps.map((step) => (
            <Box key={step.text} className="flex flex-col items-center gap-2 w-full">
              <BodyText className="!text-[13px] text-fg-secondary">{step.text}</BodyText>
              {step.command && <CommandBlock compact commands={[step.command]} label={`Copy: ${step.command}`} />}
            </Box>
          ))}
        </Box>

        {command.status === 'live' && (
          // The installer is the door; the command is its equal for people who
          // live in a terminal, directly beneath it. Phones cannot run it.
          <Box className="hidden sm:block w-full max-w-xl mt-8 text-left">
            <Typography className="text-xs font-semibold text-white mb-2 text-center">{command.title}</Typography>
            <CommandBlock commands={command.lines} label={`Copy the ${platform.name} install command`} />
            <Typography className="text-xs text-fg-secondary mt-3 text-center">{command.note}</Typography>
          </Box>
        )}
      </Box>
    </Card>
  );
};

export const DownloadHero: FC<DownloadHeroProps> = ({ platform, onSelect, version }) => {
  const page = sitePage(DESKTOP_DOWNLOAD_PATH);
  const checksumsUrl = version ? desktopChecksumsUrl(version) : null;

  return (
    <>
      <PageHero eyebrow={{ label: 'Distributions', href: '/distributions' }} title={page.title} lede={DESKTOP_DOWNLOAD.lede}>
        {version && (
          <Typography className="text-xs text-fg-muted">
            Latest release <span className="font-mono text-fg-secondary">{version}</span>
            {checksumsUrl && (
              <>
                {' \u00b7 '}
                <a href={checksumsUrl} className="text-fg-secondary underline underline-offset-4 hover:text-white">
                  Checksums
                </a>
              </>
            )}
          </Typography>
        )}
      </PageHero>
      <Box className="max-w-4xl mx-auto px-4 -mt-10 mb-4">
        <PlatformTabs selected={platform.id} onSelect={onSelect} />
        {platform.available ? <PlatformCard platform={platform} version={version} /> : <UnavailableCard platform={platform} onSelect={onSelect} />}
      </Box>
    </>
  );
};
