import { Metadata } from 'next';
import Link from 'next/link';
import { notFound } from 'next/navigation';
import { getAllInvestorUpdates, getInvestorUpdate, formatUpdateDate } from '@/lib/investor-updates';

type InvestorUpdateParams = Promise<{ slug: string }>;

interface InvestorUpdatePageProps {
  params: InvestorUpdateParams;
}

export async function generateStaticParams() {
  const updates = await getAllInvestorUpdates();
  return updates.map((update) => ({
    slug: update.slug,
  }));
}

export async function generateMetadata({ params }: InvestorUpdatePageProps): Promise<Metadata> {
  const { slug } = await params;
  const update = await getInvestorUpdate(slug);

  if (!update) {
    return {
      title: 'Update Not Found | Planton',
    };
  }

  return {
    title: `${update.frontmatter.title} | Investor Updates | Planton`,
    description: update.excerpt,
    openGraph: {
      title: update.frontmatter.title,
      description: update.excerpt,
      type: 'article',
      publishedTime: update.frontmatter.date,
    },
  };
}

/**
 * Simple Markdown to HTML renderer
 * For investor updates, we keep it simple without heavy dependencies
 */
function renderMarkdown(content: string): string {
  return content
    // Headers
    .replace(/^### (.*$)/gim, '<h3 class="text-lg font-semibold text-white mt-6 mb-2">$1</h3>')
    .replace(/^## (.*$)/gim, '<h2 class="text-xl font-bold text-white mt-8 mb-3">$1</h2>')
    .replace(/^# (.*$)/gim, '<h1 class="text-2xl font-bold text-white mt-8 mb-4">$1</h1>')
    // Bold
    .replace(/\*\*(.+?)\*\*/g, '<strong class="text-white font-semibold">$1</strong>')
    // Italic
    .replace(/\*(.+?)\*/g, '<em>$1</em>')
    // Links
    .replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" class="text-white hover:underline" target="_blank" rel="noopener noreferrer">$1</a>')
    // Unordered lists
    .replace(/^\s*[-*]\s+(.*)$/gim, '<li class="ml-4 text-white/60">$1</li>')
    // Paragraphs (double newlines)
    .replace(/\n\n/g, '</p><p class="text-white/60 mb-4">')
    // Single newlines within paragraphs
    .replace(/\n/g, '<br />');
}

export default async function InvestorUpdatePage({ params }: InvestorUpdatePageProps) {
  const { slug } = await params;
  const update = await getInvestorUpdate(slug);

  if (!update) {
    notFound();
  }

  const htmlContent = renderMarkdown(update.content);

  return (
    <div className="min-h-screen bg-[#0a0a0a]">
      {/* Content - pt-20 accounts for parent HeaderLogo */}
      <article className="container mx-auto px-4 pt-20 pb-16 max-w-3xl">
        {/* Header */}
        <header className="mb-8 pb-8 border-b border-white/10">
          {/* Date badge */}
          <div className="flex items-center gap-3 mb-4">
            <span className="px-2 py-1 bg-[#2a2a2a] border border-[#3a3a3a] rounded text-xs font-mono text-[#a0a0a0]">
              {update.frontmatter.date}
            </span>
            <time
              dateTime={update.frontmatter.date}
              className="text-sm text-white/40"
            >
              {formatUpdateDate(update.frontmatter.date)}
            </time>
          </div>

          {/* Title */}
          <h1 className="text-3xl md:text-4xl font-bold text-white mb-4">
            {update.frontmatter.title}
          </h1>

          {/* Tags */}
          {update.frontmatter.tags && update.frontmatter.tags.length > 0 && (
            <div className="flex flex-wrap gap-2">
              {update.frontmatter.tags.map((tag) => (
                <span
                  key={tag}
                  className="px-3 py-1 bg-[#2a2a2a] text-[#a0a0a0] border border-[#3a3a3a] rounded-full text-xs font-medium"
                >
                  {tag}
                </span>
              ))}
            </div>
          )}
        </header>

        {/* Markdown Content */}
        <div
          className="prose prose-invert prose-lg max-w-none"
          dangerouslySetInnerHTML={{ __html: `<p class="text-white/60 mb-4">${htmlContent}</p>` }}
        />

        {/* Navigation */}
        <nav className="mt-12 pt-8 border-t border-white/10">
          <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
            <Link
              href="/legal/investor-updates"
              className="inline-flex items-center gap-2 px-4 py-2 rounded-lg border border-white/10 text-white/60 hover:bg-white/5 transition-colors"
            >
              ← Back to All Updates
            </Link>
            <div className="flex gap-3">
              <Link
                href="/invest/opportunity"
                className="text-sm text-white/40 hover:text-white/60 transition-colors"
              >
                Market Opportunity
              </Link>
              <Link
                href="/invest/process"
                className="text-sm text-white/40 hover:text-white/60 transition-colors"
              >
                Investment Process
              </Link>
            </div>
          </div>
        </nav>
      </article>

      {/* Footer */}
      <footer className="border-t border-white/5 py-8">
        <div className="container mx-auto px-4 max-w-3xl text-center">
          <p className="text-sm text-white/30">
            Questions about this update?{' '}
            <a
              href="mailto:swarup@planton.ai"
              className="text-white hover:underline"
            >
              swarup@planton.ai
            </a>
          </p>
        </div>
      </footer>
    </div>
  );
}
