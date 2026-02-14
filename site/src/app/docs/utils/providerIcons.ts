/**
 * Centralized provider icon resolution.
 *
 * Convention: provider icon files live at `/images/providers/{provider}.svg`.
 * When a provider's filename doesn't match its directory name, an override
 * entry maps the directory name to the correct filename.
 *
 * This module is the single source of truth for provider icon paths, shared
 * by both the sidebar and the catalog provider grid.
 */

/**
 * Override map for providers whose SVG filename differs from their directory name.
 * Only non-standard names need an entry here; everything else uses the convention.
 */
const PROVIDER_ICON_OVERRIDES: Record<string, string> = {
  atlas: '/images/providers/mongodb-atlas.svg',
  digitalocean: '/images/providers/digital-ocean.svg',
};

/**
 * Resolve the icon path for a given provider directory name.
 *
 * Checks the override map first, then falls back to the naming convention
 * `/images/providers/{provider}.svg`.  The returned path may or may not
 * exist on disk — callers should handle load errors gracefully (e.g. with
 * a letter-badge fallback).
 */
export function getProviderIconPath(provider: string): string {
  return PROVIDER_ICON_OVERRIDES[provider] ?? `/images/providers/${provider}.svg`;
}
