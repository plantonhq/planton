/**
 * A door to another page of the site: the page's own title and description
 * from the route registry, never retyped, and one link line whose words the
 * section chooses ("What It Proves", "Read More"). Index pages lay these out
 * as a centered set; a page's "Read Next" lays out two.
 */
import { Box } from '@mui/material';
import Link from 'next/link';
import type { FC } from 'react';
import { Card } from './card';
import { BodyText, FeatureTitle } from './typography';
import { sitePage } from '@/data/site-pages';

export interface PageCardProps {
  /** A registered route; the card reads its title and description. */
  path: string;
  /** Title Case; the link line, rendered with a trailing arrow. */
  linkLabel: string;
}

export const PageCard: FC<PageCardProps> = ({ path, linkLabel }) => {
  const page = sitePage(path);
  return (
    <Card>
      <Box className="flex flex-col gap-3 h-full text-left">
        <FeatureTitle className="text-balance">{page.title}</FeatureTitle>
        <BodyText className="flex-1">{page.description}</BodyText>
        <Link href={path} className="text-sm text-fg-secondary hover:text-white underline underline-offset-4">
          {`${linkLabel} \u2192`}
        </Link>
      </Box>
    </Card>
  );
};
