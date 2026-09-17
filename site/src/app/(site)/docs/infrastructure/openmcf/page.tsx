import { RedirectPage } from '@/components/common/redirect-page';

// Compatibility redirect: this documentation page was renamed to "open-source".
// The hardcoded segment out-ranks the docs [[...slug]] catch-all and is emitted
// as a static file, so the previous URL redirects instead of returning a 404.
export default function Page() {
  return <RedirectPage to="/docs/infrastructure/open-source" />;
}
