import fs from 'fs';
import path from 'path';
import { Metadata } from 'next';
import { LegalContent } from '@/components/legal/LegalContent';

export const metadata: Metadata = {
  title: 'Refund Policy - Planton',
  description:
    'How refunds work for Planton team subscriptions, self-hosted licenses, and prepaid AI credits.',
};

export default function RefundPolicyPage() {
  const content = fs.readFileSync(
    path.join(process.cwd(), 'content/legal/refund-policy.md'),
    'utf-8',
  );

  return <LegalContent content={content} />;
}
