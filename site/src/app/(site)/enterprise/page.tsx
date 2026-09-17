import { Metadata } from 'next';
import { EnterpriseRedirect } from './redirect';

/**
 * The enterprise page lives at /pricing/enterprise — the website's routed
 * path prefixes cover /pricing, while a bare /enterprise falls through to
 * the console as an organization slug on the apex domain. This stub only
 * forwards old links on origins that serve the static site directly.
 */
export const metadata: Metadata = {
  title: 'Enterprise | Planton',
  robots: { index: false },
};

export default function EnterpriseMovedPage() {
  return <EnterpriseRedirect />;
}
