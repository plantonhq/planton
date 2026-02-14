'use client';

import { FC } from 'react';
import Link from 'next/link';
import { ProviderIcon } from '@/app/docs/components/ProviderIcon';

/** Shape of a single provider entry passed from the server component. */
export interface CatalogProvider {
  /** Provider directory name (e.g. "aws", "gcp", "openfga"). */
  name: string;
  /** Relative docs path (e.g. "catalog/aws"). */
  path: string;
  /** Number of component pages under this provider. */
  componentCount: number;
}

interface CatalogProviderGridProps {
  providers: CatalogProvider[];
}

/**
 * Data-driven catalog provider grid.
 *
 * Receives provider metadata extracted from the docs structure at build time
 * and renders the 2-column card grid. Each card shows the provider icon
 * (with letter-badge fallback), the display name, and the component count.
 *
 * This component replaces the previously hardcoded HTML in catalog/index.md,
 * eliminating stale component counts and broken icon images.
 */
export const CatalogProviderGrid: FC<CatalogProviderGridProps> = ({ providers }) => {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-4 my-6">
      {providers.map((provider) => (
        <Link
          key={provider.name}
          href={`/docs/${provider.path}`}
          className="flex items-center gap-3 p-4 rounded-lg border border-purple-900/30 bg-slate-900/30 hover:bg-slate-800/50 transition-colors"
        >
          <ProviderIcon provider={provider.name} size={32} />
          <div>
            <div className="font-semibold text-white">
              {provider.name.toUpperCase()}
            </div>
            <div className="text-sm text-slate-400">
              {provider.componentCount} component{provider.componentCount !== 1 ? 's' : ''}
            </div>
          </div>
        </Link>
      ))}
    </div>
  );
};
