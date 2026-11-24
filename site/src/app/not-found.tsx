import { Typography } from '@mui/material';
import { Doors, PageHero } from '@/components/marketing';
import { MarketingShell } from '@/components/marketing/MarketingShell';
import { NOT_FOUND } from '@/data/not-found';

/**
 * The site's 404, inside the same header and footer as every page, so a
 * visitor holding an old link has a way on. Every word reads from its record.
 */
export default function NotFound() {
  return (
    <MarketingShell>
      <main className="min-h-[60vh]">
        <PageHero eyebrow={{ label: NOT_FOUND.eyebrow }} title={NOT_FOUND.title} lede={NOT_FOUND.line}>
          <Typography className="text-sm text-fg-secondary max-w-xl">{NOT_FOUND.sentence}</Typography>
          <Doors {...NOT_FOUND.doors} className="justify-center mt-2" />
        </PageHero>
      </main>
    </MarketingShell>
  );
}
