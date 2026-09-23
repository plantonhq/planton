import fs from 'fs';
import path from 'path';
import { pageMetadata } from '@/lib/page-metadata';
import { LegalContent } from '@/components/legal/LegalContent';

export const metadata = pageMetadata('/legal/privacy');

export default function PrivacyPage() {
  const content = fs.readFileSync(
    path.join(process.cwd(), 'content/legal/privacy.md'),
    'utf-8',
  );

  return <LegalContent content={content} />;
}
