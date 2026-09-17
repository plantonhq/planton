import { Box } from '@mui/material';
import type { FC } from 'react';
import { BodyText, FeatureTitle } from '@/components/marketing';

/**
 * A chapter's proof points as a compact list for the split layout: a short
 * label the section chooses over the sentence the chapter states. Sits in
 * the text column beside the artifact.
 */
export interface ProofItem {
  label: string;
  text: string;
}

export const ProofList: FC<{ items: readonly ProofItem[] }> = ({ items }) => (
  <Box className="flex flex-col gap-4 mt-2">
    {items.map((item) => (
      <Box key={item.label} className="border-l-2 border-edge-hover pl-4">
        <FeatureTitle className="text-sm md:text-base mb-1">{item.label}</FeatureTitle>
        <BodyText>{item.text}</BodyText>
      </Box>
    ))}
  </Box>
);
