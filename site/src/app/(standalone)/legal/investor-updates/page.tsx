import { Metadata } from 'next';
import Link from 'next/link';
import { getAllInvestorUpdates } from '@/lib/investor-updates';
import InvestorUpdatesTimeline from '@/components/investor-updates/InvestorUpdatesTimeline';

export const metadata: Metadata = {
  title: 'Investor Updates | Planton',
  description:
    'Transparent updates on Planton progress, metrics, wins, challenges, and use of funds. Regular communication for investors and supporters.',
  openGraph: {
    title: 'Investor Updates | Planton',
    description:
      'Transparent updates on Planton progress, metrics, wins, challenges, and use of funds.',
    type: 'website',
  },
};

export default async function InvestorUpdatesPage() {
  const updates = await getAllInvestorUpdates();

  return (
    <div className="min-h-screen bg-[#0a0a0a]">
      {/* Hero Section - pt-20 accounts for parent HeaderLogo */}
      <div className="pt-20 pb-8 border-b border-white/5">
        <div className="container mx-auto px-4 py-12 max-w-3xl">
          <h1 className="text-4xl md:text-5xl font-bold text-white mb-4">
            Investor Updates
          </h1>
          <p className="text-lg text-white/50 mb-6">
            Transparency in progress. Regular updates on what we&apos;re building,
            how we&apos;re spending, and where we&apos;re headed.
          </p>
          <div className="flex flex-wrap gap-3">
            <span className="px-3 py-1 bg-[#2a2a2a] border border-[#3a3a3a] rounded-full text-sm text-[#a0a0a0]">
              Monthly Updates
            </span>
            <span className="px-3 py-1 bg-[#2a2a2a] border border-[#3a3a3a] rounded-full text-sm text-[#a0a0a0]">
              Metrics & KPIs
            </span>
            <span className="px-3 py-1 bg-[#2a2a2a] border border-[#3a3a3a] rounded-full text-sm text-[#a0a0a0]">
              Wins & Challenges
            </span>
          </div>
        </div>
      </div>

      {/* Timeline */}
      <main className="container mx-auto px-4 py-12 max-w-3xl">
        {updates.length === 0 ? (
          <div className="bg-white/5 border border-white/10 rounded-xl p-12 text-center">
            <div className="text-4xl mb-4">📝</div>
            <h2 className="text-xl font-semibold text-white mb-2">
              Updates Coming Soon
            </h2>
            <p className="text-white/50 mb-6">
              We&apos;re preparing our first investor update. Check back soon for
              regular updates on progress, metrics, and milestones.
            </p>
            <Link
              href="/invest/process"
              className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-[#2a2a2a] border border-[#3a3a3a] text-white hover:bg-[#3a3a3a] transition-colors"
            >
              Learn About Investing →
            </Link>
          </div>
        ) : (
          <InvestorUpdatesTimeline updates={updates} />
        )}
      </main>

      {/* Footer */}
      <footer className="border-t border-white/5 py-8">
        <div className="container mx-auto px-4 max-w-3xl">
          <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
            <p className="text-sm text-white/30">
              Questions?{' '}
              <a
                href="mailto:swarup@planton.ai"
                className="text-white hover:underline"
              >
                swarup@planton.ai
              </a>
            </p>
            <div className="flex gap-4">
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
        </div>
      </footer>
    </div>
  );
}
