import fs from 'fs';
import path from 'path';
import { pageMetadata } from '@/lib/page-metadata';
import { LegalContent } from '@/components/legal/LegalContent';

export const metadata = pageMetadata('/legal/terms');

export default function TermsPage() {
  const content = fs.readFileSync(
    path.join(process.cwd(), 'content/legal/terms.md'),
    'utf-8',
  );

  return <LegalContent content={content} />;
}
