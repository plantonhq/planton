'use client';

import { Slide, SlideHeader, Card, CardTitle, Callout } from '@/components/deck';
import type { SlideComponentProps } from '@/components/deck';

const whereItWent = [
  { label: 'Team Salaries', detail: 'Engineering & co-founder' },
  { label: 'Cloud Infrastructure', detail: 'GCP, AWS — dogfooding Planton' },
  { label: 'Office & Hardware', detail: 'Hyderabad WeWork, dev machines' },
  { label: 'AI Tooling', detail: 'Cursor, LLM APIs, R&D subscriptions' },
];

const useOfFunds = [
  {
    icon: '📣',
    title: 'Growth & Distribution',
    pct: '50%',
    detail: 'Developer advocate, community building, event sponsorships, organized meetups',
  },
  {
    icon: '🤝',
    title: 'Enterprise Sales',
    pct: '30%',
    detail: 'B2B sales hire to convert enterprise pipeline into paying customers',
  },
  {
    icon: '⚙️',
    title: 'Product & Reliability',
    pct: '20%',
    detail: 'Continued platform hardening, agent fleet expansion, Stigmer integration',
  },
];

const milestones = [
  '50 enterprise clients',
  '$100K MRR',
  'Series A ready',
  'Active developer community around Planton open source & Stigmer',
];

export default function S13TheAsk(_props: SlideComponentProps) {
  return (
    <Slide>
      <SlideHeader
        sectionTag="The Ask"
        title="The Product Is Built. Now We Need to Grow."
        subtitle="SAFE Note — $7M Valuation Cap"
      />

      {/* Where $500K went + what we're raising side by side */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 sm:gap-6 mb-4 sm:mb-6 max-w-4xl mx-auto">
        {/* Where it went */}
        <Card className="text-left !border-white/10">
          <CardTitle className="mb-3 !text-sm sm:!text-base">Where $500K Went</CardTitle>
          <div className="space-y-2">
            {whereItWent.map((item) => (
              <div key={item.label} className="flex items-center justify-between text-xs sm:text-sm">
                <span className="text-white/70">{item.label}</span>
                <span className="text-white/40">{item.detail}</span>
              </div>
            ))}
          </div>
          <div className="mt-3 pt-3 border-t border-white/10 flex items-center justify-between text-xs sm:text-sm">
            <span className="text-white/50">Current monthly burn</span>
            <span className="text-white font-semibold">~$8K/mo</span>
          </div>
        </Card>

        {/* What we're raising */}
        <Card className="text-left !border-[#10b981]/30 !bg-[#10b981]/5">
          <div className="text-center mb-3">
            <div className="text-2xl sm:text-3xl md:text-4xl font-bold text-white">$500K</div>
            <div className="text-xs sm:text-sm text-white/60">SAFE &middot; $7M Cap</div>
          </div>
          <div className="space-y-2">
            {useOfFunds.map((item) => (
              <div key={item.title} className="flex items-start gap-2">
                <span className="text-base shrink-0">{item.icon}</span>
                <div className="flex-1">
                  <div className="flex items-center justify-between">
                    <span className="text-xs sm:text-sm text-white font-medium">{item.title}</span>
                    <span className="text-xs sm:text-sm text-[#10b981] font-bold">{item.pct}</span>
                  </div>
                  <p className="text-xs text-white/50">{item.detail}</p>
                </div>
              </div>
            ))}
          </div>
        </Card>
      </div>

      <Callout className="max-w-3xl mx-auto mb-4 sm:mb-6">
        <p className="text-xs sm:text-sm font-semibold text-white text-center mb-2">
          18-Month Milestones
        </p>
        <div className="grid grid-cols-2 gap-x-4 gap-y-1">
          {milestones.map((m) => (
            <div key={m} className="flex items-start gap-2 text-xs sm:text-sm text-white/60">
              <span className="text-[#10b981] shrink-0">→</span>
              <span>{m}</span>
            </div>
          ))}
        </div>
      </Callout>

      <div className="text-center">
        <p className="text-xs sm:text-sm text-white/40">
          Stigmer: separate Delaware C-Corp, same $7M cap, separate investment opportunity.
        </p>
      </div>
    </Slide>
  );
}
