/**
 * Download configuration — the single source for the primary CTA target and the
 * direct installer downloads.
 *
 * The installer URLs point at the `latest` channel aliases on the Planton
 * downloads CDN. The desktop release pipeline republishes these aliases on
 * every stable release, so the URLs are version-free and always serve the
 * current installer.
 *
 * Note: the marketing surface intentionally does NOT teach CLI install — the
 * desktop launch experience owns CLI setup. So there are no CLI install commands
 * here; the CLI is acknowledged only as a companion in the landing copy.
 */

/** The primary "Download Planton" CTA target (a real /download page on this site). */
export const DOWNLOAD_HREF = "/download";

/** Platforms the desktop app targets, shown next to the download CTA. */
export const OPERATING_SYSTEMS = ["macOS", "Linux", "Windows"] as const;

/** Base URL of the stable-channel installer aliases on the downloads CDN. */
export const DOWNLOADS_BASE = "https://downloads.planton.app/desktop/latest";

/** Direct installer downloads, one per platform. */
export const DOWNLOADS = [
  {
    os: "macOS",
    href: `${DOWNLOADS_BASE}/planton-desktop-universal-macos.dmg`,
    note: "Universal — Apple Silicon + Intel",
  },
  {
    os: "Linux",
    href: `${DOWNLOADS_BASE}/planton-desktop-linux-amd64.AppImage`,
    note: "AppImage — x86_64",
  },
  {
    os: "Windows",
    href: `${DOWNLOADS_BASE}/planton-desktop-windows-x64-setup.exe`,
    note: "Installer — x64",
  },
] as const;

/** Debian/Ubuntu package, offered as a secondary Linux option. */
export const LINUX_DEB_URL = `${DOWNLOADS_BASE}/planton-desktop-linux-amd64.deb`;
