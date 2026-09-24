'use client';

import { type FC, useEffect, useState } from 'react';
import { DESKTOP_PLATFORMS, DOWNLOADS_LATEST, type DesktopPlatformId } from '@/data/desktop-download';
import { DESKTOP_RELEASE } from '@/data/desktop-release';
import { AfterInstall } from './AfterInstall';
import { DownloadHero } from './DownloadHero';
import { Verify } from './Verify';
import { useDetectedPlatform } from './use-detected-platform';

/**
 * The version the page was built with is the floor; once the downloads
 * bucket permits a cross-origin read, the browser refreshes it so the badge
 * never lags a release by a site deploy. Silent on failure: the badge simply
 * keeps the build-time value, or stays absent.
 */
function useLatestVersion(): string | null {
  const [version, setVersion] = useState<string | null>(DESKTOP_RELEASE.version);
  useEffect(() => {
    const controller = new AbortController();
    fetch(`${DOWNLOADS_LATEST}/version.txt`, { signal: controller.signal, cache: 'no-store' })
      .then((r) => (r.ok ? r.text() : Promise.reject(new Error(String(r.status)))))
      .then((text) => {
        const fresh = text.trim();
        if (/^v\d+\.\d+\.\d+/.test(fresh)) setVersion(fresh);
      })
      .catch(() => {});
    return () => controller.abort();
  }, []);
  return version;
}

/**
 * The download page, rendered from src/data/desktop.ts and the platform list
 * in src/data/desktop-download.ts. One platform selection for the whole page:
 * the hero's tabs change it, and every section below follows (the install
 * steps, the CLI note, the verification commands are the selected platform's,
 * never macOS's by default). The detected platform wins until the person
 * picks; a phone and the prerender both fall back to the first listed. A
 * platform with no installer gets its door in the hero and none of the
 * "after you install" story it cannot yet begin.
 */
export const DesktopDownload: FC = () => {
  const detected = useDetectedPlatform();
  const [override, setOverride] = useState<DesktopPlatformId | null>(null);
  const version = useLatestVersion();

  const selectedId: DesktopPlatformId = override ?? detected ?? DESKTOP_PLATFORMS[0].id;
  const platform = DESKTOP_PLATFORMS.find((p) => p.id === selectedId) ?? DESKTOP_PLATFORMS[0];

  return (
    <main className="overflow-x-hidden">
      <DownloadHero platform={platform} onSelect={setOverride} version={version} />
      {platform.available && (
        <>
          <AfterInstall platform={platform} />
          <Verify platform={platform} version={version} />
        </>
      )}
    </main>
  );
};
