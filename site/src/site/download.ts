/**
 * Download + CLI-install configuration — the single source for the primary CTA
 * target and the CLI install commands. Built to the intended final state;
 * wiring the target to a real artifact/CDN and publishing the Homebrew tap are
 * tracked "make it real" follow-ups.
 */

/** The primary "Download Planton" CTA target (a real /download page on this site). */
export const DOWNLOAD_HREF = "/download";

/** Platforms the desktop app targets, shown next to the download CTA. */
export const OPERATING_SYSTEMS = ["macOS", "Linux", "Windows"] as const;

export interface InstallMethod {
  id: string;
  /** Tab label, e.g. "Homebrew". */
  label: string;
  /** The copy-pasteable command. */
  command: string;
  /** One-line context under the command. */
  note?: string;
}

/** CLI install methods, rendered as tabs. First entry is the default. */
export const CLI_INSTALL_METHODS: InstallMethod[] = [
  {
    id: "homebrew",
    label: "Homebrew",
    command: "brew install plantonhq/tap/planton",
    note: "macOS and Linux",
  },
  {
    id: "go",
    label: "Go",
    command: "go install github.com/plantonhq/planton@latest",
    note: "Any platform with Go 1.24+",
  },
];
