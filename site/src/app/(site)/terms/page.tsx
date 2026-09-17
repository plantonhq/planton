import fs from 'fs';
import path from 'path';
import { Metadata } from 'next';
import { LegalContent } from '@/components/legal/LegalContent';

export const metadata: Metadata = {
  title: 'Terms of Service - Planton',
  description:
    'Terms of Service governing your use of the Planton infrastructure automation and service deployment platform.',
};

export default function TermsPage() {
  const content = fs.readFileSync(
    path.join(process.cwd(), 'content/legal/terms.md'),
    'utf-8',
  );

  return <LegalContent content={content} />;
}
