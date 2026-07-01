'use client';

import { FC, useState } from 'react';
import Image from 'next/image';
import { getProviderIconPath } from '@/app/docs/utils/providerIcons';

interface ProviderIconProps {
  /** Provider directory name (e.g. "aws", "gcp", "openfga"). */
  provider: string;
  /** Icon width and height in pixels. Defaults to 20. */
  size?: number;
  /** Additional CSS classes applied to the outer element. */
  className?: string;
}

/**
 * Renders a provider icon with a letter-badge fallback.
 *
 * Attempts to load the SVG resolved by `getProviderIconPath()`. If the image
 * fails to load (404, network error, etc.) a styled badge showing the first
 * letter of the provider name is displayed instead.
 *
 * Used by both the sidebar (for provider directories) and the catalog
 * provider grid (for provider cards).
 */
export const ProviderIcon: FC<ProviderIconProps> = ({
  provider,
  size = 20,
  className,
}) => {
  const [error, setError] = useState(false);

  const iconPath = getProviderIconPath(provider);
  const letter = provider.charAt(0).toUpperCase();

  if (error) {
    return (
      <span
        className={`flex items-center justify-center rounded bg-secondary font-bold text-muted-foreground flex-shrink-0 ${className ?? ''}`}
        style={{ width: size, height: size, fontSize: Math.round(size * 0.45) }}
        aria-label={provider.toUpperCase()}
      >
        {letter}
      </span>
    );
  }

  return (
    <Image
      src={iconPath}
      alt={provider.toUpperCase()}
      width={size}
      height={size}
      className={`object-contain ${className ?? ''}`}
      onError={() => setError(true)}
    />
  );
};
