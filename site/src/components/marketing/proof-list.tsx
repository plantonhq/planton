/**
 * A chapter's proof points as a compact labeled list: a short Title Case
 * label the page chooses over the sentence the chapter states. Sits in the
 * text column beside an illustrated record, or under a claim on its own.
 */
import { Box } from '@mui/material';
import type { FC } from 'react';
import { BodyText, FeatureTitle } from './typography';

export interface ProofItem {
  label: string;
  text: string;
}

export const ProofList: FC<{ items: readonly ProofItem[]; className?: string }> = ({ items, className = '' }) => (
  <Box className={`flex flex-col gap-4 mt-2 ${className}`}>
    {items.map((item) => (
      <Box key={item.label} className="border-l-2 border-edge-hover pl-4">
        <FeatureTitle className="text-sm md:text-base mb-1">{item.label}</FeatureTitle>
        <BodyText>{item.text}</BodyText>
      </Box>
    ))}
  </Box>
);
