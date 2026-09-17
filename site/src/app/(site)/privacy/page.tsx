import fs from 'fs';
import path from 'path';
import { Metadata } from 'next';
import { LegalContent } from '@/components/legal/LegalContent';

export const metadata: Metadata = {
  title: 'Privacy Policy - Planton',
  description:
    'Learn how Planton collects, uses, and protects your personal data when you use our infrastructure automation and service deployment platform.',
};

export default function PrivacyPage() {
  const content = fs.readFileSync(
    path.join(process.cwd(), 'content/legal/privacy.md'),
    'utf-8',
  );

  return <LegalContent content={content} />;
}
