import fs from 'fs';
import path from 'path';
import { pageMetadata } from '@/lib/page-metadata';
import { LegalContent } from '@/components/legal/LegalContent';

export const metadata = pageMetadata('/legal/refund-policy');

export default function RefundPolicyPage() {
  const content = fs.readFileSync(
    path.join(process.cwd(), 'content/legal/refund-policy.md'),
    'utf-8',
  );

  return <LegalContent content={content} />;
}
