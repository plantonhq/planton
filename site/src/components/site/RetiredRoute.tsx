'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { retiredRoute } from '@/data/retired-routes';

/**
 * What a retired path renders. It forwards the visitor to the live page and,
 * for the moment before that happens (or for a browser with scripts off),
 * says in one line where the page went and offers the link. The destination
 * comes from src/data/retired-routes.ts, the same list the edge redirect and
 * the sitemap read, so the three cannot disagree.
 *
 * On the apex domain the edge answers with a true redirect before this ever
 * renders; this component is the fallback for any origin that serves the
 * export directly.
 */
export function RetiredRoute({ from }: { from: string }) {
  const { to } = retiredRoute(from);
  const router = useRouter();

  useEffect(() => {
    router.replace(to);
  }, [router, to]);

  return (
    <section className="min-h-[60vh] flex flex-col items-center justify-center gap-2 px-4 text-center bg-canvas">
      <p className="text-sm text-fg-secondary">This page has moved.</p>
      <Link href={to} className="text-sm text-white underline underline-offset-4">
        Continue to the new page
      </Link>
    </section>
  );
}
