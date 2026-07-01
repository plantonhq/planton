/**
 * Download configuration — the single source for the primary CTA target and the
 * platforms the desktop app targets. Wiring the target to a real artifact/CDN is
 * a tracked "make it real" follow-up.
 *
 * Note: the marketing surface intentionally does NOT teach CLI install — the
 * desktop launch experience owns CLI setup. So there are no CLI install commands
 * here; the CLI is acknowledged only as a companion in the landing copy.
 */

/** The primary "Download Planton" CTA target (a real /download page on this site). */
export const DOWNLOAD_HREF = "/download";

/** Platforms the desktop app targets, shown next to the download CTA. */
export const OPERATING_SYSTEMS = ["macOS", "Linux", "Windows"] as const;
