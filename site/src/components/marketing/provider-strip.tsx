/**
 * The providers the catalog covers, as a quiet row of brand marks. A mark
 * keeps its own colors (identification, never decoration); a provider the
 * site holds no mark for gets a neutral monogram tile of the same size, so
 * every item in the row has the same shape and the count on the page and the
 * row always agree.
 */
import { Box } from '@mui/material';
import Image from 'next/image';
import type { FC } from 'react';
import { CLOUD_PROVIDERS } from '@/data/platform-stats';
import { asset } from '@/lib/assets';

export const ProviderStrip: FC<{ className?: string }> = ({ className = '' }) => (
  <Box component="ul" className={`flex flex-wrap items-center justify-center gap-x-8 gap-y-4 list-none m-0 p-0 ${className}`}>
    {CLOUD_PROVIDERS.map((provider) => (
      <Box component="li" key={provider.name} className="flex items-center gap-2 text-sm text-fg-secondary">
        {'logo' in provider ? (
          <Image src={asset(`/images/providers/${provider.logo}`)} alt="" width={24} height={24} className="h-6 w-6 object-contain" />
        ) : (
          <span aria-hidden className="h-6 w-6 rounded-md bg-raised border border-edge text-xs font-semibold text-fg flex items-center justify-center">
            {provider.name[0]}
          </span>
        )}
        <span>{provider.name}</span>
      </Box>
    ))}
  </Box>
);
