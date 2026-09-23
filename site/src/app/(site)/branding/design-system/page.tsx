import fs from 'fs';
import path from 'path';
import { Metadata } from 'next';
import matter from 'gray-matter';
import { MDXRenderer } from '@/lib/MDXRenderer';
import { MDXParser } from '@/lib/mdx';
import { BrandingContentLayout } from '@/components/branding/BrandingContentLayout';
import { BRANDING_DIRECTORY } from '@/lib/constants';

const BRANDING_FILE = path.join(
  /* turbopackIgnore: true */ BRANDING_DIRECTORY,
  'design-system.md',
);

export async function generateMetadata(): Promise<Metadata> {
  const content = fs.readFileSync(BRANDING_FILE, 'utf-8');
  const { data } = matter(content);

  return {
    title: `${data.title || 'Design System'} - Planton`,
    description:
      data.description ||
      'Canonical visual language reference for the Planton platform.',
  };
}

export default function BrandingDesignSystemPage() {
  const content = fs.readFileSync(BRANDING_FILE, 'utf-8');
  const { data } = matter(content);
  const mdxContent = MDXParser.reconstructMDX(content);

  return (
    <BrandingContentLayout content={content}>
      <MDXRenderer
        mdxContent={mdxContent}
        markdownContent={content}
        title={data?.title}
        path="/branding/design-system"
        rawPath="/branding/design-system.md"
      />
    </BrandingContentLayout>
  );
}
